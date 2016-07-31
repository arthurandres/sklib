package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
)

type CurrencyDto struct {
	Code                        string
	Symbol                      string
	DecimalSeparator            string
	SymbolOnLeft                bool
	SpaceBetweenAmountAndSymbol bool
	RoundingCoefficient         int
	DecimalDigits               int
}

type CurrencyDtos struct {
	Currencies []CurrencyDto `xml:"CurrencyDto"`
}

type AnywhereQuery struct {
	XMLName    xml.Name       `xml:"BrowseRoutesResponseApiDto"`
	Currencies []CurrencyDtos `xml:">CurrencyDto"`
	Routes     []RouteDto     `xml:">RouteDto"`
	Quotes     []QuoteDto     `xml:">QuoteDto"`
	Places     []PlaceDto     `xml:">PlaceDto"`
	Carriers   []CarriersDto  `xml:">CarriersDto"`
}

type RouteDto struct {
	QuoteDateTime string
	Price         json.Number
	QuoteIds      []int `xml:">int"`
	DestinationId int
	OriginId      int
}

type Leg struct {
	CarrierIds    []int `xml:">int"`
	OriginId      int
	DestinationId int
	DepartureDate string
}

type QuoteDto struct {
	QuoteId       int
	MinPrice      int
	Direct        bool
	OutboundLeg   Leg
	InboundLeg    Leg
	QuoteDateTime string
}

type PlaceDto struct {
	PlaceId        int
	Name           string
	Type           string
	SkyscannerCode string
}

type CarriersDto struct {
	CarrierId int
	Name      string
}

func ParseAnywhereQuery(data []byte) AnywhereQuery {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	var anywhere AnywhereQuery
	decoder.Decode(&anywhere)
	return anywhere
}

func parse() {

	fileName := "testdata/anywhere.xml"
	data, err := os.Open(fileName)

	if err == nil {
		defer data.Close()
		decoder := xml.NewDecoder(data)
		var anywhere AnywhereQuery
		decoder.Decode(&anywhere)
		fmt.Println(anywhere)
	} else {
		fmt.Println(err)
	}
}

func (m *AnywhereQuery) PrintStats() {

	fmt.Println("Currencies: ", len(m.Currencies))
	fmt.Println("Quotes: ", len(m.Quotes))
	fmt.Println("Routes: ", len(m.Routes))
	fmt.Println("Places: ", len(m.Places))
	fmt.Println("Carriers: ", len(m.Carriers))
}
