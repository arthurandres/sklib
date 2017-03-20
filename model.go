package sklib

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"time"
)

type Place struct {
	Name   string
	Code   string
	Type   string
	Parent *Place
}

type PlaceMap map[int]*Place
type Places []*Place

type Carrier struct {
	Name        string
	Code        string
	ImageUrl    string
	DisplayCode string
}

type CarrierMap map[int]*Carrier
type Carriers []*Carrier

type Segment struct {
	Origin           *Place
	Destination      *Place
	Departure        time.Time
	Arrival          time.Time
	Carrier          *Carrier
	OperatingCarrier *Carrier
	Duration         time.Duration
	FlightNumber     string
	JourneyMode      string
	Directionality   string
}

type SegmentMap map[int]*Segment
type Segments []*Segment

type Leg struct {
	Segments          Segments
	Origin            *Place
	Destination       *Place
	Departure         time.Time
	Arrival           time.Time
	Duration          time.Duration
	JourneyMode       string
	Stops             Places
	Carriers          Carriers
	OperatingCarriers Carriers
	Directionality    string
	FlightNumbers     FlightNumbers
}

type LegMap map[string]*Leg
type Legs []Leg

type Agent struct {
	Name               string
	ImageUrl           string
	Status             string
	OptimisedForMobile bool
	Type               string
}

type Agents []*Agent
type AgentMap map[int]*Agent

type PricingOption struct {
	Agents      []*Agent
	Age         time.Duration
	Price       float64
	DeeplinkUrl string
}

type PricingOptions []*PricingOption

type Itinerary struct {
	OutboundLeg    *Leg
	InboundLeg     *Leg
	PricingOptions PricingOptions
}

type Itineraries []*Itinerary

type Currency CurrencyDto
type Currencies []*Currency

type FlightsData struct {
	Currencies  Currencies
	Itineraries Itineraries
}

type FlightNumber struct {
	Number  string
	Carrier *Carrier
}

type FlightNumbers []FlightNumber

type ItineraryFilter interface {
	Filter(itinerary *Itinerary) bool
}

func ReadLiveReply(input *LiveReply) (*FlightsData, error) {
	places, err := MapPlaces(input.Places)
	if err != nil {
		return nil, err
	}
	carriers, err := MapCarriers(input.Carriers)
	if err != nil {
		return nil, err
	}
	segments, err := ReadSegments(input.Segments, places, carriers)
	if err != nil {
		return nil, err
	}
	legs, err := ReadLegs(input.Legs, places, carriers, segments)
	if err != nil {
		return nil, err
	}
	agents, err := ReadAgents(input.Agents)
	if err != nil {
		return nil, err
	}
	itineraries, err := ReadItineraries(input.Itineraries, legs, agents)
	if err != nil {
		return nil, err
	}
	currencies, err := ReadCurrencies(input.Currencies)
	if err != nil {
		return nil, err
	}
	return &FlightsData{
			Currencies:  currencies,
			Itineraries: itineraries},
		nil
}

func getParentPlace(parentId string, mapping PlaceMap) *Place {
	if len(parentId) == 0 {
		return nil
	}
	parentIdInt, err := strconv.Atoi(parentId)
	if err != nil {
		return nil
	}
	parent, ok := mapping[parentIdInt]
	if !ok {
		fmt.Printf("%s %s %s\n", parentIdInt, parentId, len(mapping))
		panic("Could not find parent " + parentId)
	}
	return parent

}

type PlaceApiDtos []PlaceApiDto

func (slice PlaceApiDtos) Len() int {
	return len(slice)
}

func GetPlaceTypeValue(placeType string) int {
	switch placeType {
	case CountryValue:
		return 1
	case CityValue:
		return 2
	case AirportValue:
		return 3
	default:
		panic(fmt.Errorf("Unknown place type: %s", placeType))
	}
}

func ComparePlaceType(left, right string) bool {
	return GetPlaceTypeValue(left) < GetPlaceTypeValue(right)
}

func (slice PlaceApiDtos) Less(i, j int) bool {
	left := slice[i]
	right := slice[j]
	if left.Type == right.Type {
		return left.Code < right.Code
	} else {
		return ComparePlaceType(left.Type, right.Type)
	}
}

