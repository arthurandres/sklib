package sklib

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"math"
	"net/url"
	"sort"
	"strconv"
)

type LiveRequest struct {
	Localisation  Localisation
	Origin        string
	Destination   string
	DepartureDate string
	ReturnDate    string
}

// SearchRequest Not a DTO?
type SearchRequest struct {
	Localisation  Localisation
	Origin        string
	Destinations  []string
	DepartureDate string
	ReturnDate    string
}

type Localisation struct {
	Country  string
	Currency string
	Language string
}

type BrowseRoutesRequest struct {
	Localisation  Localisation
	Origin        string
	Destination   string
	DepartureDate string
	ReturnDate    string
}

type CurrencyDto struct {
	Code                        string
	Symbol                      string
	ThousandsSeparator          string
	DecimalSeparator            string
	SymbolOnLeft                bool
	SpaceBetweenAmountAndSymbol bool
	RoundingCoefficient         int
	DecimalDigits               int
}

type LocalesReply struct {
	XMLName xml.Name    `xml:"ReferenceServiceResponseDto"`
	Locales []LocaleDto `xml:">LocaleDto"`
}

type LocaleDto struct {
	Code string
	Name string
}

type CountriesReply struct {
	XMLName   xml.Name     `xml:"ReferenceServiceResponseDto"`
	Countries []CountryDto `xml:">CountryDto"`
}

type CountryDto struct {
	Code string
	Name string
}

type CurrenciesReply struct {
	XMLName    xml.Name      `xml:"ReferenceServiceResponseDto"`
	Currencies []CurrencyDto `xml:">CurrencyDto"`
}

type BrowseRoutesReply struct {
	XMLName    xml.Name      `xml:"BrowseRoutesResponseApiDto"`
	Currencies []CurrencyDto `xml:">CurrencyDto"`
	Routes     []RouteDto    `xml:">RouteDto"`
	Quotes     []QuoteDto    `xml:">QuoteDto"`
	Places     []PlaceDto    `xml:">PlaceDto"`
	Carriers   []CarriersDto `xml:">CarriersDto"`
}

type LiveReply struct {
	XMLName     xml.Name `xml:"PollSessionResponseDto"`
	SessionKey  string
	Status      string
	Query       LiveQueryDto
	Segments    []SegmentApiDto      `xml:">SegmentApiDto"`
	Carriers    []CarrierApiDto      `xml:">CarrierApiDto"`
	Agents      []AgentApiDto        `xml:">AgentApiDto"`
	Places      []PlaceApiDto        `xml:">PlaceApiDto"`
	Currencies  []CurrencyDto        `xml:">CurrencyDto"`
	Legs        []ItineraryLegApiDto `xml:">ItineraryLegApiDto"`
	Itineraries []ItineraryApiDto    `xml:">ItineraryApiDto"`
}

type RouteDto struct {
	QuoteDateTime string
	Price         json.Number
	QuoteIds      []int `xml:">int"`
	DestinationId int
	OriginId      int
}

type Routes []RouteDto

type LegDto struct {
	CarrierIds    []int `xml:">int"`
	OriginId      int
	DestinationId int
	DepartureDate string
}

type QuoteDto struct {
	QuoteId       int
	MinPrice      float64
	Direct        bool
	OutboundLeg   LegDto
	InboundLeg    LegDto
	QuoteDateTime string
}

type Quotes []QuoteDto

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

type CarrierApiDto struct {
	Id          int
	Code        string
	Name        string
	ImageUrl    string
	DisplayCode string
}

type PlaceApiDto struct {
	Id       int
	ParentId json.Number
	Code     string
	Type     string
	Name     string
}

type AgentApiDto struct {
	Id                 int
	Name               string
	ImageUrl           string
	Status             string
	OptimisedForMobile bool
	BookingNumber      string
	Type               string
}

type SegmentApiDto struct {
	Id                 int
	OriginStation      int
	DestinationStation int
	DepartureDateTime  string
	ArrivalDateTime    string
	Carrier            int
	OperatingCarrier   int
	Duration           int
	FlightNumber       string
	JourneyMode        string
	Directionality     string
}

type FlightNumberDto struct {
	FlightNumber string
	CarrierId    int
}

type ItineraryLegApiDto struct {
	Id                 string
	SegmentIds         []int `xml:">int"`
	OriginStation      int
	DestinationStation int
	Departure          string
	Arrival            string
	Duration           int
	JourneyMode        string
	Stops              []int `xml:">int"`
	Carriers           []int `xml:">int"`
	OperatingCarriers  []int `xml:">int"`
	Directionality     string
	FlightNumbers      []FlightNumberDto `xml:">FlightNumberDto"`
}

type PricingOptionApiDto struct {
	Agents            []int `xml:">int"`
	QuoteAgeInMinutes int
	Price             float64
	DeeplinkUrl       string
}

type BookingDetailsLinkDto struct {
	Uri    string
	Body   string
	Method string
}

type LiveQueryDto struct {
	Country          string
	Currency         string
	Locale           string
	Adults           int
	Children         int
	Infants          int
	OriginPlace      json.Number
	DestinationPlace json.Number
	OutboundDate     string
	InboundDate      string
	LocationSchema   string
	CabinClass       string
	GroupPricing     bool
}

type ItineraryApiDto struct {
	OutboundLegId      string
	InboundLegId       string
	PricingOptions     []PricingOptionApiDto `xml:">PricingOptionApiDto"`
	BookingDetailsLink BookingDetailsLinkDto
}

type FullQuote struct {
	Quote       QuoteDto
	Destination PlaceDto
}

type FullQuotes []FullQuote

