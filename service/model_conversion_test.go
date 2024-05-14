package service

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	tid "github.com/Financial-Times/transactionid-utils-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testAggregateConceptModelJSON = `{"prefUUID":"56388858-38d6-4dfc-a001-506394259b51","prefLabel":"Smartlogics Brands PrefLabel","type":"Brand","strapline":"Some strapline","descriptionXML":"Some description","_imageUrl":"Some image url","sourceRepresentations":[{"uuid":"4ebbd9c4-3bb7-4d18-a14c-4c45aac5d966","prefLabel":"TMEs PrefLabel","type":"Brand","authority":"TME","authorityValue":"745212"},{"uuid":"56388858-38d6-4dfc-a001-506394259b51","prefLabel":"Smartlogics Brands PrefLabel","type":"Brand","authority":"Smartlogic","authorityValue":"123456789","lastModifiedEpoch":1498127042,"strapline":"Some strapline","descriptionXML":"Some description","_imageUrl":"Some image url"}]}`

var testConceptModelJSON = `{"uuid":"2384fa7a-d514-3d6a-a0ea-3a711f66d0d8","type":"PublicCompany","properName":"Apple, Inc.","prefLabel":"Apple, Inc.","legalName":"Apple Inc.","shortName":"Apple","alternativeIdentifiers":{"TME":["TnN0ZWluX09OX0ZvcnR1bmVDb21wYW55X0FBUEw=-T04="],"uuids":["2384fa7a-d514-3d6a-a0ea-3a711f66d0d8","2abff0bd-544d-31c3-899b-fba2f60d53dd"],"factsetIdentifier":"000C7F-E","leiCode":"HWUPKR0MPOU8FGXBT394"},"formerNames":["Apple Computer, Inc."],"aliases":["Apple Inc","Apple Computers","Apple","Apple Canada","Apple Computer","Apple Computer, Inc.","APPLE INC","Apple Incorporated","Apple Computer Inc","Apple Inc.","Apple, Inc."],"industryClassification":"7a01c847-a9bd-33be-b991-c6fbd8871a46"}`

var testIntermediateConceptModelJSON = `{"uuid":"7e3f1354-53ba-3c3e-b9bb-5fcb8941df8c","prefLabel":"ICOmedy","type":"AlphavilleSeries","authority":"TME","authorityValue":"NDQ1NjhiMzktMjJmNy00OWEzLWExNDctNDFiNDk4OGU2MTdj-QWxwaGF2aWxsZVNlcmllc0NsYXNzaWZpY2F0aW9u"}`

var publicAPIHost = "http://api.ft.com"

