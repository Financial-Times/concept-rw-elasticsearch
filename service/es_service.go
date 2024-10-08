package service

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"io"

	log "github.com/Financial-Times/go-logger"
	tid "github.com/Financial-Times/transactionid-utils-go"
	"github.com/olivere/elastic/v7"
)

var (
	ErrNoElasticClient = errors.New("no ElasticSearch client available")
)

const (
	conceptTypeField   = "conceptType"
	uuidField          = "uuid"
	concordedUUIDField = "concordedUUID"
	prefUUIDField      = "prefUUID"
	statusField        = "status"
	operationField     = "operation"
	writeOperation     = "write"
	deleteOperation    = "delete"
	unknownStatus      = "unknown"
	tidNotFound        = "not found"
	ftOrgUUID          = "7bcfe07b-0fb1-49ce-a5fa-e51d5c01c3e0"
	columnistUUID      = "7ef75a6a-b6bf-4eb7-a1da-03e0acabef1b"
	journalistUUID     = "33ee38a4-c677-4952-a141-2ae14da3aedd"
	notFoundResult     = "not_found"
	allConceptsAlias   = "all-concepts"
)

type esService struct {
	sync.RWMutex
	elasticClient       *elastic.Client
	bulkProcessor       *elastic.BulkProcessor
	indexName           string
	bulkProcessorConfig *BulkProcessorConfig
	getCurrentTime      func() time.Time
}

type EsService interface {
	LoadData(ctx context.Context, conceptType string, uuid string, payload EsModel) (bool, *elastic.IndexResponse, error)
	ReadData(uuid string) (*elastic.GetResult, error)
	DeleteData(ctx context.Context, conceptType string, uuid string) (*elastic.DeleteResponse, error)
	LoadBulkData(uuid string, payload interface{})
	CleanupData(ctx context.Context, concept Concept)
	PatchUpdateConcept(uuid string, payload PayloadPatch)
	CloseBulkProcessor() error
	GetClusterHealth() (*elastic.ClusterHealthResponse, error)
	IsIndexReadOnly() (bool, string, error)
	GetAllIDs(ctx context.Context, includeTypes bool, excludeFTPinkAuthorities bool) chan EsIDTypePair
}

func NewEsService(ch chan *elastic.Client, indexName string, bulkProcessorConfig *BulkProcessorConfig) EsService {
	es := &esService{bulkProcessorConfig: bulkProcessorConfig, indexName: indexName, getCurrentTime: time.Now}
	go func() {
		for ec := range ch {
			es.setElasticClient(ec)
		}
	}()
	return es
}

func (es *esService) setElasticClient(ec *elastic.Client) {
	es.Lock()
	defer es.Unlock()

	es.elasticClient = ec

	if es.bulkProcessor != nil {
		err := es.CloseBulkProcessor()
		if err != nil {
			log.Errorf("Error closing bulk processor: %v", err)
		}
	}

	if es.bulkProcessorConfig != nil {
		bulkProcessor, err := newBulkProcessor(ec, es.bulkProcessorConfig)
		if err != nil {
			log.Errorf("Creating bulk processor failed with error=[%v]", err)
		}
		es.bulkProcessor = bulkProcessor
	}
}

func (es *esService) GetClusterHealth() (*elastic.ClusterHealthResponse, error) {
	es.RLock()
	defer es.RUnlock()

	if err := es.checkElasticClient(); err != nil {
		return nil, err
	}

	return es.elasticClient.ClusterHealth().Do(context.Background())
}

func (es *esService) IsIndexReadOnly() (bool, string, error) {
	es.RLock()
	defer es.RUnlock()

	if err := es.checkElasticClient(); err != nil {
		return false, "", err
	}

	resp, err := es.elasticClient.IndexGetSettings(es.indexName).Do(context.Background())
	if err != nil {
		return false, "", err
	}

	for k, v := range resp {
		readOnly, err := es.isIndexReadOnly(v.Settings)
		return readOnly, k, err
	}

	return false, "", errors.New("no index settings found")
}

