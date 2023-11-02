package service

// Concept contains common function between both concept models
type Concept interface {
	// GetAuthorities returns an array containing all authorities that this concept is identified by
	GetAuthorities() []string
	// ConcordedUUIDs returns an array containing all concorded concept uuids - N.B. it will not contain the canonical prefUUID.
	ConcordedUUIDs() []string

	PreferredUUID() string
}

type PayloadPatch interface{}

type ConceptModel struct {
	UUID                   string                 `json:"uuid"`
	DirectType             string                 `json:"type"`
	PrefLabel              string                 `json:"prefLabel"`
	Authority              string                 `json:"authority,omitempty"`
	Aliases                []string               `json:"aliases,omitempty"`
	AlternativeIdentifiers map[string]interface{} `json:"alternativeIdentifiers,omitempty"`
	IsDeprecated           bool                   `json:"isDeprecated,omitempty"`
	ScopeNote              string                 `json:"scopeNote,omitempty"`
}

type AggregateMembershipRole struct {
	RoleUUID        string `json:"membershipRoleUUID,omitempty"`
	InceptionDate   string `json:"inceptionDate,omitempty"`
	TerminationDate string `json:"terminationDate,omitempty"`
}

type AggregateConceptModel struct {
	// Required fields
	PrefUUID   string `json:"prefUUID"`
	DirectType string `json:"type"`
	PrefLabel  string `json:"prefLabel"`
	// Additional fields
	Aliases   []string `json:"aliases,omitempty"`
	ScopeNote string   `json:"scopeNote,omitempty"`
	// Membership
	MembershipRoles  []AggregateMembershipRole `json:"membershipRoles,omitempty"`
	OrganisationUUID string                    `json:"organisationUUID,omitempty"`
	PersonUUID       string                    `json:"personUUID,omitempty"`
	// Organisation
	CountryCode            string `json:"countryCode,omitempty"`
	CountryOfIncorporation string `json:"countryOfIncorporation,omitempty"`
	IsDeprecated           bool   `json:"isDeprecated,omitempty"`
	// Source representations
	SourceRepresentations []SourceConcept `json:"sourceRepresentations"`
	// NAICS
	NAICS []NAICS `json:"naicsIndustryClassifications"`
}

type SourceConcept struct {
	UUID      string `json:"uuid"`
	Authority string `json:"authority"`
}

type NAICS struct {
	UUID string `json:"uuid"`
	Rank int    `json:"rank"`
}
type EsModel interface{}

type EsConceptModel struct {
	Id                     string          `json:"id"`
	Type                   string          `json:"type,omitempty"`
	ApiUrl                 string          `json:"apiUrl"`
	PrefLabel              string          `json:"prefLabel"`
	Types                  []string        `json:"types"`
	Authorities            []string        `json:"authorities"`
	DirectType             string          `json:"directType"`
	Aliases                []string        `json:"aliases,omitempty"`
	LastModified           string          `json:"lastModified"`
	PublishReference       string          `json:"publishReference"`
	IsDeprecated           bool            `json:"isDeprecated,omitempty"` // stored only if this is true
	ScopeNote              string          `json:"scopeNote,omitempty"`
	CountryCode            string          `json:"countryCode,omitempty"`
	CountryOfIncorporation string          `json:"countryOfIncorporation,omitempty"`
	Metrics                *ConceptMetrics `json:"metrics,omitempty"`
	NAICS                  []NAICS         `json:"NAICS"`
}

type EsMembershipModel struct {
	Id             string   `json:"id"`
	PersonId       string   `json:"personId"`
	OrganisationId string   `json:"organisationId"`
	Memberships    []string `json:"memberships"`
}

type EsIDTypePair struct {
	ID   string `json:"id,omitempty"`
	Type string `json:"type,omitempty"`
}

type EsConceptModelPatch struct {
	Metrics *ConceptMetrics `json:"metrics"`
}

type ConceptMetrics struct {
	AnnotationsCount         int `json:"annotationsCount"`
	PrevWeekAnnotationsCount int `json:"prevWeekAnnotationsCount"`
}

type EsPersonConceptModel struct {
	*EsConceptModel
	IsFTAuthor string `json:"isFTAuthor,omitempty"`
}

type EsPersonConceptPatch struct {
	Metrics    *ConceptMetrics `json:"metrics"`
	IsFTAuthor string          `json:"isFTAuthor"`
}

func (c AggregateConceptModel) PreferredUUID() string {
	return c.PrefUUID
}

func (c ConceptModel) PreferredUUID() string {
	return c.UUID
}

func (c ConceptModel) GetAuthorities() []string {
	var authorities []string

	if c.AlternativeIdentifiers == nil && c.Authority != "" {
		return []string{c.Authority}
	}

	for authority := range c.AlternativeIdentifiers {
		if authority == "uuids" {
			continue // exclude the "uuids" alternativeIdentifier
		}
		authorities = append(authorities, authority)
	}
	return authorities
}

func (c AggregateConceptModel) GetAuthorities() []string {
	var authorities []string
	for _, src := range c.SourceRepresentations {
		authorities = append(authorities, src.Authority)
	}
	return authorities
}

func (c ConceptModel) ConcordedUUIDs() []string {
	return make([]string, 0) // we don't want to remove concorded concepts for the original concept model.
}

func (c AggregateConceptModel) ConcordedUUIDs() []string {
	var uuids []string
	for _, src := range c.SourceRepresentations {
		if src.UUID != c.PrefUUID {
			uuids = append(uuids, src.UUID)
		}
	}
	return uuids
}