func TestConvertToESConceptModel(t *testing.T) {

	tests := []struct {
		conceptModel   ConceptModel
		esConceptModel EsConceptModel
		conceptType    string
	}{
		{
			conceptModel: ConceptModel{
				UUID:       "2384fa7a-d514-3d6a-a0ea-3a711f66d0d8",
				DirectType: "PublicCompany",
				PrefLabel:  "Apple, Inc.",
				AlternativeIdentifiers: map[string]interface{}{
					"Factset": "789",
					"TME":     []string{"123", "456"},
					"uuids":   []string{"uuid"},
				},
				Aliases: []string{"Apple Inc", "Apple Computers",
					"Apple",
					"Apple Canada",
					"Apple Computer",
					"Apple Computer, Inc.",
					"APPLE INC",
					"Apple Incorporated",
					"Apple Computer Inc",
					"Apple Inc.",
					"Apple, Inc."},
			},
			esConceptModel: EsConceptModel{
				Id:        "http://api.ft.com/things/2384fa7a-d514-3d6a-a0ea-3a711f66d0d8",
				Type:      "organisations",
				ApiUrl:    "http://api.ft.com/organisations/2384fa7a-d514-3d6a-a0ea-3a711f66d0d8",
				PrefLabel: "Apple, Inc.",
				Types: []string{
					"http://www.ft.com/ontology/core/Thing",
					"http://www.ft.com/ontology/concept/Concept",
					"http://www.ft.com/ontology/organisation/Organisation",
					"http://www.ft.com/ontology/company/Company",
					"http://www.ft.com/ontology/company/PublicCompany",
				},
				Authorities: []string{"Factset", "TME"},
				DirectType:  "http://www.ft.com/ontology/company/PublicCompany",
				Aliases: []string{
					"Apple Inc",
					"Apple Computers",
					"Apple",
					"Apple Canada",
					"Apple Computer",
					"Apple Computer, Inc.",
					"APPLE INC",
					"Apple Incorporated",
					"Apple Computer Inc",
					"Apple Inc.",
					"Apple, Inc.",
				},
			},
			conceptType: "organisations",
		},
		{
			conceptModel: ConceptModel{
				UUID:         "2384fa7a-d514-3d6a-a0ea-3a711f66d0d8",
				DirectType:   "PublicCompany",
				PrefLabel:    "Apple, Inc.",
				Aliases:      []string{},
				IsDeprecated: true,
			},
			esConceptModel: EsConceptModel{
				Id:        "http://api.ft.com/things/2384fa7a-d514-3d6a-a0ea-3a711f66d0d8",
				Type:      "organisations",
				ApiUrl:    "http://api.ft.com/organisations/2384fa7a-d514-3d6a-a0ea-3a711f66d0d8",
				PrefLabel: "Apple, Inc.",
				Types: []string{
					"http://www.ft.com/ontology/core/Thing",
					"http://www.ft.com/ontology/concept/Concept",
					"http://www.ft.com/ontology/organisation/Organisation",
					"http://www.ft.com/ontology/company/Company",
					"http://www.ft.com/ontology/company/PublicCompany",
				},
				DirectType:   "http://www.ft.com/ontology/company/PublicCompany",
				Aliases:      []string{},
				IsDeprecated: true,
			},
			conceptType: "organisations",
		},
		{
			conceptModel: ConceptModel{
				UUID:         "2384fa7a-d514-3d6a-a0ea-3a711f66d0d9",
				DirectType:   "PublicCompany",
				PrefLabel:    "Apple, Inc.",
				Aliases:      []string{},
				IsDeprecated: true,
				ScopeNote:    "The Apple company used as a PublicCompany concept",
			},
			esConceptModel: EsConceptModel{
				Id:        "http://api.ft.com/things/2384fa7a-d514-3d6a-a0ea-3a711f66d0d9",
				Type:      "organisations",
				ApiUrl:    "http://api.ft.com/organisations/2384fa7a-d514-3d6a-a0ea-3a711f66d0d9",
				PrefLabel: "Apple, Inc.",
				Types: []string{
					"http://www.ft.com/ontology/core/Thing",
					"http://www.ft.com/ontology/concept/Concept",
					"http://www.ft.com/ontology/organisation/Organisation",
					"http://www.ft.com/ontology/company/Company",
					"http://www.ft.com/ontology/company/PublicCompany",
				},
				DirectType:   "http://www.ft.com/ontology/company/PublicCompany",
				Aliases:      []string{},
				IsDeprecated: true,
				ScopeNote:    "The Apple company used as a PublicCompany concept",
			},
			conceptType: "organisations",
		},
		{
			conceptModel: ConceptModel{
				UUID:       "02b8f22b-5b76-5600-89a2-df2676c0a923",
				DirectType: "SVProvision",
				PrefLabel:  "WaterSense",
				Aliases:    []string{"WaterSense"},
			},
			esConceptModel: EsConceptModel{
				Id:        "http://api.ft.com/things/02b8f22b-5b76-5600-89a2-df2676c0a923",
				Type:      "sv-provisions",
				ApiUrl:    "http://api.ft.com/concepts/02b8f22b-5b76-5600-89a2-df2676c0a923",
				PrefLabel: "WaterSense",
				Types: []string{
					"http://www.ft.com/ontology/core/Thing",
					"http://www.ft.com/ontology/concept/Concept",
					"http://www.ft.com/ontology/Provision",
					"http://www.ft.com/ontology/provision/SVProvision",
				},
				DirectType: "http://www.ft.com/ontology/provision/SVProvision",
				Aliases:    []string{"WaterSense"},
			},
			conceptType: "sv-provisions",
		},
		{
			conceptModel: ConceptModel{
				UUID:       "e0fc58d1-8dc5-47c6-90b1-59ccf8217366",
				DirectType: "SVCategory",
				PrefLabel:  "Policy and Regulation",
				Aliases:    []string{"Policy and Regulation"},
			},
			esConceptModel: EsConceptModel{
				Id:        "http://api.ft.com/things/e0fc58d1-8dc5-47c6-90b1-59ccf8217366",
				Type:      "sv-categories",
				ApiUrl:    "http://api.ft.com/things/e0fc58d1-8dc5-47c6-90b1-59ccf8217366",
				PrefLabel: "Policy and Regulation",
				Types: []string{
					"http://www.ft.com/ontology/core/Thing",
					"http://www.ft.com/ontology/concept/Concept",
					"http://www.ft.com/ontology/SVCategory",
				},
				DirectType: "http://www.ft.com/ontology/SVCategory",
				Aliases:    []string{"Policy and Regulation"},
			},
			conceptType: "sv-categories",
		},
	}

	for _, testModel := range tests {
		testTID := tid.NewTransactionID()

		actual, _ := ConvertConceptToESConceptModel(testModel.conceptModel, testModel.conceptType, testTID, publicAPIHost)
		esModel := actual.(*EsConceptModel)
		assert.Equal(t, testModel.esConceptModel.Id, esModel.Id, fmt.Sprintf("Expected Id %s differs from actual id %s ", testModel.esConceptModel.Id, esModel.Id))
		assert.Equal(t, testModel.esConceptModel.Type, esModel.Type, fmt.Sprintf("Expected Type %s differs from actual Type %s ", testModel.esConceptModel.Type, esModel.Type))
		assert.Equal(t, testModel.esConceptModel.ApiUrl, esModel.ApiUrl, fmt.Sprintf("Expected ApiUrl %s differs from actual ApiUrl %s ", testModel.esConceptModel.ApiUrl, esModel.ApiUrl))
		assert.Equal(t, testModel.esConceptModel.DirectType, esModel.DirectType, fmt.Sprintf("Expected DirectType %s differs from actual DirectType %s ", testModel.esConceptModel.DirectType, esModel.DirectType))
		assert.Equal(t, testModel.esConceptModel.PrefLabel, esModel.PrefLabel, fmt.Sprintf("Expected PrefLabel %s differs from actual PrefLabel %s ", testModel.esConceptModel.PrefLabel, esModel.PrefLabel))
		assert.Equal(t, testModel.esConceptModel.Types, esModel.Types, fmt.Sprintf("Expected Types %s differ from actual Types %s ", testModel.esConceptModel.Types, esModel.Types))
		assert.Equal(t, testModel.esConceptModel.Aliases, esModel.Aliases, fmt.Sprintf("Expected Aliases %s differ from actual Aliases %s ", testModel.esConceptModel.Aliases, esModel.Aliases))
		assert.Subset(t, testModel.esConceptModel.Authorities, esModel.Authorities, fmt.Sprintf("Expected Authorities %s differ from actual Authorities %s ", testModel.esConceptModel.Authorities, esModel.Authorities))
		assert.Equal(t, testTID, esModel.PublishReference)
		assert.Equal(t, testModel.esConceptModel.IsDeprecated, esModel.IsDeprecated, fmt.Sprintf("Expected IsDeprecated %t differ from actual IsDeprecated %t", testModel.esConceptModel.IsDeprecated, esModel.IsDeprecated))
		assert.Equal(t, testModel.esConceptModel.ScopeNote, esModel.ScopeNote, fmt.Sprintf("Expected ScopeNote %s differ from actual ScopeNote %s", testModel.esConceptModel.ScopeNote, esModel.ScopeNote))

		actualLastModified, err := time.Parse(time.RFC3339, esModel.LastModified)
		assert.NoError(t, err)
		assert.WithinDuration(t, time.Now(), actualLastModified, 3*time.Second)

	}
}

