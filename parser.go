package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"sort"
	"strconv"
)

const (
	Country = "Country"
	Station = "Station"
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

type BrowseRoutesReply struct {
	XMLName    xml.Name      `xml:"BrowseRoutesResponseApiDto"`
	Currencies []CurrencyDto `xml:">CurrencyDto"`
	Routes     []RouteDto    `xml:">RouteDto"`
	Quotes     []QuoteDto    `xml:">QuoteDto"`
	Places     []PlaceDto    `xml:">PlaceDto"`
	Carriers   []CarriersDto `xml:">CarriersDto"`
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
	MinPrice      float64
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

func ParseBrowseRoutesReply(data []byte) *BrowseRoutesReply {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	anywhere := &BrowseRoutesReply{}
	decoder.Decode(anywhere)
	return anywhere
}

func ParseBrowseRoutesReplyJson(data []byte) *BrowseRoutesReply {
	decoder := json.NewDecoder(bytes.NewReader(data))
	anywhere := &BrowseRoutesReply{}
	decoder.Decode(anywhere)
	return anywhere
}

func parse() {

	fileName := "testdata/anywhere.xml"
	data, err := os.Open(fileName)

	if err != nil {
		panic(err)
	}
	defer data.Close()
	decoder := xml.NewDecoder(data)
	var anywhere BrowseRoutesReply
	decoder.Decode(&anywhere)
	fmt.Println(anywhere)
}

func (m *BrowseRoutesReply) PrintStats() {

	fmt.Println("Currencies: ", len(m.Currencies))
	fmt.Println("Quotes: ", len(m.Quotes))
	fmt.Println("Routes: ", len(m.Routes))
	fmt.Println("Places: ", len(m.Places))
	fmt.Println("Carriers: ", len(m.Carriers))
}

type Quotes []QuoteDto

func (slice Quotes) Len() int {
	return len(slice)
}

func (slice Quotes) Less(i, j int) bool {
	return slice[i].MinPrice < slice[j].MinPrice
}

func (slice Quotes) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

type Routes []RouteDto

func (slice Routes) Less(i, j int) bool {
	return slice[i].GetPrice() < slice[j].GetPrice()
}

func (slice Routes) Len() int {
	return len(slice)
}

func (slice Routes) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

func (m *BrowseRoutesReply) FilterRoutes() []RouteDto {
	results := make([]RouteDto, 0, len(m.Routes))
	for _, route := range m.Routes {
		if route.Valid() {
			results = append(results, route)
		}
	}
	return results
}

func (m *BrowseRoutesReply) GetPlacesByPrice2() []string {

	var routes Routes
	routes = m.FilterRoutes()
	sort.Sort(routes)

	places := m.GetPlacesById()

	results := make([]string, len(routes))
	for index, route := range routes {
		place := places[route.DestinationId]
		fmt.Printf("%s %s\n", route.Price, place.SkyscannerCode)
		results[index] = place.SkyscannerCode
	}

	return results
}

func (m *BrowseRoutesReply) GetPriceByDestination() map[string]float64 {
	results := make(map[string]float64)
	places := m.GetPlacesById()

	for _, quote := range m.Quotes {
		if quote.IsReturn() {
			place := places[quote.OutboundLeg.DestinationId]
			code := place.SkyscannerCode
			results[code] = quote.MinPrice
		}
	}
	return results
}

func (m *BrowseRoutesReply) GetPlacesById() map[int]PlaceDto {
	places := make(map[int]PlaceDto)
	for _, place := range m.Places {
		places[place.PlaceId] = place
	}
	return places
}

func (m *BrowseRoutesReply) GetPlacesByPrice() []string {
	var quotes Quotes
	quotes = m.Quotes
	sort.Sort(quotes)

	places := m.GetPlacesById()

	results := make([]string, len(quotes))
	for index, quote := range quotes {
		destination := quote.OutboundLeg.DestinationId
		if destination != 0 {
			place := places[quote.OutboundLeg.DestinationId]
			results[index] = place.SkyscannerCode
		}
	}

	return results
}

func (m *BrowseRoutesReply) GetCountries() []PlaceDto {
	results := make([]PlaceDto, 0, len(m.Places))
	for _, place := range m.Places {
		if place.IsCountry() {
			results = append(results, place)
		}
	}
	return results

}

func (m *RouteDto) GetPrice() int {
	price, err := strconv.Atoi(string(m.Price))
	if err != nil {
		panic(err)
	}
	return price
}

func (m *RouteDto) Valid() bool {
	_, err := strconv.Atoi(string(m.Price))
	return err == nil
}

func (m *PlaceDto) IsCountry() bool {
	return m.Type == Country
}

func (m *QuoteDto) IsReturn() bool {
	return m.OutboundLeg.DestinationId != 0 &&
		m.InboundLeg.DestinationId != 0
}

type FullQuote struct {
	Quote       QuoteDto
	Destination PlaceDto
}

type FullQuotes []FullQuote

func (slice FullQuotes) Len() int {
	return len(slice)
}

func (slice FullQuotes) Less(i, j int) bool {
	return slice[i].Quote.MinPrice < slice[j].Quote.MinPrice
}

func (slice FullQuotes) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

func (m *BrowseRoutesReply) GetFullQuotes() FullQuotes {
	results := make(FullQuotes, 0, len(m.Quotes))

	places := m.GetPlacesById()
	for _, quote := range m.Quotes {
		if quote.IsReturn() {

			destination := places[quote.OutboundLeg.DestinationId]
			results = append(results, FullQuote{quote, destination})
		}
	}
	return results
}
