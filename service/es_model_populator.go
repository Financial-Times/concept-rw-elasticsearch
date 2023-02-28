package service

import (
	"fmt"
	"time"

	ontology "github.com/Financial-Times/cm-graph-ontology"
	log "github.com/Financial-Times/go-logger"
)

const (
	person                  = "people"
	memberships             = "memberships"
	organisation            = "organisations"
	defaultIsFTAuthor       = "false"
	directTypePublicCompany = "PublicCompany"
	thingURL                = "http://api.ft.com/things/"
)

func ConvertConceptToESConceptModel(concept ConceptModel, conceptType, publishRef, publicAPIHost string) (EsModel, error) {
	esModel, err := newESConceptModel(
		concept.UUID,
		conceptType,
		concept.DirectType,
		concept.PrefLabel,
		publishRef,
		concept.ScopeNote,
		publicAPIHost,
		concept.Aliases,
		concept.GetAuthorities(),
		concept.IsDeprecated)

	if err != nil {
		return nil, err
	}

	switch conceptType {
	case person: // person type should not come through as the old model.
		esPersonModel := &EsPersonConceptModel{
			EsConceptModel: esModel,
		}
		return esPersonModel, nil
	default:
		return esModel, nil
	}
}

func ConvertAggregateConceptToESConceptModel(concept AggregateConceptModel, conceptType, publishRef, publicAPIHost string) (EsModel, error) {
	var esModel EsModel
	var esConceptModel *EsConceptModel
	var err error

	switch conceptType {
	case memberships:
		ms := make([]string, len(concept.MembershipRoles))
		for i, m := range concept.MembershipRoles {
			ms[i] = m.RoleUUID
		}
		esModel = &EsMembershipModel{
			Id:             concept.PrefUUID,
			PersonId:       concept.PersonUUID,
			OrganisationId: concept.OrganisationUUID,
			Memberships:    ms,
		}
	case person:
		esConceptModel, err = getEsConcept(concept, conceptType, publishRef, publicAPIHost)

		esModel = &EsPersonConceptModel{
			EsConceptModel: esConceptModel,
			IsFTAuthor:     defaultIsFTAuthor, // default as controlled by memberships concept
		}
	case organisation:
		esConceptModel, err = getEsConcept(concept, conceptType, publishRef, publicAPIHost)

		if concept.DirectType == directTypePublicCompany {
			esConceptModel.CountryCode = concept.CountryCode
			esConceptModel.CountryOfIncorporation = concept.CountryOfIncorporation
		}
		esModel = esConceptModel
	default:
		esModel, err = getEsConcept(concept, conceptType, publishRef, publicAPIHost)
	}

	return esModel, err
}

func getEsConcept(concept AggregateConceptModel, conceptType, publishRef, publicAPIHost string) (*EsConceptModel, error) {
	return newESConceptModel(
		concept.PrefUUID,
		conceptType,
		concept.DirectType,
		concept.PrefLabel,
		publishRef,
		concept.ScopeNote,
		publicAPIHost,
		concept.Aliases,
		concept.GetAuthorities(),
		concept.IsDeprecated)
}

func newESConceptModel(uuid, conceptType, directType, prefLabel, publishRef, scopeNote, publicAPIHost string, aliases, authorities []string, isDeprecated bool) (*EsConceptModel, error) {
	apiURL, err := ontology.APIURL(uuid, []string{directType}, publicAPIHost)
	if err != nil {
		return nil, err
	}

	typeURIs, err := ontology.FullTypeHierarchy(directType)
	if err != nil {
		return nil, fmt.Errorf("getting type hierarchy for %q: %w", directType, err)
	}

	directTypeURIs, err := ontology.TypeURIs([]string{directType})
	if err != nil {
		return nil, fmt.Errorf("getting type uris for %q: %w", directType, err)
	}

	var directTypeURI string
	if len(directTypeURIs) != 1 {
		log.
			WithField("conceptType", conceptType).
			WithField("prefUUID", uuid).
			WithField("typeURIs", directTypeURIs).
			Warn("Exactly one directType is expected during type mapping.")
	} else {
		directTypeURI = directTypeURIs[0]
	}

	esModel := &EsConceptModel{}
	esModel.Type = conceptType
	esModel.ApiUrl = apiURL
	esModel.Id = thingIDURL(uuid)
	esModel.DirectType = directTypeURI
	esModel.Types = typeURIs
	esModel.Aliases = aliases
	esModel.PrefLabel = prefLabel
	esModel.Authorities = authorities
	esModel.LastModified = time.Now().Format(time.RFC3339)
	esModel.PublishReference = publishRef
	esModel.IsDeprecated = isDeprecated
	esModel.ScopeNote = scopeNote

	return esModel, nil
}

func reverse(strings []string) []string {
	if strings == nil {
		return nil
	}
	if len(strings) == 0 {
		return strings
	}
	var reversed []string
	for i := len(strings) - 1; i >= 0; i = i - 1 {
		reversed = append(reversed, strings[i])
	}
	return reversed
}

func thingIDURL(uuid string) string {
	return thingURL + uuid
}