func TestConvertAggregateConceptToESConceptModel(t *testing.T) {

	tests := []struct {
		testName       string
		conceptModel   AggregateConceptModel
		esConceptModel EsConceptModel
	}{
		{
			testName: "AggregateConceptModel with Aliases",
			conceptModel: AggregateConceptModel{
				PrefUUID:   "2384fa7a-d514-3d6a-a0ea-3a711f66d0d8",
				DirectType: "PublicCompany",
				PrefLabel:  "Apple, Inc.",
				Aliases: []string{
					"Apple Inc",
					"Apple Computers",
					"Apple",
					"Apple Canada",
					"Apple Computer",
					"Apple Computer, Inc.",
					"APPLE INC",
					"Apple Incorporated",
					"Apple Computer Inc",
					"Apple Inc.",
					"Apple, Inc.",
				},
				SourceRepresentations: []SourceConcept{
					{
						UUID:      "xyz",
						Authority: "TME",
					},
					{
						UUID:      "abc",
						Authority: "Factset",
					},
				},
				CountryCode:            "US",
				CountryOfIncorporation: "US",
			},
			esConceptModel: EsConceptModel{
				Id:        "http://api.ft.com/things/2384fa7a-d514-3d6a-a0ea-3a711f66d0d8",
				Type:      "organisations",
				ApiUrl:    "http://api.ft.com/organisations/2384fa7a-d514-3d6a-a0ea-3a711f66d0d8",
				PrefLabel: "Apple, Inc.",
				Types: []string{
					"http://www.ft.com/ontology/core/Thing",
					"http://www.ft.com/ontology/concept/Concept",
					"http://www.ft.com/ontology/organisation/Organisation",
					"http://www.ft.com/ontology/company/Company",
					"http://www.ft.com/ontology/company/PublicCompany",
				},
				Authorities: []string{"TME", "Factset"},
				DirectType:  "http://www.ft.com/ontology/company/PublicCompany",
				Aliases: []string{
					"Apple Inc",
					"Apple Computers",
					"Apple",
					"Apple Canada",
					"Apple Computer",
					"Apple Computer, Inc.",
					"APPLE INC",
					"Apple Incorporated",
					"Apple Computer Inc",
					"Apple Inc.",
					"Apple, Inc.",
				},
				CountryCode:            "US",
				CountryOfIncorporation: "US",
			},
		},
		{
			testName: "Simple AggregateConceptModel",
			conceptModel: AggregateConceptModel{
				PrefUUID:     "2384fa7a-d514-3d6a-a0ea-3a711f66d0d8",
				DirectType:   "PublicCompany",
				PrefLabel:    "Apple, Inc.",
				Aliases:      []string{},
				IsDeprecated: true,
			},
			esConceptModel: EsConceptModel{
				Id:        "http://api.ft.com/things/2384fa7a-d514-3d6a-a0ea-3a711f66d0d8",
				Type:      "organisations",
				ApiUrl:    "http://api.ft.com/organisations/2384fa7a-d514-3d6a-a0ea-3a711f66d0d8",
				PrefLabel: "Apple, Inc.",
				Types: []string{
					"http://www.ft.com/ontology/core/Thing",
					"http://www.ft.com/ontology/concept/Concept",
					"http://www.ft.com/ontology/organisation/Organisation",
					"http://www.ft.com/ontology/company/Company",
					"http://www.ft.com/ontology/company/PublicCompany",
				},
				DirectType:   "http://www.ft.com/ontology/company/PublicCompany",
				Aliases:      []string{},
				IsDeprecated: true,
			},
		},
		{
			testName: "AggregateConceptModel with ScopeNote",
			conceptModel: AggregateConceptModel{
				PrefUUID:     "2384fa7a-d514-3d6a-a0ea-3a711f66d0d9",
				DirectType:   "PublicCompany",
				PrefLabel:    "Apple, Inc.",
				Aliases:      []string{},
				IsDeprecated: true,
				ScopeNote:    "The Apple company used as a PublicCompany concept",
			},
			esConceptModel: EsConceptModel{
				Id:        "http://api.ft.com/things/2384fa7a-d514-3d6a-a0ea-3a711f66d0d9",
				Type:      "organisations",
				ApiUrl:    "http://api.ft.com/organisations/2384fa7a-d514-3d6a-a0ea-3a711f66d0d9",
				PrefLabel: "Apple, Inc.",
				Types: []string{
					"http://www.ft.com/ontology/core/Thing",
					"http://www.ft.com/ontology/concept/Concept",
					"http://www.ft.com/ontology/organisation/Organisation",
					"http://www.ft.com/ontology/company/Company",
					"http://www.ft.com/ontology/company/PublicCompany",
				},
				DirectType:   "http://www.ft.com/ontology/company/PublicCompany",
				Aliases:      []string{},
				IsDeprecated: true,
				ScopeNote:    "The Apple company used as a PublicCompany concept",
			},
		},
		{
			testName: "AggregateConceptModel with NAICS",
			conceptModel: AggregateConceptModel{
				PrefUUID:   "2384fa7a-d514-3d6a-a0ea-3a711f66d0d8",
				DirectType: "PublicCompany",
				PrefLabel:  "Apple, Inc.",
				NAICS: []NAICS{
					{UUID: "1111-1111-1111-1111", Rank: 1},
					{UUID: "1111-1111-1111-1112", Rank: 2},
					{UUID: "1111-1111-1111-1113", Rank: 3},
				},
				SourceRepresentations: []SourceConcept{
					{
						UUID:      "xyz",
						Authority: "TME",
					},
					{
						UUID:      "abc",
						Authority: "Factset",
					},
				},
				CountryCode:            "US",
				CountryOfIncorporation: "US",
			},
			esConceptModel: EsConceptModel{
				Id:        "http://api.ft.com/things/2384fa7a-d514-3d6a-a0ea-3a711f66d0d8",
				Type:      "organisations",
				ApiUrl:    "http://api.ft.com/organisations/2384fa7a-d514-3d6a-a0ea-3a711f66d0d8",
				PrefLabel: "Apple, Inc.",
				Types: []string{
					"http://www.ft.com/ontology/core/Thing",
					"http://www.ft.com/ontology/concept/Concept",
					"http://www.ft.com/ontology/organisation/Organisation",
					"http://www.ft.com/ontology/company/Company",
					"http://www.ft.com/ontology/company/PublicCompany",
				},
				Authorities: []string{"TME", "Factset"},
				DirectType:  "http://www.ft.com/ontology/company/PublicCompany",
				NAICS: []NAICS{
					{UUID: "1111-1111-1111-1111", Rank: 1},
					{UUID: "1111-1111-1111-1112", Rank: 2},
					{UUID: "1111-1111-1111-1113", Rank: 3},
				},
				CountryCode:            "US",
				CountryOfIncorporation: "US",
			},
		},
	}

	for _, testModel := range tests {
		t.Run(testModel.testName, func(t *testing.T) {
			testTID := tid.NewTransactionID()

			actual, err := ConvertAggregateConceptToESConceptModel(testModel.conceptModel, "organisations", testTID, publicAPIHost)
			assert.NoError(t, err)
			esModel := actual.(*EsConceptModel)
			assert.Equal(t, testModel.esConceptModel.Id, esModel.Id, fmt.Sprintf("Expected Id %s differs from actual id %s ", testModel.esConceptModel.Id, esModel.Id))
			assert.Equal(t, testModel.esConceptModel.Type, esModel.Type, fmt.Sprintf("Expected Type %s differs from actual Type %s ", testModel.esConceptModel.Type, esModel.Type))
			assert.Equal(t, testModel.esConceptModel.ApiUrl, esModel.ApiUrl, fmt.Sprintf("Expected ApiUrl %s differs from actual ApiUrl %s ", testModel.esConceptModel.ApiUrl, esModel.ApiUrl))
			assert.Equal(t, testModel.esConceptModel.DirectType, esModel.DirectType, fmt.Sprintf("Expected DirectType %s differs from actual DirectType %s ", testModel.esConceptModel.DirectType, esModel.DirectType))
			assert.Equal(t, testModel.esConceptModel.PrefLabel, esModel.PrefLabel, fmt.Sprintf("Expected PrefLabel %s differs from actual PrefLabel %s ", testModel.esConceptModel.PrefLabel, esModel.PrefLabel))
			assert.Equal(t, testModel.esConceptModel.Types, esModel.Types, fmt.Sprintf("Expected Types %s differ from actual Types %s ", testModel.esConceptModel.Types, esModel.Types))
			assert.Equal(t, testModel.esConceptModel.Aliases, esModel.Aliases, fmt.Sprintf("Expected Aliases %s differ from actual Aliases %s ", testModel.esConceptModel.Aliases, esModel.Aliases))
			assert.Subset(t, testModel.esConceptModel.Authorities, esModel.Authorities, fmt.Sprintf("Expected Authorities %s differ from actual Authorities %s ", testModel.esConceptModel.Authorities, esModel.Authorities))
			assert.Equal(t, testTID, esModel.PublishReference)
			assert.Equal(t, testModel.esConceptModel.IsDeprecated, esModel.IsDeprecated, fmt.Sprintf("Expected IsDeprecated %t differ from actual IsDeprecated %t", testModel.esConceptModel.IsDeprecated, esModel.IsDeprecated))
			assert.Equal(t, testModel.esConceptModel.ScopeNote, esModel.ScopeNote, fmt.Sprintf("Expected ScopeNote %s differ from actual ScopeNote %s", testModel.esConceptModel.ScopeNote, esModel.ScopeNote))
			assert.Equal(t, testModel.esConceptModel.CountryCode, esModel.CountryCode, fmt.Sprintf("Expected CountryCode %s differ from actual CountryCode %s", testModel.esConceptModel.CountryCode, esModel.CountryCode))
			assert.Equal(t, testModel.esConceptModel.CountryOfIncorporation, esModel.CountryOfIncorporation, fmt.Sprintf("Expected CountryOfIncorporation %s differ from actual CountryOfIncorporation %s", testModel.esConceptModel.CountryOfIncorporation, esModel.CountryOfIncorporation))
			assert.Equal(t, testModel.esConceptModel.NAICS, esModel.NAICS, fmt.Sprintf("Expected NAICS %v differ from actual NAICS %v ", testModel.esConceptModel.NAICS, esModel.NAICS))

			actualLastModified, err := time.Parse(time.RFC3339, esModel.LastModified)
			assert.NoError(t, err)
			assert.WithinDuration(t, time.Now(), actualLastModified, 3*time.Second)
		})
	}
}