func (slice PlaceApiDtos) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

func MapPlaces(inputNotSorted []PlaceApiDto) (PlaceMap, error) {
	input := PlaceApiDtos(inputNotSorted)
	sort.Sort(input)

	results := make(PlaceMap)
	for _, placeDto := range input {
		if _, exists := results[placeDto.Id]; exists {
			return nil, fmt.Errorf("Duplicate place %s", placeDto)
		}
		parent := getParentPlace(string(placeDto.ParentId), results)
		place := &Place{Code: placeDto.Code, Name: placeDto.Name, Parent: parent, Type: placeDto.Type}
		results[placeDto.Id] = place
	}
	return results, nil
}

func ReadSegment(
	dto SegmentApiDto,
	places PlaceMap,
	carriers CarrierMap) (*Segment, error) {

	origin := places[dto.OriginStation]
	destination := places[dto.DestinationStation]
	departure, err := ParseDateTime(dto.DepartureDateTime)
	if err != nil {
		return nil, err
	}
	arrival, err := ParseDateTime(dto.ArrivalDateTime)
	if err != nil {
		return nil, err
	}
	carrier := carriers[dto.Carrier]
	operatingCarrier := carriers[dto.OperatingCarrier]
	duration := time.Minute * time.Duration(dto.Duration)

	return &Segment{
			Origin:           origin,
			Destination:      destination,
			Departure:        departure,
			Arrival:          arrival,
			Carrier:          carrier,
			OperatingCarrier: operatingCarrier,
			Duration:         duration,
			FlightNumber:     dto.FlightNumber,
			JourneyMode:      dto.JourneyMode,
			Directionality:   dto.Directionality},
		nil
}

func ReadSegments(
	dtos []SegmentApiDto,
	places PlaceMap,
	carriers CarrierMap) (SegmentMap, error) {
	results := make(SegmentMap)

	for _, dto := range dtos {
		segment, err := ReadSegment(dto, places, carriers)
		if err != nil {
			return nil, err
		}
		if _, exists := results[dto.Id]; exists {
			return nil, fmt.Errorf("Duplicate segment %d", dto.Id)
		}
		results[dto.Id] = segment
	}

	return results, nil
}

func MapCarriers(input []CarrierApiDto) (CarrierMap, error) {
	results := make(CarrierMap)
	for _, dto := range input {
		if _, e := results[dto.Id]; e {
			return nil, fmt.Errorf("Duplicate carrier %s", dto.Name)
		}
		results[dto.Id] = &Carrier{Code: dto.Code, Name: dto.Name, ImageUrl: dto.ImageUrl, DisplayCode: dto.DisplayCode}
	}
	return results, nil
}

func ReadLegs(
	dtos []ItineraryLegApiDto,
	placeMap PlaceMap,
	carrierMap CarrierMap,
	segmentMap SegmentMap) (LegMap, error) {
	results := make(LegMap)
	for _, dto := range dtos {
		if _, e := results[dto.Id]; e {
			return nil, fmt.Errorf("Duplicat leg %s", dto.Id)
		}
		leg, err := ReadLeg(dto, placeMap, carrierMap, segmentMap)
		if err != nil {
			return nil, err
		}
		results[dto.Id] = leg
	}
	return results, nil
}

func ReadLeg(
	dto ItineraryLegApiDto,
	placeMap PlaceMap,
	carrierMap CarrierMap,
	segmentMap SegmentMap) (*Leg, error) {

	segments, err := FindSegments(dto.SegmentIds, segmentMap)
	if err != nil {
		return nil, err
	}
	origin, exists := placeMap[dto.OriginStation]
	if !exists {
		return nil, fmt.Errorf("Missing origin %d", dto.OriginStation)
	}
	destination, exists := placeMap[dto.DestinationStation]
	if !exists {
		return nil, fmt.Errorf("Missing destination %d", dto.DestinationStation)
	}
	departure, err := ParseDateTime(dto.Departure)
	if err != nil {
		return nil, err
	}
	arrival, err := ParseDateTime(dto.Arrival)
	if err != nil {
		return nil, err
	}
	duration := time.Minute * time.Duration(dto.Duration)

	stops, err := FindPlaces(dto.Stops, placeMap)
	if err != nil {
		return nil, err
	}

	carriers, err := FindCarriers(dto.Carriers, carrierMap)
	if err != nil {
		return nil, err
	}
	operatingCarriers, err := FindCarriers(dto.OperatingCarriers, carrierMap)
	if err != nil {
		return nil, err
	}

	flightNumbers, err := FindFlightNumbers(dto.FlightNumbers, carrierMap)
	if err != nil {
		return nil, err
	}

	return &Leg{
			Segments:          segments,
			Origin:            origin,
			Destination:       destination,
			Departure:         departure,
			Arrival:           arrival,
			Duration:          duration,
			JourneyMode:       dto.JourneyMode,
			Stops:             stops,
			Carriers:          carriers,
			OperatingCarriers: operatingCarriers,
			Directionality:    dto.Directionality,
			FlightNumbers:     flightNumbers},
		nil
}

