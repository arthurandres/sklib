package sklib

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"math"
	"sort"
	"strconv"
	"time"
)

const (
	CountryValue          = "Country"
	StationValue          = "Station"
	CityValue             = "City"
	AirportValue          = "Airport"
	UpdatesPendingStatus  = "UpdatesPending"
	UpdatesCompleteStatus = "UpdatesComplete"
	DateFormatUrl         = "20060102"
	DateFormatForm        = "2006-01-02"
	DateTimeFormat        = "2006-01-02T15:04:05"
)

func ParseDateTime(dateTime string) (time.Time, error) {
	return time.Parse(DateTimeFormat, dateTime)
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

func FormatDate(input string) string {
	date, err := time.Parse(DateFormatUrl, input)
	if err != nil {
		panic(err)
	}
	return date.Format(DateFormatForm)
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

func ParseJson(data []byte, output interface{}) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	return decoder.Decode(output)
}