func TestConceptFuncsForConceptModel(t *testing.T) {
	concept := ConceptModel{}
	err := json.Unmarshal([]byte(testConceptModelJSON), &concept)
	require.NoError(t, err)

	expected := []string{"TME", "factsetIdentifier", "leiCode"}
	actual := concept.GetAuthorities()

	assert.Len(t, actual, 3)
	for _, val := range expected {
		assert.Contains(t, actual, val)
	}

	actual = concept.ConcordedUUIDs()
	assert.Empty(t, actual)
	assert.Equal(t, "2384fa7a-d514-3d6a-a0ea-3a711f66d0d8", concept.PreferredUUID())
}

func TestConceptFuncsForIntermediateConceptModel(t *testing.T) {
	concept := ConceptModel{}
	err := json.Unmarshal([]byte(testIntermediateConceptModelJSON), &concept)
	require.NoError(t, err)

	expected := []string{"TME"}
	actual := concept.GetAuthorities()

	assert.Len(t, actual, 1)
	for _, val := range expected {
		assert.Contains(t, actual, val)
	}

	expected = []string{}
	actual = concept.ConcordedUUIDs()
	assert.Equal(t, expected, actual)
	assert.Equal(t, "7e3f1354-53ba-3c3e-b9bb-5fcb8941df8c", concept.PreferredUUID())
}