func ReadAgents(dtos []AgentApiDto) (AgentMap, error) {
	results := make(AgentMap)
	for _, dto := range dtos {
		if _, e := results[dto.Id]; e {
			return nil, fmt.Errorf("Duplicate agent %d", dto.Id)
		}
		agent, err := ReadAgent(dto)
		if err != nil {
			return nil, err
		}
		results[dto.Id] = agent
	}
	return results, nil
}

func ReadAgent(dto AgentApiDto) (*Agent, error) {
	return &Agent{Name: dto.Name, ImageUrl: dto.ImageUrl, Status: dto.Status, Type: dto.Type}, nil
}

/*
	Agents      []*Agent
	Age         time.Duration
	Price       float64
	DeeplinkUrl string
*/
func FindCarriers(ids []int, carriers CarrierMap) (Carriers, error) {
	results := make(Carriers, len(ids))
	for index, id := range ids {
		carrier, exists := carriers[id]
		if !exists {
			return nil, fmt.Errorf("Missing carrier %d", id)
		}
		results[index] = carrier
	}
	return results, nil
}

func FindPlaces(ids []int, places PlaceMap) (Places, error) {
	results := make(Places, len(ids))
	for index, id := range ids {
		place, exists := places[id]
		if !exists && id != 0 {
			return nil, fmt.Errorf("Missing place %d", id)
		}
		results[index] = place
	}
	return results, nil
}

func FindSegments(ids []int, segments SegmentMap) (Segments, error) {
	results := make(Segments, len(ids))
	for index, id := range ids {
		element, exists := segments[id]
		if !exists {
			return nil, fmt.Errorf("Missing segment %d", id)
		}
		results[index] = element
	}
	return results, nil
}

func FindFlightNumbers(input []FlightNumberDto, carriers CarrierMap) (FlightNumbers, error) {
	results := make(FlightNumbers, len(input))
	for index, fn := range input {
		carrier, exists := carriers[fn.CarrierId]
		if !exists {
			return nil, fmt.Errorf("Missing carrier %d", fn.CarrierId)
		}
		results[index] = FlightNumber{fn.FlightNumber, carrier}
	}
	return results, nil
}

func ReadPricingOptions(dtos []PricingOptionApiDto, agentMap AgentMap) (PricingOptions, error) {
	results := make(PricingOptions, len(dtos))
	for index, dto := range dtos {
		po, err := ReadPricingOption(dto, agentMap)
		if err != nil {
			return nil, err
		}
		results[index] = po
	}
	return results, nil
}

func ReadPricingOption(dto PricingOptionApiDto, agentMap AgentMap) (*PricingOption, error) {
	agents, err := FindAgents(dto.Agents, agentMap)
	if err != nil {
		return nil, err
	}
	// TODO: input the current time
	age := time.Duration(dto.QuoteAgeInMinutes) * time.Minute
	return &PricingOption{
			Agents:      agents,
			Age:         age,
			Price:       dto.Price,
			DeeplinkUrl: dto.DeeplinkUrl},
		nil

}

func FindAgents(ids []int, mapping AgentMap) (Agents, error) {
	results := make(Agents, len(ids))
	for index, id := range ids {
		element, exists := mapping[id]
		if !exists {
			return nil, fmt.Errorf("Missing agent %d", id)
		}
		results[index] = element
	}
	return results, nil
}