func (es *esService) isIndexReadOnly(settings map[string]interface{}) (bool, error) {
	indexSettings := settings["index"].(map[string]interface{})
	if block, hasBlockSetting := indexSettings["blocks"]; hasBlockSetting {
		if writeBlocked, hasWriteBlockSetting := block.(map[string]interface{})["write"]; hasWriteBlockSetting {
			readOnly, err := strconv.ParseBool(writeBlocked.(string))
			return readOnly, err
		}
	}

	return false, nil
}

func isFtAuthor(memberships []string) bool {
	for _, m := range memberships {
		if m == journalistUUID || m == columnistUUID {
			return true
		}
	}
	return false
}

func (es *esService) LoadData(ctx context.Context, conceptType string, uuid string, payload EsModel) (
	updated bool, resp *elastic.IndexResponse, err error) {

	loadDataLog := log.WithField(conceptTypeField, conceptType).
		WithField(uuidField, uuid).
		WithField(operationField, writeOperation)

	transactionID, err := tid.GetTransactionIDFromContext(ctx)
	if err != nil {
		transactionID = tidNotFound
		err = nil
	}
	loadDataLog = loadDataLog.WithField(tid.TransactionIDKey, transactionID)

	es.RLock()
	defer es.RUnlock()

	if err = es.checkElasticClient(); err != nil {
		loadDataLog.WithError(err).WithField(statusField, unknownStatus).Error("Failed operation to Elasticsearch")
		return updated, resp, err
	}

	// Check if membership is FT
	if conceptType == memberships {
		emm := payload.(*EsMembershipModel)
		if emm.OrganisationId != ftOrgUUID || len(emm.Memberships) < 1 || !isFtAuthor(emm.Memberships) { // drop as not FT Author
			return updated, resp, err
		}
		uuid = emm.PersonId // membership is for person
	}
	readResult, err := es.ReadData(uuid)

	patchData := getPatchData(err, loadDataLog, conceptType, readResult)

	if readResult != nil && !readResult.Found && conceptType == memberships {
		//we write a dummy person
		p := EsPersonConceptModel{
			EsConceptModel: &EsConceptModel{
				Id:           uuid,
				Type:         person,
				LastModified: es.getCurrentTime().Format(time.RFC3339),
			},
			IsFTAuthor: "true",
		}
		logDebugPersonData(loadDataLog, &p, "Writing a dummy person")
		return es.writeToEs(ctx, loadDataLog, uuid, p)
	}

	if conceptType != memberships {
		updated, resp, err = es.writeToEs(ctx, loadDataLog, uuid, payload)
	}

	//check if patchData is empty
	if patchData != nil {
		if conceptType == memberships {
			// `patchData` is for a person
			logDebugPatchData(loadDataLog, patchData, "patch for person ")
		} else {
			logDebugPatchData(loadDataLog, patchData, "patch for concept ")
		}
		es.PatchUpdateConcept(uuid, patchData)
		updated = true
	}
	return updated, resp, err
}

func (es *esService) writeToEs(ctx context.Context, loadDataLog *logrus.Entry, uuid string, payload EsModel) (updated bool, resp *elastic.IndexResponse, err error) {
	loadDataLog.Debugf("Writing: %s", uuid)
	resp, err = es.elasticClient.Index().
		Index(es.indexName).
		Id(uuid).
		BodyJson(payload).
		Do(ctx)

	if err != nil {
		status := unknownStatus
		var esErr *elastic.Error
		if errors.As(err, &esErr) {
			status = strconv.Itoa(esErr.Status)
		}
		loadDataLog.WithError(err).WithField(statusField, status).Error("Failed operation to Elasticsearch")
		return false, resp, err
	}
	return true, resp, nil
}