func TestConceptFuncsForAggregatedConceptModel(t *testing.T) {
	concept := AggregateConceptModel{}
	err := json.Unmarshal([]byte(testAggregateConceptModelJSON), &concept)
	require.NoError(t, err)

	expected := []string{"TME", "Smartlogic"}
	actual := concept.GetAuthorities()

	assert.Len(t, actual, 2)
	for _, val := range expected {
		assert.Contains(t, actual, val)
	}

	expected = []string{"4ebbd9c4-3bb7-4d18-a14c-4c45aac5d966"}
	actual = concept.ConcordedUUIDs()
	assert.Equal(t, expected, actual)
	assert.Equal(t, "56388858-38d6-4dfc-a001-506394259b51", concept.PreferredUUID())
}

func TestConvertPersonToESConceptModel(t *testing.T) {

	tests := []struct {
		conceptModel         ConceptModel
		esPersonConceptModel EsPersonConceptModel
	}{
		{
			conceptModel: ConceptModel{
				UUID:       "0f07d468-fc37-3c44-bf19-a81f2aae9f36",
				DirectType: "Person",
				PrefLabel:  "Martin Wolf",
				Aliases:    []string{},
			},
			esPersonConceptModel: EsPersonConceptModel{
				EsConceptModel: &EsConceptModel{
					Id:        "http://api.ft.com/things/0f07d468-fc37-3c44-bf19-a81f2aae9f36",
					Type:      "people",
					ApiUrl:    "http://api.ft.com/people/0f07d468-fc37-3c44-bf19-a81f2aae9f36",
					PrefLabel: "Martin Wolf",
					Types: []string{
						"http://www.ft.com/ontology/core/Thing",
						"http://www.ft.com/ontology/concept/Concept",
						"http://www.ft.com/ontology/person/Person",
					},
					DirectType: "http://www.ft.com/ontology/person/Person",
					Aliases:    []string{},
				},
			},
		},
	}

	for _, testModel := range tests {
		testTID := tid.NewTransactionID()

		actual, err := ConvertConceptToESConceptModel(testModel.conceptModel, "people", testTID, publicAPIHost)
		assert.NoError(t, err)
		esModel := actual.(*EsPersonConceptModel)
		assert.Equal(t, testModel.esPersonConceptModel.Id, esModel.Id, fmt.Sprintf("Expected Id %s differs from actual id %s ", testModel.esPersonConceptModel.Id, esModel.Id))
		assert.Equal(t, testModel.esPersonConceptModel.Type, esModel.Type, fmt.Sprintf("Expected Type %s differs from actual Type %s ", testModel.esPersonConceptModel.Type, esModel.Type))
		assert.Equal(t, testModel.esPersonConceptModel.ApiUrl, esModel.ApiUrl, fmt.Sprintf("Expected ApiUrl %s differs from actual ApiUrl %s ", testModel.esPersonConceptModel.ApiUrl, esModel.ApiUrl))
		assert.Equal(t, testModel.esPersonConceptModel.DirectType, esModel.DirectType, fmt.Sprintf("Expected DirectType %s differs from actual DirectType %s ", testModel.esPersonConceptModel.DirectType, esModel.DirectType))
		assert.Equal(t, testModel.esPersonConceptModel.PrefLabel, esModel.PrefLabel, fmt.Sprintf("Expected PrefLabel %s differs from actual PrefLabel %s ", testModel.esPersonConceptModel.PrefLabel, esModel.PrefLabel))
		assert.Equal(t, testModel.esPersonConceptModel.Types, esModel.Types, fmt.Sprintf("Expected Types %s differ from actual Types %s ", testModel.esPersonConceptModel.Types, esModel.Types))
		assert.Equal(t, testModel.esPersonConceptModel.Aliases, esModel.Aliases, fmt.Sprintf("Expected Aliases %s differ from actual Aliases %s ", testModel.esPersonConceptModel.Aliases, esModel.Aliases))
		assert.Equal(t, testTID, esModel.PublishReference)

		actualLastModified, err := time.Parse(time.RFC3339, esModel.LastModified)
		assert.NoError(t, err)
		assert.WithinDuration(t, time.Now(), actualLastModified, 3*time.Second)
	}
}

