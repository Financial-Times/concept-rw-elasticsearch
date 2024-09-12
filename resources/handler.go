package resources

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/Financial-Times/concept-rw-elasticsearch/service"
	log "github.com/Financial-Times/go-logger"
	tid "github.com/Financial-Times/transactionid-utils-go"
	"github.com/gorilla/mux"
)

var (
	errPathUUID               = errors.New("Provided path UUID does not match request body")
	errInvalidConceptModel    = errors.New("Invalid or incomplete concept model")
	errUnsupportedConceptType = errors.New("Unsupported or invalid concept type")
	errProcessingBody         = errors.New("Request body is not in the expected concept model format")
)

const (
	notFoundResult = "not_found"
)

// Handler handles http calls
type Handler struct {
	elasticService      service.EsService
	allowedConceptTypes map[string]bool
	publicAPIHost       string
}

func NewHandler(elasticService service.EsService, allowedConceptTypes []string, publicAPIHost string) (*Handler, error) {
	if _, err := url.ParseRequestURI(publicAPIHost); err != nil {
		return nil, err
	}

	allowedTypes := make(map[string]bool)
	for _, v := range allowedConceptTypes {
		allowedTypes[v] = true
	}

	return &Handler{
		elasticService:      elasticService,
		allowedConceptTypes: allowedTypes,
		publicAPIHost:       publicAPIHost,
	}, nil
}

// LoadData processes a single ES concept entity
func (h *Handler) LoadData(w http.ResponseWriter, r *http.Request) {
	transactionID := tid.GetTransactionIDFromRequest(r)
	ctx := tid.TransactionAwareContext(r.Context(), transactionID)

	conceptType, concept, esModel, err := h.processPayload(r.WithContext(ctx))

	if err != nil {
		var errStatus int
		switch err {
		case errUnsupportedConceptType:
			errStatus = http.StatusNotFound
		default:
			errStatus = http.StatusBadRequest
		}
		writeMessage(w, err.Error(), errStatus)
		return
	}

	up, _, err := h.elasticService.LoadData(ctx, conceptType, concept.PreferredUUID(), esModel)

	if err != nil {
		if err == service.ErrNoElasticClient {
			writeMessage(w, "ES unavailable", http.StatusServiceUnavailable)
			return
		}

		log.WithError(err).Warn("Failed to write data to elasticsearch.")
		writeMessage(w, "Failed to write data to ES", http.StatusInternalServerError)
		return
	}

	if !up {
		writeMessage(w, "Concept dropped", http.StatusNotModified)
		return
	}

	h.elasticService.CleanupData(ctx, concept)
	writeMessage(w, "Concept written successfully", http.StatusOK)
}

