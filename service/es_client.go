package service

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	log "github.com/Financial-Times/go-logger"
	"github.com/aws/aws-sdk-go/aws/credentials"
	awsSigner "github.com/aws/aws-sdk-go/aws/signer/v4"
	"github.com/olivere/elastic/v7"
)

type EsAccessConfig struct {
	awsCreds     *credentials.Credentials
	esEndpoint   string
	traceLogging bool
}

func NewAccessConfig(awsCreds *credentials.Credentials, endpoint string, tracelogging bool) EsAccessConfig {
	return EsAccessConfig{awsCreds: awsCreds, esEndpoint: endpoint, traceLogging: tracelogging}
}

type AWSSigningTransport struct {
	HTTPClient  *http.Client
	Credentials *credentials.Credentials
	Region      string
}

// RoundTrip implementation
func (a AWSSigningTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	signer := awsSigner.NewSigner(a.Credentials)
	body := strings.NewReader("")
	_, err := signer.Sign(req, body, "es", a.Region, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to sign request: %w", err)
	}
	return a.HTTPClient.Do(req)
}

func newAmazonClient(config EsAccessConfig, region string) (*elastic.Client, error) {
	signingTransport := AWSSigningTransport{
		Credentials: config.awsCreds,
		HTTPClient:  http.DefaultClient,
		Region:      region,
	}
	signingClient := &http.Client{Transport: http.RoundTripper(signingTransport)}

	log.Infof("connecting with AWSSigningTransport to %s", config.esEndpoint)
	return newClient(config.esEndpoint, config.traceLogging,
		elastic.SetScheme("https"),
		elastic.SetHttpClient(signingClient),
	)
}

func newSimpleClient(config EsAccessConfig) (*elastic.Client, error) {
	log.Infof("connecting with default transport to %s", config.esEndpoint)
	return newClient(config.esEndpoint, config.traceLogging)
}

func newClient(endpoint string, traceLogging bool, options ...elastic.ClientOptionFunc) (*elastic.Client, error) {
	optionFuncs := []elastic.ClientOptionFunc{
		elastic.SetURL(endpoint),
		elastic.SetSniff(false), //needs to be disabled due to EAS behavior. Healthcheck still operates as normal.
	}
	optionFuncs = append(optionFuncs, options...)

	if traceLogging {
		optionFuncs = append(optionFuncs, elastic.SetTraceLog(log.Logger()))
	}

	return elastic.NewClient(optionFuncs...)
}

func NewElasticClient(region string, config EsAccessConfig) (*elastic.Client, error) {
	if region == "local" {
		return newSimpleClient(config)
	} else {
		return newAmazonClient(config, region)
	}
}