func TestConvertMembershipToAggregateConceptModel(t *testing.T) {
	tests := []struct {
		testName              string
		aggregateConceptModel AggregateConceptModel
		esMembershipModel     EsMembershipModel
	}{
		{
			testName: "basic membership",
			aggregateConceptModel: AggregateConceptModel{
				PrefUUID:         "b159a539-527e-42ba-b5ee-29c33c0e016a",
				PersonUUID:       []string{"d52d8fdf-656c-4db3-b27c-06b16cdbb580"},
				OrganisationUUID: []string{"fa2b743d-f535-4deb-8524-df65bd536d09"},
				MembershipRoles: []AggregateMembershipRole{
					{RoleUUID: "c55f1d31-00fc-47a5-8a2e-19a967e07955", InceptionDate: "InceptionDate", TerminationDate: "TerminationDate"},
					{RoleUUID: "5c1f6da5-596e-4853-89b9-7f08652d366a", InceptionDate: "InceptionDate"},
				},
			},
			esMembershipModel: EsMembershipModel{
				Id:             "b159a539-527e-42ba-b5ee-29c33c0e016a",
				PersonId:       "d52d8fdf-656c-4db3-b27c-06b16cdbb580",
				OrganisationId: "fa2b743d-f535-4deb-8524-df65bd536d09",
				Memberships:    []string{"c55f1d31-00fc-47a5-8a2e-19a967e07955", "5c1f6da5-596e-4853-89b9-7f08652d366a"},
			},
		},
		{
			testName: "empty membership",
			aggregateConceptModel: AggregateConceptModel{
				PrefUUID:         "b159a539-527e-42ba-b5ee-29c33c0e016a",
				PersonUUID:       []string{"d52d8fdf-656c-4db3-b27c-06b16cdbb580"},
				OrganisationUUID: []string{"fa2b743d-f535-4deb-8524-df65bd536d09"},
			},
			esMembershipModel: EsMembershipModel{
				Id:             "b159a539-527e-42ba-b5ee-29c33c0e016a",
				PersonId:       "d52d8fdf-656c-4db3-b27c-06b16cdbb580",
				OrganisationId: "fa2b743d-f535-4deb-8524-df65bd536d09",
				Memberships:    make([]string, 0),
			},
		},
	}
	for _, testModel := range tests {
		t.Run(testModel.testName, func(t *testing.T) {
			testTID := tid.NewTransactionID()
			actual, _ := ConvertAggregateConceptToESConceptModel(testModel.aggregateConceptModel, "memberships", testTID, publicAPIHost)
			esModel := actual.(*EsMembershipModel)
			assert.Equal(t, testModel.esMembershipModel, *esModel)
		})
	}
}