func getPatchData(err error, loadDataLog *logrus.Entry, conceptType string, readResult *elastic.GetResult) (patchData PayloadPatch) {
	if err != nil {
		loadDataLog.WithError(err).Error("Failed operation to Elasticsearch, could not retrieve current values before write")
		return patchData
	} else {
		//we need to write the annotation count separately as it is sourced from neo.
		//there is a race condition between the dataload and the patchData patch this will be solved by querying for the latest patchData
		//from neo before writing the patchData back
		switch conceptType {
		case person, memberships:
			esConcept := new(EsPersonConceptModel)
			if readResult.Found {
				if err := json.Unmarshal(readResult.Source, esConcept); err != nil {
					loadDataLog.WithError(err).Error("Failed to read patchData from Elasticsearch")
					return patchData
				} else {
					if conceptType == memberships {
						return &EsPersonConceptPatch{Metrics: esConcept.Metrics, IsFTAuthor: "true"} // we only process FT members who are FT authors
					}
					return &EsPersonConceptPatch{Metrics: esConcept.Metrics, IsFTAuthor: esConcept.IsFTAuthor}
				}
			}
		default:
			esConcept := new(EsConceptModel)
			if readResult.Found {
				if err := json.Unmarshal(readResult.Source, esConcept); err != nil {
					loadDataLog.WithError(err).Error("Failed to read patchData from Elasticsearch")
					return patchData
				}
				return &EsConceptModelPatch{Metrics: esConcept.Metrics}
			}
		}
	}
	return patchData
}

func (es *esService) checkElasticClient() error {
	if es.elasticClient == nil {
		return ErrNoElasticClient
	}

	return nil
}

func (es *esService) ReadData(uuid string) (*elastic.GetResult, error) {
	es.RLock()
	defer es.RUnlock()

	if err := es.checkElasticClient(); err != nil {
		return nil, err
	}

	resp, err := es.elasticClient.Get().
		Index(es.indexName).
		Id(uuid).
		Do(context.Background())

	if elastic.IsNotFound(err) {
		return &elastic.GetResult{Found: false}, nil
	} else {
		return resp, err
	}
}

func (es *esService) CleanupData(ctx context.Context, concept Concept) {
	cleanupDataLog := log.WithField(prefUUIDField, concept.PreferredUUID())
	transactionID, err := tid.GetTransactionIDFromContext(ctx)
	if err != nil {
		transactionID = tidNotFound
	}
	cleanupDataLog = cleanupDataLog.WithTransactionID(transactionID)

	conceptTypeMap, err := es.findConceptTypes(ctx, concept.ConcordedUUIDs())
	if err != nil {
		cleanupDataLog.WithError(err).Error("Impossible to find concorded concepts in elasticsearch")
		return
	}

	for concordedUUID, conceptType := range conceptTypeMap {
		cleanupDataLog.WithField(concordedUUIDField, concordedUUID).
			WithField(conceptTypeField, conceptType).
			Info("Cleaning up concorded uuids")
		_, err := es.DeleteData(ctx, conceptType, concordedUUID)
		if err != nil {
			cleanupDataLog.WithError(err).WithField(concordedUUIDField, concordedUUID).
				WithField(conceptTypeField, conceptType).
				Error("Failed to delete concorded uuid.")
		}
	}
}

func (es *esService) findConceptTypes(ctx context.Context, uuids []string) (map[string]string, error) {
	if err := es.checkElasticClient(); err != nil {
		return nil, err
	}

	query := elastic.NewIdsQuery().Ids(uuids...)
	result, err := es.elasticClient.Search(es.indexName).Query(query).Do(ctx)
	if err != nil {
		return nil, err
	}

	conceptTypeMap := make(map[string]string)
	for _, hit := range result.Hits.Hits {
		esModel := EsConceptModel{}
		err = json.Unmarshal(hit.Source, &esModel)
		if err != nil {
			return nil, err
		}
		conceptTypeMap[hit.Id] = esModel.Type
	}

	return conceptTypeMap, nil
}

func (es *esService) DeleteData(ctx context.Context, conceptType string, uuid string) (*elastic.DeleteResponse, error) {
	deleteDataLog := log.WithField(conceptTypeField, conceptType).
		WithField(uuidField, uuid).
		WithField(operationField, deleteOperation)

	transactionID, err := tid.GetTransactionIDFromContext(ctx)
	if err != nil {
		transactionID = tidNotFound
	}
	deleteDataLog = deleteDataLog.WithField(tid.TransactionIDKey, transactionID)

	if err := es.checkElasticClient(); err != nil {
		deleteDataLog.WithError(err).
			WithField(statusField, unknownStatus).
			Error("Failed operation to Elasticsearch")
		return nil, err
	}

	resp, err := es.elasticClient.Delete().
		Index(es.indexName).
		Id(uuid).
		Do(ctx)

	if elastic.IsNotFound(err) {
		return &elastic.DeleteResponse{Result: notFoundResult}, nil
	}

	if err != nil {
		var status string
		switch err.(type) {
		case *elastic.Error:
			status = strconv.Itoa(err.(*elastic.Error).Status)
		default:
			status = unknownStatus
		}
		deleteDataLog.WithError(err).
			WithField(statusField, status).
			Error("Failed operation to Elasticsearch")
	}

	return resp, err
}

