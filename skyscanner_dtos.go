package sklib

import (
	"encoding/json"
	"encoding/xml"
)

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