func TestConvertPersonToAggregateConceptModel(t *testing.T) {
	tests := []struct {
		name                  string
		aggregateConceptModel AggregateConceptModel
		esPersonConceptModel  EsPersonConceptModel
	}{
		{
			name: "default to false",
			aggregateConceptModel: AggregateConceptModel{
				PrefUUID:   "0f07d468-fc37-3c44-bf19-a81f2aae9f36",
				DirectType: "Person",
				PrefLabel:  "Martin Wolf",
				Aliases:    []string{},
			},
			esPersonConceptModel: EsPersonConceptModel{
				EsConceptModel: &EsConceptModel{
					Id:        "http://api.ft.com/things/0f07d468-fc37-3c44-bf19-a81f2aae9f36",
					Type:      "people",
					ApiUrl:    "http://api.ft.com/people/0f07d468-fc37-3c44-bf19-a81f2aae9f36",
					PrefLabel: "Martin Wolf",
					Types: []string{
						"http://www.ft.com/ontology/core/Thing",
						"http://www.ft.com/ontology/concept/Concept",
						"http://www.ft.com/ontology/person/Person",
					},
					DirectType: "http://www.ft.com/ontology/person/Person",
					Aliases:    []string{},
				},
				IsFTAuthor: "false",
			},
		},
		{
			name: "matches on true",
			aggregateConceptModel: AggregateConceptModel{
				PrefUUID:   "0f07d468-fc37-3c44-bf19-a81f2aae9f36",
				DirectType: "Person",
				PrefLabel:  "Martin Wolf",
				Aliases:    []string{},
			},
			esPersonConceptModel: EsPersonConceptModel{
				EsConceptModel: &EsConceptModel{
					Id:        "http://api.ft.com/things/0f07d468-fc37-3c44-bf19-a81f2aae9f36",
					Type:      "people",
					ApiUrl:    "http://api.ft.com/people/0f07d468-fc37-3c44-bf19-a81f2aae9f36",
					PrefLabel: "Martin Wolf",
					Types: []string{
						"http://www.ft.com/ontology/core/Thing",
						"http://www.ft.com/ontology/concept/Concept",
						"http://www.ft.com/ontology/person/Person",
					},
					DirectType: "http://www.ft.com/ontology/person/Person",
					Aliases:    []string{},
				},
				IsFTAuthor: "false",
			},
		},
		{name: "matches on false",
			aggregateConceptModel: AggregateConceptModel{
				PrefUUID:   "0f07d468-fc37-3c44-bf19-a81f2aae9f36",
				DirectType: "Person",
				PrefLabel:  "Martin Wolf",
				Aliases:    []string{},
			},
			esPersonConceptModel: EsPersonConceptModel{
				EsConceptModel: &EsConceptModel{
					Id:        "http://api.ft.com/things/0f07d468-fc37-3c44-bf19-a81f2aae9f36",
					Type:      "people",
					ApiUrl:    "http://api.ft.com/people/0f07d468-fc37-3c44-bf19-a81f2aae9f36",
					PrefLabel: "Martin Wolf",
					Types: []string{
						"http://www.ft.com/ontology/core/Thing",
						"http://www.ft.com/ontology/concept/Concept",
						"http://www.ft.com/ontology/person/Person",
					},
					DirectType: "http://www.ft.com/ontology/person/Person",
					Aliases:    []string{},
				},
				IsFTAuthor: "false",
			},
		},
	}

	for _, testModel := range tests {
		t.Run(testModel.name, func(t *testing.T) {
			testTID := tid.NewTransactionID()

			actual, err := ConvertAggregateConceptToESConceptModel(testModel.aggregateConceptModel, "people", testTID, publicAPIHost)
			assert.NoError(t, err)
			esModel := actual.(*EsPersonConceptModel)
			assert.Equal(t, testModel.esPersonConceptModel.Id, esModel.Id, fmt.Sprintf("Expected Id %s differs from actual id %s ", testModel.esPersonConceptModel.Id, esModel.Id))
			assert.Equal(t, testModel.esPersonConceptModel.Type, esModel.Type, fmt.Sprintf("Expected Type %s differs from actual Type %s ", testModel.esPersonConceptModel.Type, esModel.Type))
			assert.Equal(t, testModel.esPersonConceptModel.ApiUrl, esModel.ApiUrl, fmt.Sprintf("Expected ApiUrl %s differs from actual ApiUrl %s ", testModel.esPersonConceptModel.ApiUrl, esModel.ApiUrl))
			assert.Equal(t, testModel.esPersonConceptModel.DirectType, esModel.DirectType, fmt.Sprintf("Expected DirectType %s differs from actual DirectType %s ", testModel.esPersonConceptModel.DirectType, esModel.DirectType))
			assert.Equal(t, testModel.esPersonConceptModel.PrefLabel, esModel.PrefLabel, fmt.Sprintf("Expected PrefLabel %s differs from actual PrefLabel %s ", testModel.esPersonConceptModel.PrefLabel, esModel.PrefLabel))
			assert.Equal(t, testModel.esPersonConceptModel.Types, esModel.Types, fmt.Sprintf("Expected Types %s differ from actual Types %s ", testModel.esPersonConceptModel.Types, esModel.Types))
			assert.Equal(t, testModel.esPersonConceptModel.Aliases, esModel.Aliases, fmt.Sprintf("Expected Aliases %s differ from actual Aliases %s ", testModel.esPersonConceptModel.Aliases, esModel.Aliases))
			assert.Equal(t, testTID, esModel.PublishReference)
			assert.Equal(t, testModel.esPersonConceptModel.IsFTAuthor, esModel.IsFTAuthor)

			actualLastModified, err := time.Parse(time.RFC3339, esModel.LastModified)
			assert.NoError(t, err)
			assert.WithinDuration(t, time.Now(), actualLastModified, 3*time.Second)
		})
	}
}

func TestReverse(t *testing.T) {
	tests := []struct {
		input          []string
		expectedResult []string
	}{
		{
			input:          []string{},
			expectedResult: []string{},
		},
		{
			input:          nil,
			expectedResult: nil,
		},
		{
			input:          []string{"foo"},
			expectedResult: []string{"foo"},
		},
		{
			input:          []string{"foo", "bar"},
			expectedResult: []string{"bar", "foo"},
		},
		{
			input:          []string{"foo", "bar", "word"},
			expectedResult: []string{"word", "bar", "foo"},
		},
	}

	for _, testCase := range tests {
		actualResult := reverse(testCase.input)
		assert.Equal(t, testCase.expectedResult, actualResult)
	}
}

