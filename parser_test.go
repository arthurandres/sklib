package main

import (
	"encoding/xml"
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

const quote = `
		<QuoteDto>
      <QuoteId>1</QuoteId>
      <MinPrice>326</MinPrice>
      <Direct>false</Direct>
      <OutboundLeg>
        <CarrierIds>
          <int>929</int>
        </CarrierIds>
        <OriginId>65698</OriginId>
        <DestinationId>42795</DestinationId>
        <DepartureDate>2017-02-18T00:00:00</DepartureDate>
      </OutboundLeg>
      <InboundLeg>
        <CarrierIds>
          <int>929</int>
        </CarrierIds>
        <OriginId>42795</OriginId>
        <DestinationId>65698</DestinationId>
        <DepartureDate>2017-02-27T00:00:00</DepartureDate>
      </InboundLeg>
      <QuoteDateTime>2016-07-20T16:54:00</QuoteDateTime>
		</QuoteDto>`

func TestParseQuoteDto(t *testing.T) {
	var e QuoteDto
	if err := xml.Unmarshal([]byte(quote), &e); err != nil {
		panic(err)
	}

	assert.Equal(t, 1, e.QuoteId)
	assert.Equal(t, 326, e.MinPrice)
	assert.Equal(t, false, e.Direct)
	assert.Equal(t, 65698, e.OutboundLeg.OriginId)
	assert.Equal(t, 65698, e.InboundLeg.DestinationId)
	assert.Equal(t, "2016-07-20T16:54:00", e.QuoteDateTime)

	fmt.Println(e)

}

const leg = `
		<Leg>
        <CarrierIds>
          <int>929</int>
        </CarrierIds>
        <OriginId>65698</OriginId>
        <DestinationId>42795</DestinationId>
        <DepartureDate>2017-02-18T00:00:00</DepartureDate>
    </Leg>`

func TestParseLeg(t *testing.T) {
	var e Leg
	xml.Unmarshal([]byte(leg), &e)
	fmt.Println(e)
	assert.Equal(t, 65698, e.OriginId, "diff")

}

const place = `
    <PlaceDto>
      <PlaceId>837</PlaceId>
      <Name>United Arab Emirates</Name>
      <Type>Country</Type>
      <SkyscannerCode>AE</SkyscannerCode>
    </PlaceDto>`

func TestPlace(t *testing.T) {
	var e PlaceDto
	if err := xml.Unmarshal([]byte(place), &e); err != nil {
		panic(err)
	}
	fmt.Println(e)

	assert.Equal(t, 837, e.PlaceId)
	assert.Equal(t, "United Arab Emirates", e.Name)
	assert.Equal(t, "Country", e.Type)
	assert.Equal(t, "AE", e.SkyscannerCode)

}

const carrier = `
    <CarriersDto>
      <CarrierId>6</CarrierId>
      <Name>Thomson Airways</Name>
    </CarriersDto>`

func TestCarrier(t *testing.T) {
	var e CarriersDto
	if err := xml.Unmarshal([]byte(carrier), &e); err != nil {
		panic(err)
	}
	fmt.Println(e)

	assert.Equal(t, 6, e.CarrierId)
	assert.Equal(t, "Thomson Airways", e.Name)

}

const anywhereLocation = "testdata/anywhere.xml"

func TestAnywhere(t *testing.T) {

	data, err := os.Open(anywhereLocation)

	if err == nil {
		defer data.Close()
		decoder := xml.NewDecoder(data)
		var anywhere AnywhereQuery
		decoder.Decode(&anywhere)
		fmt.Println(anywhere)
		anywhere.PrintStats()

		assert.Equal(t, 1, len(anywhere.Currencies))
		assert.Equal(t, 312, len(anywhere.Quotes))
		assert.Equal(t, 222, len(anywhere.Routes))
		assert.Equal(t, 478, len(anywhere.Places))
		assert.Equal(t, 73, len(anywhere.Carriers))

	} else {
		panic(err)
	}

}
