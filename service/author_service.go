package service

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	transactionid "github.com/Financial-Times/transactionid-utils-go"
	log "github.com/Sirupsen/logrus"
)

const contentType = "application/json"
const idsPath = "/transformers/authors/__ids"
const gtgPath = "/__gtg"

type AuthorUUID struct {
	UUID string `json:"ID"`
}

type AuthorService interface {
	LoadAuthorIdentifiers() error
	IsFTAuthor(UUID string) string
	IsGTG() error
}

//uses v1 transformer author list
type curatedAuthorService struct {
	httpClient             *http.Client
	serviceURL             string
	authorIds              []AuthorUUID
	publishClusterUser     string
	publishClusterpassword string
}

func NewAuthorService(authorIdsURL string, authorCredKey string, client *http.Client) (AuthorService, error) {
	creds := strings.Split(authorCredKey, ":")
	cas := &curatedAuthorService{client, authorIdsURL, nil, creds[0], creds[1]}
	return cas, cas.LoadAuthorIdentifiers()
}

func (as *curatedAuthorService) LoadAuthorIdentifiers() error {
	tid := transactionid.NewTransactionID()
	req, err := http.NewRequest("GET", as.serviceURL+idsPath, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", contentType)
	req.Header.Add("X-Request-Id", tid)
	req.Header.Add("User-Agent", "UPP concept-rw-elasticsearch")
	req.SetBasicAuth(as.publishClusterUser, as.publishClusterpassword)
	log.WithField("transaction_id", tid).Info("Requesting author list from v1 authors transformer." + req.RequestURI)

	resp, err := as.httpClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("A non 2xx error code from v1 authors transformer! Status: %v", resp.StatusCode)
	}

	scan := bufio.NewScanner(resp.Body)
	fmt.Print("HERE1")
	fmt.Print(scan.Text())
	for scan.Scan() {
		fmt.Print("HERE")
		var id AuthorUUID
		err = json.Unmarshal(scan.Bytes(), &id)
		if err != nil {
			return err
		}
		as.authorIds = append(as.authorIds, id)
	}

	log.Info("we have authos " + strconv.Itoa(len(as.authorIds)))

	return nil
}

func (as *curatedAuthorService) IsFTAuthor(UUID string) string {
	for _, authorId := range as.authorIds {
		if UUID == authorId.UUID {
			return "true"
		}
	}
	return "false"
}

func (as *curatedAuthorService) IsGTG() error {
	resp, err := http.Get(as.serviceURL + gtgPath)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("gtg endpoint returned a non-200 status: %v", resp.StatusCode)
	}
	return nil
}