func (m *BrowseRoutesReply) FilterRoutes() []RouteDto {
	results := make([]RouteDto, 0, len(m.Routes))
	for _, route := range m.Routes {
		if route.Valid() {
			results = append(results, route)
		}
	}
	return results
}

func (m *BrowseRoutesReply) GetPlacesByPrice() []string {

	var routes Routes
	routes = m.FilterRoutes()
	sort.Sort(routes)

	places := m.GetPlacesById()

	results := make([]string, len(routes))
	for index, route := range routes {
		place := places[route.DestinationId]
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

func (m *BrowseRoutesReply) GetCountries() []PlaceDto {
	results := make([]PlaceDto, 0, len(m.Places))
	for _, place := range m.Places {
		if place.IsCountry() {
			results = append(results, place)
		}
	}
	return results
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

func (m *BrowseRoutesReply) GetBestQuotes() FullQuotes {
	results := make(FullQuotes, 0, len(m.Quotes))
	mapping := make(map[int]QuoteDto)

	places := m.GetPlacesById()
	for _, quote := range m.Quotes {
		if quote.IsReturn() {
			bestQuote, exists := mapping[quote.OutboundLeg.DestinationId]
			if !exists || quote.MinPrice < bestQuote.MinPrice {
				mapping[quote.OutboundLeg.DestinationId] = quote
			}
		}
	}

	for _, quote := range mapping {
		destination := places[quote.OutboundLeg.DestinationId]
		results = append(results, FullQuote{quote, destination})
	}
	return results
}

func (m *BrowseRoutesReply) GetBestPrice() float64 {
	result := math.MaxFloat64
	for _, quote := range m.Quotes {
		if quote.IsReturn() && quote.MinPrice < result {
			result = quote.MinPrice
		}
	}
	return result
}

func (slice Quotes) Len() int {
	return len(slice)
}

func (slice Quotes) Less(i, j int) bool {
	return slice[i].MinPrice < slice[j].MinPrice
}

func (slice Quotes) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

func (slice Routes) Less(i, j int) bool {
	return slice[i].GetPrice() < slice[j].GetPrice()
}

func (slice Routes) Len() int {
	return len(slice)
}

func (slice Routes) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
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
	return m.Type == CountryValue
}

func (m *QuoteDto) IsReturn() bool {
	return m.OutboundLeg.DestinationId != 0 &&
		m.InboundLeg.DestinationId != 0
}

func (slice FullQuotes) Len() int {
	return len(slice)
}

func (slice FullQuotes) Less(i, j int) bool {
	return slice[i].Quote.MinPrice < slice[j].Quote.MinPrice
}

func (slice FullQuotes) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

func (m *LiveReply) Stats() map[string]int {
	results := make(map[string]int)
	results["Itineraries"] = len(m.Itineraries)
	results["Legs"] = len(m.Legs)
	results["Segments"] = len(m.Segments)
	results["Carriers"] = len(m.Carriers)
	results["Agents"] = len(m.Agents)
	results["Places"] = len(m.Places)
	results["Currencies"] = len(m.Currencies)
	return results
}

func (m *BrowseRoutesReply) Stats() map[string]int {

	results := make(map[string]int)
	results["Currencies"] = len(m.Currencies)
	results["Quotes"] = len(m.Quotes)
	results["Routes"] = len(m.Routes)
	results["Places"] = len(m.Places)
	results["Carriers"] = len(m.Carriers)
	return results
}

func (m *Localisation) SubURL() string {
	return fmt.Sprintf("%s/%s/%s",
		m.Country,
		m.Currency,
		m.Language)
}

func NewBrowseRouteRequest(local Localisation, origin string, departureDate string, returnDate string) BrowseRoutesRequest {
	return BrowseRoutesRequest{local, origin, anywhere, departureDate, returnDate}
}

func (m BrowseRoutesRequest) Url() string {
	return fmt.Sprintf(
		browseRouteFormat,
		m.Localisation.SubURL(),
		m.Origin,
		m.Destination,
		m.DepartureDate,
		m.ReturnDate)
}

func (m *LiveRequest) Values() url.Values {
	return url.Values{
		"country":          {m.Localisation.Country},
		"currency":         {m.Localisation.Currency},
		"locale":           {m.Localisation.Language},
		"originplace":      {m.Origin},
		"destinationplace": {m.Destination},
		"outbounddate":     {FormatDateToForm(m.DepartureDate)},
		"inbounddate":      {FormatDateToForm(m.ReturnDate)},
		"locationschema":   {"Iata"}}
}

func (m *LiveRequest) Encode() string {
	return m.Values().Encode()
}

func NewLiveRequest(
	localisation Localisation,
	origin string,
	destination string,
	departureDate string,
	returnDate string) LiveRequest {

	return LiveRequest{
		localisation,
		origin,
		destination,
		departureDate,
		returnDate}
}

func (m FullQuotes) GetTowns() []PlaceDto {
	townsMap := make(map[string]PlaceDto)
	for _, v := range m {
		townsMap[v.Destination.SkyscannerCode] = v.Destination
	}
	towns := make([]PlaceDto, 0, 0)
	for _, v := range townsMap {
		towns = append(towns, v)
	}
	return towns
}

func (m FullQuotes) FilterDirects() FullQuotes {
	results := make(FullQuotes, 0, len(m))
	for _, quote := range m {
		if quote.Quote.Direct {
			results = append(results, quote)
		}
	}
	return results
}

func GetSearchLink(from string, to string, departureDate string, returnDate string) string {
	return fmt.Sprintf(linkBase, from, to, departureDate, returnDate)
}