// LoadBulkData write a concept to ES via the ES Bulk API
func (h *Handler) LoadBulkData(w http.ResponseWriter, r *http.Request) {
	transactionID := tid.GetTransactionIDFromRequest(r)
	ctx := tid.TransactionAwareContext(r.Context(), transactionID)

	_, concept, payload, err := h.processPayload(r.WithContext(ctx))
	if err == errUnsupportedConceptType {
		writeMessage(w, err.Error(), http.StatusNotFound)
		return
	}

	if err != nil {
		writeMessage(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.elasticService.LoadBulkData(concept.PreferredUUID(), payload)
	h.elasticService.CleanupData(ctx, concept)
	writeMessage(w, "Concept written successfully", http.StatusOK)
}

// LoadMetrics updates a concept with new metric data
func (h *Handler) LoadMetrics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["id"]
	conceptType := vars["concept-type"]

	if !h.allowedConceptTypes[conceptType] {
		writeMessage(w, errUnsupportedConceptType.Error(), http.StatusNotFound)
		return
	}

	dec := json.NewDecoder(r.Body)

	metrics := service.EsConceptModelPatch{}
	err := dec.Decode(&metrics)

	if err != nil {
		writeMessage(w, err.Error(), http.StatusBadRequest)
		return
	}

	if metrics.Metrics == nil {
		writeMessage(w, "Please supply metrics as a JSON object with a single property 'metrics'", http.StatusBadRequest)
		return
	}

	h.elasticService.PatchUpdateConcept(uuid, &metrics)
	writeMessage(w, "Concept updated with metrics successfully", http.StatusOK)
}

func (h *Handler) processPayload(r *http.Request) (conceptType string, concept service.Concept, esModel service.EsModel, err error) {
	vars := mux.Vars(r)
	uuid := vars["id"]
	conceptType = vars["concept-type"]

	if !h.allowedConceptTypes[conceptType] {
		return "", nil, nil, errUnsupportedConceptType
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.WithError(err).Error("Failed to read request body")
		return "", nil, nil, errProcessingBody
	}

	aggConceptModel, err := isAggregateConceptModel(body)
	if err != nil {
		log.WithError(err).Error("Failed to check if body json is an aggregate concept model or not")
		return "", nil, nil, errProcessingBody
	}

	if aggConceptModel {
		concept, esModel, err = processAggregateConceptModel(r.Context(), uuid, conceptType, h.publicAPIHost, body)
	} else {
		concept, esModel, err = processConceptModel(r.Context(), uuid, conceptType, h.publicAPIHost, body)
	}

	return conceptType, concept, esModel, err
}

func processConceptModel(ctx context.Context, uuid, conceptType, publicAPIHost string, body []byte) (concept service.ConceptModel, payload service.EsModel, err error) {
	err = json.Unmarshal(body, &concept)
	if err != nil {
		log.WithError(err).Info("Failed to unmarshal body into concept model.")
		return concept, payload, errProcessingBody
	}

	if concept.UUID != uuid {
		return concept, payload, errPathUUID
	}

	if concept.DirectType == "" || concept.PrefLabel == "" {
		return concept, payload, errInvalidConceptModel
	}

	transactionID, err := tid.GetTransactionIDFromContext(ctx)

	if err != nil {
		transactionID = tid.NewTransactionID()
		log.WithError(err).WithField(tid.TransactionIDKey, transactionID).Warn("Transaction ID not found to process concept model. Generated new transaction ID")
		err = nil // blank error just in case
	}

	payload, err = service.ConvertConceptToESConceptModel(concept, conceptType, transactionID, publicAPIHost)
	return concept, payload, err
}

func processAggregateConceptModel(ctx context.Context, uuid, conceptType, publicAPIHost string, body []byte) (concept service.AggregateConceptModel, esModel service.EsModel, err error) {
	err = json.Unmarshal(body, &concept)
	if err != nil {
		log.WithError(err).Info("Failed to unmarshal body into aggregate concept model.")
		return concept, nil, errProcessingBody
	}

	if concept.PrefUUID != uuid {
		return concept, nil, errPathUUID
	}

	if concept.DirectType == "" || concept.PrefLabel == "" {
		return concept, nil, errInvalidConceptModel
	}

	transactionID, err := tid.GetTransactionIDFromContext(ctx)

	if err != nil {
		transactionID = tid.NewTransactionID()
		log.WithError(err).WithField(tid.TransactionIDKey, transactionID).Warn("Transaction ID not found to process aggregate concept model. Generated new transaction ID")
	}

	esModel, err = service.ConvertAggregateConceptToESConceptModel(concept, conceptType, transactionID, publicAPIHost)
	return concept, esModel, err
}

func (h *Handler) ReadData(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	uuid := vars["id"]
	conceptType := vars["concept-type"]

	if !h.allowedConceptTypes[conceptType] {
		writeMessage(writer, errUnsupportedConceptType.Error(), http.StatusBadRequest)
		return
	}

	getResult, err := h.elasticService.ReadData(uuid)

	if err != nil {
		log.Error(err.Error())

		if err == service.ErrNoElasticClient {
			writer.WriteHeader(http.StatusServiceUnavailable)
		} else {
			writer.WriteHeader(http.StatusInternalServerError)
		}

		return
	}

	if !getResult.Found {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	// remove es type field from the result source, because it is not available in the current read api
	esModel := service.EsPersonConceptModel{}
	err = json.Unmarshal(getResult.Source, &esModel)
	if err != nil {
		log.Error(err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	esModel.Type = ""

	writer.Header().Add("Content-Type", "application/json")
	enc := json.NewEncoder(writer)
	err = enc.Encode(esModel)
	if err != nil {
		log.Error(err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// DeleteData handles a delete for a concept
func (h *Handler) DeleteData(writer http.ResponseWriter, request *http.Request) {
	transactionID := tid.GetTransactionIDFromRequest(request)
	ctx := tid.TransactionAwareContext(context.Background(), transactionID)

	uuid := mux.Vars(request)["id"]
	conceptType := mux.Vars(request)["concept-type"]

	if !h.allowedConceptTypes[conceptType] {
		writeMessage(writer, errUnsupportedConceptType.Error(), http.StatusBadRequest)
		return
	}

	res, err := h.elasticService.DeleteData(ctx, conceptType, uuid)

	if err != nil {
		log.Errorf(err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	if res.Result == notFoundResult {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	writer.WriteHeader(http.StatusOK)
}

func (h *Handler) GetAllIDs(writer http.ResponseWriter, request *http.Request) {
	transactionID := tid.GetTransactionIDFromRequest(request)
	ctx := tid.TransactionAwareContext(context.Background(), transactionID)

	includeTypes := strings.ToLower(request.URL.Query().Get("includeTypes")) == "true"
	excludeFTPinkAuthorities := strings.ToLower(request.URL.Query().Get("excludeFTPinkAuthorities")) == "true"

	writer.Header().Set("Content-Type", "text/plain")
	writer.WriteHeader(http.StatusOK)
	ids := h.elasticService.GetAllIDs(ctx, includeTypes, excludeFTPinkAuthorities)
	i := 0
	for id := range ids {
		if includeTypes {
			fmt.Fprintf(writer, "{\"uuid\":\"%s\",\"type\":\"%s\"}\n", id.ID, id.Type)
		} else {
			fmt.Fprintf(writer, "{\"uuid\":\"%s\"}\n", id.ID)
		}
		i++
	}
	log.Infof("wrote %v uuids", i)
}

// Close terminates the underlying ES bulk processor
func (h *Handler) Close() {
	h.elasticService.CloseBulkProcessor()
}

type responseMessage struct {
	Msg string `json:"message"`
}

func writeMessage(w http.ResponseWriter, msg string, status int) {
	w.Header().Add("Content-Type", "application/json")
	data, _ := json.Marshal(responseMessage{Msg: msg})

	w.WriteHeader(status)
	w.Write(data)
}

func isAggregateConceptModel(body []byte) (bool, error) {
	data := make(map[string]interface{})
	err := json.Unmarshal(body, &data)
	if err != nil {
		return false, err
	}

	_, ok := data["prefUUID"]
	return ok, nil
}