func TestValidateEsConceptModelMarshalling(t *testing.T) {
	tests := []struct {
		testName           string
		inConcept          interface{}
		expectedResultJSON string
	}{
		{
			testName: "Check true deprecation flag",
			inConcept: &EsConceptModel{
				Id:           "http://api.ft.com/things/2384fa7a-d514-3d6a-a0ea-3a711f66d0d8",
				Type:         "organisations",
				ApiUrl:       "http://api.ft.com/organisations/2384fa7a-d514-3d6a-a0ea-3a711f66d0d8",
				PrefLabel:    "Apple, Inc.",
				DirectType:   "http://www.ft.com/ontology/company/PublicCompany",
				Aliases:      []string{},
				Types:        []string{},
				Authorities:  []string{},
				IsDeprecated: true,
			},
			expectedResultJSON: `{"id":"http://api.ft.com/things/2384fa7a-d514-3d6a-a0ea-3a711f66d0d8","type":"organisations","apiUrl":"http://api.ft.com/organisations/2384fa7a-d514-3d6a-a0ea-3a711f66d0d8","prefLabel":"Apple, Inc.","types":[],"authorities":[],"directType":"http://www.ft.com/ontology/company/PublicCompany","lastModified":"","publishReference":"","isDeprecated":true}`,
		},
		{
			testName: "Check false deprecation flag",
			inConcept: &EsConceptModel{
				Id:           "http://api.ft.com/things/2384fa7a-d514-3d6a-a0ea-3a711f66d0d8",
				Type:         "organisations",
				ApiUrl:       "http://api.ft.com/organisations/2384fa7a-d514-3d6a-a0ea-3a711f66d0d8",
				PrefLabel:    "Apple, Inc.",
				DirectType:   "http://www.ft.com/ontology/company/PublicCompany",
				Aliases:      []string{},
				Types:        []string{},
				Authorities:  []string{},
				IsDeprecated: false,
			},
			expectedResultJSON: `{"id":"http://api.ft.com/things/2384fa7a-d514-3d6a-a0ea-3a711f66d0d8","type":"organisations","apiUrl":"http://api.ft.com/organisations/2384fa7a-d514-3d6a-a0ea-3a711f66d0d8","prefLabel":"Apple, Inc.","types":[],"authorities":[],"directType":"http://www.ft.com/ontology/company/PublicCompany","lastModified":"","publishReference":""}`,
		},
		{
			testName: "Check scopeNote with value",
			inConcept: &EsConceptModel{
				Id:          "http://api.ft.com/things/2384fa7a-d514-3d6a-a0ea-3a711f66d0d8",
				Type:        "organisations",
				ApiUrl:      "http://api.ft.com/organisations/2384fa7a-d514-3d6a-a0ea-3a711f66d0d8",
				PrefLabel:   "Apple, Inc.",
				DirectType:  "http://www.ft.com/ontology/company/PublicCompany",
				Aliases:     []string{},
				Types:       []string{},
				Authorities: []string{},
				ScopeNote:   "scope note dummy value",
			},
			expectedResultJSON: `{"id":"http://api.ft.com/things/2384fa7a-d514-3d6a-a0ea-3a711f66d0d8","type":"organisations","apiUrl":"http://api.ft.com/organisations/2384fa7a-d514-3d6a-a0ea-3a711f66d0d8","prefLabel":"Apple, Inc.","types":[],"authorities":[],"directType":"http://www.ft.com/ontology/company/PublicCompany","lastModified":"","publishReference":"","scopeNote":"scope note dummy value"}`,
		},
		{
			testName: "Check scopeNote with no value",
			inConcept: &EsConceptModel{
				Id:          "http://api.ft.com/things/2384fa7a-d514-3d6a-a0ea-3a711f66d0d8",
				Type:        "organisations",
				ApiUrl:      "http://api.ft.com/organisations/2384fa7a-d514-3d6a-a0ea-3a711f66d0d8",
				PrefLabel:   "Apple, Inc.",
				DirectType:  "http://www.ft.com/ontology/company/PublicCompany",
				Aliases:     []string{},
				Types:       []string{},
				Authorities: []string{},
				ScopeNote:   "",
			},
			expectedResultJSON: `{"id":"http://api.ft.com/things/2384fa7a-d514-3d6a-a0ea-3a711f66d0d8","type":"organisations","apiUrl":"http://api.ft.com/organisations/2384fa7a-d514-3d6a-a0ea-3a711f66d0d8","prefLabel":"Apple, Inc.","types":[],"authorities":[],"directType":"http://www.ft.com/ontology/company/PublicCompany","lastModified":"","publishReference":""}`,
		},
	}

	for _, testCase := range tests {
		inConceptByteArr, err := json.Marshal(testCase.inConcept)
		assert.NoError(t, err, fmt.Sprintf("%s -> error during marshalling", testCase.testName))

		assert.Equal(t, string(inConceptByteArr), testCase.expectedResultJSON, fmt.Sprintf("%s -> expected json string not equals with actual", testCase.testName))
	}
}