func ReadCurrencies(dtos []CurrencyDto) (Currencies, error) {
	results := make(Currencies, len(dtos))
	for index, dto := range dtos {
		currency := Currency(dto)
		results[index] = &currency
	}
	return results, nil

}

func ReadItineraries(dtos []ItineraryApiDto, legs LegMap, agentMap AgentMap) (Itineraries, error) {
	results := make(Itineraries, len(dtos))
	for index, dto := range dtos {
		element, err := ReadItinerary(dto, legs, agentMap)
		if err != nil {
			return nil, err
		}
		results[index] = element
	}
	return results, nil
}

func ReadItinerary(dto ItineraryApiDto, legs LegMap, agentMap AgentMap) (*Itinerary, error) {
	outbound, exists := legs[dto.OutboundLegId]
	if !exists {
		return nil, fmt.Errorf("Missing leg %s", dto.OutboundLegId)
	}
	inbound, exists := legs[dto.InboundLegId]
	if !exists {
		return nil, fmt.Errorf("Missing leg %s", dto.InboundLegId)
	}
	pos, err := ReadPricingOptions(dto.PricingOptions, agentMap)
	if err != nil {
		return nil, err
	}

	return &Itinerary{
			OutboundLeg:    outbound,
			InboundLeg:     inbound,
			PricingOptions: pos},
		nil
}

type CompositeFilter []ItineraryFilter

type DepartureAfterFilter struct {
	Limit     time.Duration
	Before    bool
	Departure bool
}

type DirectFilter struct {
	Direct bool
}

func (m *DirectFilter) Filter(itinerary *Itinerary) bool {
	direct := itinerary.OutboundLeg.IsDirect() && itinerary.InboundLeg.IsDirect()
	return direct == m.Direct
}

func (m *DepartureAfterFilter) Filter(itinerary *Itinerary) bool {
	var departureAt time.Time
	if m.Departure {
		departureAt = itinerary.OutboundLeg.Departure
	} else {
		departureAt = itinerary.InboundLeg.Departure
	}
	departureDuration := time.Minute*time.Duration(departureAt.Minute()) +
		time.Hour*time.Duration(departureAt.Hour())
	diff := departureDuration.Nanoseconds() - m.Limit.Nanoseconds()
	if m.Before {
		return diff <= 0
	} else {
		return diff >= 0
	}

}

func (m CompositeFilter) Filter(itinerary *Itinerary) bool {
	for _, f := range m {
		if !f.Filter(itinerary) {
			return false
		}
	}
	return true
}

func (m *Leg) IsDirect() bool {
	return len(m.Stops) == 0
}

func ApplyFilter(input Itineraries, filter ItineraryFilter) Itineraries {
	results := make(Itineraries, 0)
	for _, itinerary := range input {
		if filter.Filter(itinerary) {
			results = append(results, itinerary)
		}
	}
	return results

}

func (m *Leg) Display() string {
	return fmt.Sprintf("%s=>%s %d %s %s",
		m.Origin.Code,
		m.Destination.Code,
		len(m.Segments),
		m.Departure,
		m.Arrival)
}

func (m *Itinerary) GetPrice() float64 {
	price := math.MaxFloat64
	for _, po := range m.PricingOptions {
		price = math.Min(price, po.Price)
	}
	return price
}

func (a Itineraries) Len() int           { return len(a) }
func (a Itineraries) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a Itineraries) Less(i, j int) bool { return a[i].GetPrice() < a[j].GetPrice() }

func (m *Itinerary) Carriers() Carriers {
	carriers := make(Carriers, 0)
	carriers = append(carriers, m.OutboundLeg.Carriers...)
	carriers = append(carriers, m.InboundLeg.Carriers...)
	return carriers
}

func (m CompositeFilter) AppendDirectOnly() CompositeFilter {
	return append(m, &DirectFilter{Direct: true})
}

func (m CompositeFilter) AppendTimeFilter(time *time.Duration, before, departure bool) CompositeFilter {
	if time != nil {
		return append(m, &DepartureAfterFilter{
			Limit:     *time,
			Before:    before,
			Departure: departure})

	}
	return m
}
