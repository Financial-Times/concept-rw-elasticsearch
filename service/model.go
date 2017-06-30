package service

type ConceptModel struct {
	UUID       string   `json:"uuid"`
	DirectType string   `json:"type"`
	PrefLabel  string   `json:"prefLabel"`
	Aliases    []string `json:"aliases,omitempty"`
}

type AggregateConceptModel struct {
	PrefUUID              string          `json:"prefUUID"`
	DirectType            string          `json:"type"`
	PrefLabel             string          `json:"prefLabel"`
	Aliases               []string        `json:"aliases,omitempty"`
	SourceRepresentations []SourceConcept `json:"sourceRepresentations"`
}

type SourceConcept struct {
	UUID      string `json:"uuid"`
	Authority string `json:"authority"`
}

type EsConceptModel struct {
	Id         string   `json:"id"`
	ApiUrl     string   `json:"apiUrl"`
	PrefLabel  string   `json:"prefLabel"`
	Types      []string `json:"types"`
	DirectType string   `json:"directType"`
	Aliases    []string `json:"aliases,omitempty"`
}

type EsPersonConceptModel struct {
	EsConceptModel
	IsFTAuthor string `json:"isFTAuthor"`
}