func (es *esService) LoadBulkData(uuid string, payload interface{}) {
	r := elastic.NewBulkIndexRequest().Index(es.indexName).Id(uuid).Doc(payload)

	es.RLock()
	defer es.RUnlock()

	es.bulkProcessor.Add(r)
}

// PatchUpdateConcept updates a concept document with metrics. See https://www.elastic.co/guide/en/elasticsearch/reference/current/docs-update.html#_updates_with_a_partial_document
func (es *esService) PatchUpdateConcept(uuid string, payload PayloadPatch) {
	r := elastic.NewBulkUpdateRequest().Index(es.indexName).Id(uuid).Doc(payload)

	es.RLock()
	defer es.RUnlock()

	es.bulkProcessor.Add(r)
}

func (es *esService) CloseBulkProcessor() error {
	return es.bulkProcessor.Close()
}

func (es *esService) GetAllIDs(ctx context.Context, includeTypes bool, excludeFTPinkAuthorities bool) chan EsIDTypePair {
	ids := make(chan EsIDTypePair)

	go func() {
		defer close(ids)
		var r *elastic.ScrollService
		if excludeFTPinkAuthorities {
			r = elastic.NewScrollService(es.elasticClient).
				Index(allConceptsAlias).
				Query(elastic.NewBoolQuery().
					MustNot(elastic.NewTermsQuery("authorities", "TME", "Smartlogic"))).
				Sort("_doc", true).
				Size(1000).
				FetchSource(includeTypes)
		} else {
			r = elastic.NewScrollService(es.elasticClient).
				Index(es.indexName).
				Query(elastic.NewMatchAllQuery()).
				Sort("_doc", true).
				Size(1000).
				FetchSource(includeTypes)
		}

		es.RLock()
		defer es.RUnlock()

		var err error
		for {
			r, err = es.processScrollPage(ctx, r, ids, includeTypes)
			if r == nil || err != nil {
				return
			}
		}
	}()

	return ids
}

func (es *esService) processScrollPage(ctx context.Context, r *elastic.ScrollService, ch chan EsIDTypePair, includeTypes bool) (*elastic.ScrollService, error) {
	res, err := r.Do(ctx)
	if err == io.EOF {
		return nil, nil
	} else if err != nil {
		log.Error("error while fetching ids", err)
		return nil, err
	}

	scrollId := res.ScrollId
	for _, c := range res.Hits.Hits {
		if includeTypes {
			esModel := EsConceptModel{}
			err = json.Unmarshal(c.Source, &esModel)
			if err != nil {
				return nil, err
			}
			ch <- EsIDTypePair{ID: c.Id, Type: esModel.Type}
		} else {
			ch <- EsIDTypePair{ID: c.Id}
		}
	}

	return elastic.NewScrollService(es.elasticClient).ScrollId(scrollId), nil
}

func logDebugPatchData(log *logrus.Entry, payload PayloadPatch, msg string) {

	var data []byte
	var err error
	switch payload := payload.(type) {
	case *EsConceptModelPatch, *EsPersonConceptPatch:
		data, err = json.Marshal(payload)
	default:
		log.Error("called logDebugPersonPatchData with unexpected payload type")
		return
	}
	if err != nil {
		log.Errorf("could not log payload: %v", err)
		return
	}
	log.Debugf("%s: %s", msg, string(data))
}

func logDebugPersonData(log *logrus.Entry, concept *EsPersonConceptModel, msg string) {
	data, err := json.Marshal(concept)
	if err != nil {
		log.Errorf("could not log concept: %v", err)
		return
	}
	log.Debugf("%s: %s", msg, string(data))
}
