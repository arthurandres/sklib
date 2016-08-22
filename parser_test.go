package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"os"
	"testing"
)

const (
	TestDataBase             = "testdata/"
	AnywhereLocationBase     = TestDataBase + "anywhere"
	AnywhereLocationXml      = AnywhereLocationBase + ".xml"
	AnywhereLocationJson     = AnywhereLocationBase + ".json"
	LivePendingLocation      = TestDataBase + "live_pending.xml"
	LiveCompleteLocation     = TestDataBase + "live_complete.xml"
	LiveCompleteJsonLocation = TestDataBase + "live_complete.json"

	quote = `
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

	leg = `
		<Leg>
        <CarrierIds>
          <int>929</int>
        </CarrierIds>
        <OriginId>65698</OriginId>
        <DestinationId>42795</DestinationId>
        <DepartureDate>2017-02-18T00:00:00</DepartureDate>
    </Leg>`

	place = `
    <PlaceDto>
      <PlaceId>837</PlaceId>
      <Name>United Arab Emirates</Name>
      <Type>Country</Type>
      <SkyscannerCode>AE</SkyscannerCode>
   </PlaceDto>`

	carrier = `
    <CarriersDto>
      <CarrierId>6</CarrierId>
      <Name>Thomson Airways</Name>
    </CarriersDto>`

	route = `
    {
      "QuoteDateTime": "2016-07-31T07:51:00",
      "Price": 568,
      "QuoteIds": [
        1,
        2
      ],
      "DestinationId": 1752,
      "OriginId": 3169460
    }`

	currency = `
   <CurrencyDto>
      <Code>GBP</Code>
      <Symbol>£</Symbol>
      <ThousandsSeparator>,</ThousandsSeparator>
      <DecimalSeparator>.</DecimalSeparator>
      <SymbolOnLeft>true</SymbolOnLeft>
      <SpaceBetweenAmountAndSymbol>false</SpaceBetweenAmountAndSymbol>
      <RoundingCoefficient>0</RoundingCoefficient>
      <DecimalDigits>2</DecimalDigits>
    </CurrencyDto>`

	placeApi = `
   <PlaceApiDto>
      <Id>13542</Id>
      <ParentId>4698</ParentId>
      <Code>LGW</Code>
      <Type>Airport</Type>
      <Name>London Gatwick</Name>
    </PlaceApiDto>`

	agentApi = `
 	 	<AgentApiDto>
      <Id>2363321</Id>
      <Name>easyJet</Name>
      <ImageUrl>http://s1.apideeplink.com/images/websites/easy.png</ImageUrl>
      <Status>UpdatesComplete</Status>
      <OptimisedForMobile>true</OptimisedForMobile>
      <BookingNumber>08431045000</BookingNumber>
      <Type>Airline</Type>
    </AgentApiDto>`

	carriersApi = `
    <CarrierApiDto>
      <Id>1050</Id>
      <Code>U2</Code>
      <Name>easyJet</Name>
				<ImageUrl>http://s1.apideeplink.com/images/airlines/EZ.png</ImageUrl>
				<DisplayCode>EZY</DisplayCode>
    </CarrierApiDto>`

	segmentApi = `
    <SegmentApiDto>
      <Id>303</Id>
      <OriginStation>12585</OriginStation>
      <DestinationStation>13554</DestinationStation>
      <DepartureDateTime>2016-11-03T18:45:00</DepartureDateTime>
      <ArrivalDateTime>2016-11-03T21:05:00</ArrivalDateTime>
      <Carrier>1755</Carrier>
      <OperatingCarrier>1755</OperatingCarrier>
      <Duration>260</Duration>
      <FlightNumber>1983</FlightNumber>
      <JourneyMode>Flight</JourneyMode>
      <Directionality>Outbound</Directionality>
    </SegmentApiDto>`

	flightNumber = `
				<FlightNumberDto>
          <FlightNumber>5351</FlightNumber>
          <CarrierId>1050</CarrierId>
        </FlightNumberDto>`

	itineraryLegApi = `
		<ItineraryLegApiDto>
      <Id>13542-1611010820-EZ-0-17517-1611011135</Id>
      <SegmentIds>
        <int>1</int>
      </SegmentIds>
      <OriginStation>13542</OriginStation>
      <DestinationStation>17517</DestinationStation>
      <Departure>2016-11-01T08:20:00</Departure>
      <Arrival>2016-11-01T11:35:00</Arrival>
      <Duration>135</Duration>
      <JourneyMode>Flight</JourneyMode>
      <Stops/>
      <Carriers>
        <int>1050</int>
      </Carriers>
      <OperatingCarriers>
        <int>1050</int>
      </OperatingCarriers>
      <Directionality>Outbound</Directionality>
      <FlightNumbers>
        <FlightNumberDto>
          <FlightNumber>5351</FlightNumber>
          <CarrierId>1050</CarrierId>
        </FlightNumberDto>
      </FlightNumbers>
    </ItineraryLegApiDto>`

	pricingOptionApi = `
        <PricingOptionApiDto>
          <Agents>
            <int>2363321</int>
          </Agents>
          <QuoteAgeInMinutes>17</QuoteAgeInMinutes>
          <Price>67.06</Price>
          <Rank xsi:nil="true"/>
          <D1 xsi:nil="true"/>
          <D2 xsi:nil="true"/>
          <D3 xsi:nil="true"/>
          <Price2 xsi:nil="true"/>
          <DeeplinkUrl>http://partners.api.skyscanner.net/apiservices/deeplink/v2?_cje=50NFWwKqwhqrumQqUleN2%2fLpO0RuaosbZ9ToDdnt0UNSSsT%2fc%2fS5%2bluEpN8FQudC&amp;url=http%3a%2f%2fwww.apideeplink.com%2ftransport_deeplink%2f4.0%2fUK%2fen-gb%2fGBP%2feasy%2f2%2f13542.17517.2016-11-01%2c17517.13542.2016-11-03%2fair%2fairli%2fflights%3fitinerary%3dflight%7c-32356%7c5351%7c13542%7c2016-11-01T08%3a20%7c17517%7c2016-11-01T11%3a35%2cflight%7c-32356%7c5354%7c17517%7c2016-11-03T20%3a45%7c13542%7c2016-11-03T22%3a15%26carriers%3d-32356%26passengers%3d1%2c0%2c0%26channel%3ddataapi%26cabin_class%3deconomy%26facilitated%3dfalse%26ticket_price%3d67.06%26is_npt%3dfalse%26is_multipart%3dfalse%26client_id%3dskyscanner_b2b%26request_id%3df2a63db4-2b9f-4578-b116-c61bee45df20%26commercial_filters%3dfalse%26q_datetime_utc%3d2016-08-16T19%3a54%3a46</DeeplinkUrl>
        </PricingOptionApiDto>`

	itineraryApi = `
    <ItineraryApiDto>
      <OutboundLegId>13542-1611010820-EZ-0-17517-1611011135</OutboundLegId>
      <InboundLegId>17517-1611032045-EZ-0-13542-1611032215</InboundLegId>
      <PricingOptionsCount xsi:nil="true"/>
      <PricingOptions>
        <PricingOptionApiDto>
          <Agents>
            <int>2363321</int>
          </Agents>
          <QuoteAgeInMinutes>17</QuoteAgeInMinutes>
          <Price>67.06</Price>
          <Rank xsi:nil="true"/>
          <D1 xsi:nil="true"/>
          <D2 xsi:nil="true"/>
          <D3 xsi:nil="true"/>
          <Price2 xsi:nil="true"/>
          <DeeplinkUrl>http://partners.api.skyscanner.net/apiservices/deeplink/v2?_cje=50NFWwKqwhqrumQqUleN2%2fLpO0RuaosbZ9ToDdnt0UNSSsT%2fc%2fS5%2bluEpN8FQudC&amp;url=http%3a%2f%2fwww.apideeplink.com%2ftransport_deeplink%2f4.0%2fUK%2fen-gb%2fGBP%2feasy%2f2%2f13542.17517.2016-11-01%2c17517.13542.2016-11-03%2fair%2fairli%2fflights%3fitinerary%3dflight%7c-32356%7c5351%7c13542%7c2016-11-01T08%3a20%7c17517%7c2016-11-01T11%3a35%2cflight%7c-32356%7c5354%7c17517%7c2016-11-03T20%3a45%7c13542%7c2016-11-03T22%3a15%26carriers%3d-32356%26passengers%3d1%2c0%2c0%26channel%3ddataapi%26cabin_class%3deconomy%26facilitated%3dfalse%26ticket_price%3d67.06%26is_npt%3dfalse%26is_multipart%3dfalse%26client_id%3dskyscanner_b2b%26request_id%3df2a63db4-2b9f-4578-b116-c61bee45df20%26commercial_filters%3dfalse%26q_datetime_utc%3d2016-08-16T19%3a54%3a46</DeeplinkUrl>
        </PricingOptionApiDto>
        <PricingOptionApiDto>
          <Agents>
            <int>3503883</int>
          </Agents>
          <QuoteAgeInMinutes>16</QuoteAgeInMinutes>
          <Price>70.12</Price>
          <Rank xsi:nil="true"/>
          <D1 xsi:nil="true"/>
          <D2 xsi:nil="true"/>
          <D3 xsi:nil="true"/>
          <Price2 xsi:nil="true"/>
          <DeeplinkUrl>http://partners.api.skyscanner.net/apiservices/deeplink/v2?_cje=50NFWwKqwhqrumQqUleN2%2fLpO0RuaosbZ9ToDdnt0UNSSsT%2fc%2fS5%2bluEpN8FQudC&amp;url=http%3a%2f%2fwww.apideeplink.com%2ftransport_deeplink%2f4.0%2fUK%2fen-gb%2fGBP%2fopuk%2f2%2f13542.17517.2016-11-01%2c17517.13542.2016-11-03%2fair%2ftrava%2fflights%3fitinerary%3dflight%7c-32356%7c5351%7c13542%7c2016-11-01T08%3a20%7c17517%7c2016-11-01T11%3a35%2cflight%7c-32356%7c5354%7c17517%7c2016-11-03T20%3a45%7c13542%7c2016-11-03T22%3a15%26carriers%3d-32356%26passengers%3d1%2c0%2c0%26channel%3ddataapi%26cabin_class%3deconomy%26facilitated%3dfalse%26ticket_price%3d70.12%26is_npt%3dfalse%26is_multipart%3dfalse%26client_id%3dskyscanner_b2b%26request_id%3df2a63db4-2b9f-4578-b116-c61bee45df20%26deeplink_ids%3deu-west-1.prod_b745ccc2ba161009323201ec985860f5%26commercial_filters%3dfalse%26q_datetime_utc%3d2016-08-16T19%3a55%3a16</DeeplinkUrl>
        </PricingOptionApiDto>
        <PricingOptionApiDto>
          <Agents>
            <int>2370315</int>
          </Agents>
          <QuoteAgeInMinutes>16</QuoteAgeInMinutes>
          <Price>76.81</Price>
          <Rank xsi:nil="true"/>
          <D1 xsi:nil="true"/>
          <D2 xsi:nil="true"/>
          <D3 xsi:nil="true"/>
          <Price2 xsi:nil="true"/>
          <DeeplinkUrl>http://partners.api.skyscanner.net/apiservices/deeplink/v2?_cje=50NFWwKqwhqrumQqUleN2%2fLpO0RuaosbZ9ToDdnt0UNSSsT%2fc%2fS5%2bluEpN8FQudC&amp;url=http%3a%2f%2fwww.apideeplink.com%2ftransport_deeplink%2f4.0%2fUK%2fen-gb%2fGBP%2feduk%2f2%2f13542.17517.2016-11-01%2c17517.13542.2016-11-03%2fair%2ftrava%2fflights%3fitinerary%3dflight%7c-32356%7c5351%7c13542%7c2016-11-01T08%3a20%7c17517%7c2016-11-01T11%3a35%2cflight%7c-32356%7c5354%7c17517%7c2016-11-03T20%3a45%7c13542%7c2016-11-03T22%3a15%26carriers%3d-32356%26passengers%3d1%2c0%2c0%26channel%3ddataapi%26cabin_class%3deconomy%26facilitated%3dfalse%26ticket_price%3d76.81%26is_npt%3dfalse%26is_multipart%3dfalse%26client_id%3dskyscanner_b2b%26request_id%3df2a63db4-2b9f-4578-b116-c61bee45df20%26deeplink_ids%3deu-west-1.prod_54a0ca47220269c994e04278719e4da5%26commercial_filters%3dfalse%26q_datetime_utc%3d2016-08-16T19%3a55%3a19</DeeplinkUrl>
        </PricingOptionApiDto>
        <PricingOptionApiDto>
          <Agents>
            <int>2043147</int>
          </Agents>
          <QuoteAgeInMinutes>16</QuoteAgeInMinutes>
          <Price>77.67</Price>
          <Rank xsi:nil="true"/>
          <D1 xsi:nil="true"/>
          <D2 xsi:nil="true"/>
          <D3 xsi:nil="true"/>
          <Price2 xsi:nil="true"/>
          <DeeplinkUrl>http://partners.api.skyscanner.net/apiservices/deeplink/v2?_cje=50NFWwKqwhqrumQqUleN2%2fLpO0RuaosbZ9ToDdnt0UNSSsT%2fc%2fS5%2bluEpN8FQudC&amp;url=http%3a%2f%2fwww.apideeplink.com%2ftransport_deeplink%2f4.0%2fUK%2fen-gb%2fGBP%2fbfuk%2f2%2f13542.17517.2016-11-01%2c17517.13542.2016-11-03%2fair%2ftrava%2fflights%3fitinerary%3dflight%7c-32356%7c5351%7c13542%7c2016-11-01T08%3a20%7c17517%7c2016-11-01T11%3a35%2cflight%7c-32356%7c5354%7c17517%7c2016-11-03T20%3a45%7c13542%7c2016-11-03T22%3a15%26carriers%3d-32356%26passengers%3d1%2c0%2c0%26channel%3ddataapi%26cabin_class%3deconomy%26facilitated%3dfalse%26ticket_price%3d77.67%26is_npt%3dfalse%26is_multipart%3dfalse%26client_id%3dskyscanner_b2b%26request_id%3df2a63db4-2b9f-4578-b116-c61bee45df20%26deeplink_ids%3deu-west-1.prod_1a9baaf707343f90adf420d03124c3e9%26commercial_filters%3dfalse%26q_datetime_utc%3d2016-08-16T19%3a55%3a04</DeeplinkUrl>
        </PricingOptionApiDto>
        <PricingOptionApiDto>
          <Agents>
            <int>2158117</int>
          </Agents>
          <QuoteAgeInMinutes>16</QuoteAgeInMinutes>
          <Price>79.06</Price>
          <Rank xsi:nil="true"/>
          <D1 xsi:nil="true"/>
          <D2 xsi:nil="true"/>
          <D3 xsi:nil="true"/>
          <Price2 xsi:nil="true"/>
          <DeeplinkUrl>http://partners.api.skyscanner.net/apiservices/deeplink/v2?_cje=50NFWwKqwhqrumQqUleN2%2fLpO0RuaosbZ9ToDdnt0UNSSsT%2fc%2fS5%2bluEpN8FQudC&amp;url=http%3a%2f%2fwww.apideeplink.com%2ftransport_deeplink%2f4.0%2fUK%2fen-gb%2fGBP%2fchpu%2f2%2f13542.17517.2016-11-01%2c17517.13542.2016-11-03%2fair%2ftrava%2fflights%3fitinerary%3dflight%7c-32356%7c5351%7c13542%7c2016-11-01T08%3a20%7c17517%7c2016-11-01T11%3a35%2cflight%7c-32356%7c5354%7c17517%7c2016-11-03T20%3a45%7c13542%7c2016-11-03T22%3a15%26carriers%3d-32356%26passengers%3d1%2c0%2c0%26channel%3ddataapi%26cabin_class%3deconomy%26facilitated%3dfalse%26ticket_price%3d79.06%26is_npt%3dfalse%26is_multipart%3dfalse%26client_id%3dskyscanner_b2b%26request_id%3df2a63db4-2b9f-4578-b116-c61bee45df20%26deeplink_ids%3deu-west-1.prod_49b00eb7997e7b9c1de79fe12afc55a8%26commercial_filters%3dfalse%26q_datetime_utc%3d2016-08-16T19%3a55%3a04</DeeplinkUrl>
        </PricingOptionApiDto>
        <PricingOptionApiDto>
          <Agents>
            <int>3165195</int>
          </Agents>
          <QuoteAgeInMinutes>16</QuoteAgeInMinutes>
          <Price>83.74</Price>
          <Rank xsi:nil="true"/>
          <D1 xsi:nil="true"/>
          <D2 xsi:nil="true"/>
          <D3 xsi:nil="true"/>
          <Price2 xsi:nil="true"/>
          <DeeplinkUrl>http://partners.api.skyscanner.net/apiservices/deeplink/v2?_cje=50NFWwKqwhqrumQqUleN2%2fLpO0RuaosbZ9ToDdnt0UNSSsT%2fc%2fS5%2bluEpN8FQudC&amp;url=http%3a%2f%2fwww.apideeplink.com%2ftransport_deeplink%2f4.0%2fUK%2fen-gb%2fGBP%2flmuk%2f2%2f13542.17517.2016-11-01%2c17517.13542.2016-11-03%2fair%2ftrava%2fflights%3fitinerary%3dflight%7c-32356%7c5351%7c13542%7c2016-11-01T08%3a20%7c17517%7c2016-11-01T11%3a35%2cflight%7c-32356%7c5354%7c17517%7c2016-11-03T20%3a45%7c13542%7c2016-11-03T22%3a15%26carriers%3d-32356%26passengers%3d1%2c0%2c0%26channel%3ddataapi%26cabin_class%3deconomy%26facilitated%3dfalse%26ticket_price%3d83.74%26is_npt%3dfalse%26is_multipart%3dfalse%26client_id%3dskyscanner_b2b%26request_id%3df2a63db4-2b9f-4578-b116-c61bee45df20%26deeplink_ids%3deu-west-1.prod_1940c74ae8702f190f598fb0cc4e0fca%26commercial_filters%3dfalse%26q_datetime_utc%3d2016-08-16T19%3a55%3a08</DeeplinkUrl>
        </PricingOptionApiDto>
      </PricingOptions>
      <BookingDetailsLink>
        <Uri>/apiservices/pricing/v1.0/7973f72c4c37493c9d4fddf626a2efc8_ecilpojl_96FF83C678D7E7C1F630A16E8F3068D6/booking</Uri>
        <Body>OutboundLegId=13542-1611010820-EZ-0-17517-1611011135&amp;InboundLegId=17517-1611032045-EZ-0-13542-1611032215</Body>
        <Method>PUT</Method>
      </BookingDetailsLink>
    </ItineraryApiDto>
    <ItineraryApiDto>
      <OutboundLegId>13542-1611010820-EZ-0-17517-1611011135</OutboundLegId>
      <InboundLegId>17517-1611031210-EZ-0-13542-1611031340</InboundLegId>
      <PricingOptionsCount xsi:nil="true"/>
      <PricingOptions>
        <PricingOptionApiDto>
          <Agents>
            <int>2363321</int>
          </Agents>
          <QuoteAgeInMinutes>17</QuoteAgeInMinutes>
          <Price>69.06</Price>
          <Rank xsi:nil="true"/>
          <D1 xsi:nil="true"/>
          <D2 xsi:nil="true"/>
          <D3 xsi:nil="true"/>
          <Price2 xsi:nil="true"/>
          <DeeplinkUrl>http://partners.api.skyscanner.net/apiservices/deeplink/v2?_cje=50NFWwKqwhqrumQqUleN2%2fLpO0RuaosbZ9ToDdnt0UNSSsT%2fc%2fS5%2bluEpN8FQudC&amp;url=http%3a%2f%2fwww.apideeplink.com%2ftransport_deeplink%2f4.0%2fUK%2fen-gb%2fGBP%2feasy%2f2%2f13542.17517.2016-11-01%2c17517.13542.2016-11-03%2fair%2fairli%2fflights%3fitinerary%3dflight%7c-32356%7c5351%7c13542%7c2016-11-01T08%3a20%7c17517%7c2016-11-01T11%3a35%2cflight%7c-32356%7c5352%7c17517%7c2016-11-03T12%3a10%7c13542%7c2016-11-03T13%3a40%26carriers%3d-32356%26passengers%3d1%2c0%2c0%26channel%3ddataapi%26cabin_class%3deconomy%26facilitated%3dfalse%26ticket_price%3d69.06%26is_npt%3dfalse%26is_multipart%3dfalse%26client_id%3dskyscanner_b2b%26request_id%3df2a63db4-2b9f-4578-b116-c61bee45df20%26commercial_filters%3dfalse%26q_datetime_utc%3d2016-08-16T19%3a54%3a46</DeeplinkUrl>
        </PricingOptionApiDto>
        <PricingOptionApiDto>
          <Agents>
            <int>3503883</int>
          </Agents>
          <QuoteAgeInMinutes>16</QuoteAgeInMinutes>
          <Price>72.12</Price>
          <Rank xsi:nil="true"/>
          <D1 xsi:nil="true"/>
          <D2 xsi:nil="true"/>
          <D3 xsi:nil="true"/>
          <Price2 xsi:nil="true"/>
          <DeeplinkUrl>http://partners.api.skyscanner.net/apiservices/deeplink/v2?_cje=50NFWwKqwhqrumQqUleN2%2fLpO0RuaosbZ9ToDdnt0UNSSsT%2fc%2fS5%2bluEpN8FQudC&amp;url=http%3a%2f%2fwww.apideeplink.com%2ftransport_deeplink%2f4.0%2fUK%2fen-gb%2fGBP%2fopuk%2f2%2f13542.17517.2016-11-01%2c17517.13542.2016-11-03%2fair%2ftrava%2fflights%3fitinerary%3dflight%7c-32356%7c5351%7c13542%7c2016-11-01T08%3a20%7c17517%7c2016-11-01T11%3a35%2cflight%7c-32356%7c5352%7c17517%7c2016-11-03T12%3a10%7c13542%7c2016-11-03T13%3a40%26carriers%3d-32356%26passengers%3d1%2c0%2c0%26channel%3ddataapi%26cabin_class%3deconomy%26facilitated%3dfalse%26ticket_price%3d72.12%26is_npt%3dfalse%26is_multipart%3dfalse%26client_id%3dskyscanner_b2b%26request_id%3df2a63db4-2b9f-4578-b116-c61bee45df20%26deeplink_ids%3deu-west-1.prod_880d9f42cc449944575697d761a1228b%26commercial_filters%3dfalse%26q_datetime_utc%3d2016-08-16T19%3a55%3a16</DeeplinkUrl>
        </PricingOptionApiDto>
        <PricingOptionApiDto>
          <Agents>
            <int>2370315</int>
          </Agents>
          <QuoteAgeInMinutes>16</QuoteAgeInMinutes>
          <Price>78.87</Price>
          <Rank xsi:nil="true"/>
          <D1 xsi:nil="true"/>
          <D2 xsi:nil="true"/>
          <D3 xsi:nil="true"/>
          <Price2 xsi:nil="true"/>
          <DeeplinkUrl>http://partners.api.skyscanner.net/apiservices/deeplink/v2?_cje=50NFWwKqwhqrumQqUleN2%2fLpO0RuaosbZ9ToDdnt0UNSSsT%2fc%2fS5%2bluEpN8FQudC&amp;url=http%3a%2f%2fwww.apideeplink.com%2ftransport_deeplink%2f4.0%2fUK%2fen-gb%2fGBP%2feduk%2f2%2f13542.17517.2016-11-01%2c17517.13542.2016-11-03%2fair%2ftrava%2fflights%3fitinerary%3dflight%7c-32356%7c5351%7c13542%7c2016-11-01T08%3a20%7c17517%7c2016-11-01T11%3a35%2cflight%7c-32356%7c5352%7c17517%7c2016-11-03T12%3a10%7c13542%7c2016-11-03T13%3a40%26carriers%3d-32356%26passengers%3d1%2c0%2c0%26channel%3ddataapi%26cabin_class%3deconomy%26facilitated%3dfalse%26ticket_price%3d78.87%26is_npt%3dfalse%26is_multipart%3dfalse%26client_id%3dskyscanner_b2b%26request_id%3df2a63db4-2b9f-4578-b116-c61bee45df20%26deeplink_ids%3deu-west-1.prod_3dc8e47eaea18b493224b8d5b8e0b79a%26commercial_filters%3dfalse%26q_datetime_utc%3d2016-08-16T19%3a55%3a19</DeeplinkUrl>
        </PricingOptionApiDto>
        <PricingOptionApiDto>
          <Agents>
            <int>2043147</int>
          </Agents>
          <QuoteAgeInMinutes>16</QuoteAgeInMinutes>
          <Price>79.66</Price>
          <Rank xsi:nil="true"/>
          <D1 xsi:nil="true"/>
          <D2 xsi:nil="true"/>
          <D3 xsi:nil="true"/>
          <Price2 xsi:nil="true"/>
          <DeeplinkUrl>http://partners.api.skyscanner.net/apiservices/deeplink/v2?_cje=50NFWwKqwhqrumQqUleN2%2fLpO0RuaosbZ9ToDdnt0UNSSsT%2fc%2fS5%2bluEpN8FQudC&amp;url=http%3a%2f%2fwww.apideeplink.com%2ftransport_deeplink%2f4.0%2fUK%2fen-gb%2fGBP%2fbfuk%2f2%2f13542.17517.2016-11-01%2c17517.13542.2016-11-03%2fair%2ftrava%2fflights%3fitinerary%3dflight%7c-32356%7c5351%7c13542%7c2016-11-01T08%3a20%7c17517%7c2016-11-01T11%3a35%2cflight%7c-32356%7c5352%7c17517%7c2016-11-03T12%3a10%7c13542%7c2016-11-03T13%3a40%26carriers%3d-32356%26passengers%3d1%2c0%2c0%26channel%3ddataapi%26cabin_class%3deconomy%26facilitated%3dfalse%26ticket_price%3d79.66%26is_npt%3dfalse%26is_multipart%3dfalse%26client_id%3dskyscanner_b2b%26request_id%3df2a63db4-2b9f-4578-b116-c61bee45df20%26deeplink_ids%3deu-west-1.prod_844546321b0cdc066df42aae040bab07%26commercial_filters%3dfalse%26q_datetime_utc%3d2016-08-16T19%3a55%3a04</DeeplinkUrl>
        </PricingOptionApiDto>
        <PricingOptionApiDto>
          <Agents>
            <int>2158117</int>
          </Agents>
          <QuoteAgeInMinutes>16</QuoteAgeInMinutes>
          <Price>81.06</Price>
          <Rank xsi:nil="true"/>
          <D1 xsi:nil="true"/>
          <D2 xsi:nil="true"/>
          <D3 xsi:nil="true"/>
          <Price2 xsi:nil="true"/>
          <DeeplinkUrl>http://partners.api.skyscanner.net/apiservices/deeplink/v2?_cje=50NFWwKqwhqrumQqUleN2%2fLpO0RuaosbZ9ToDdnt0UNSSsT%2fc%2fS5%2bluEpN8FQudC&amp;url=http%3a%2f%2fwww.apideeplink.com%2ftransport_deeplink%2f4.0%2fUK%2fen-gb%2fGBP%2fchpu%2f2%2f13542.17517.2016-11-01%2c17517.13542.2016-11-03%2fair%2ftrava%2fflights%3fitinerary%3dflight%7c-32356%7c5351%7c13542%7c2016-11-01T08%3a20%7c17517%7c2016-11-01T11%3a35%2cflight%7c-32356%7c5352%7c17517%7c2016-11-03T12%3a10%7c13542%7c2016-11-03T13%3a40%26carriers%3d-32356%26passengers%3d1%2c0%2c0%26channel%3ddataapi%26cabin_class%3deconomy%26facilitated%3dfalse%26ticket_price%3d81.06%26is_npt%3dfalse%26is_multipart%3dfalse%26client_id%3dskyscanner_b2b%26request_id%3df2a63db4-2b9f-4578-b116-c61bee45df20%26deeplink_ids%3deu-west-1.prod_39246678904cbc566ca30da73bd67e42%26commercial_filters%3dfalse%26q_datetime_utc%3d2016-08-16T19%3a55%3a04</DeeplinkUrl>
        </PricingOptionApiDto>
        <PricingOptionApiDto>
          <Agents>
            <int>3165195</int>
          </Agents>
          <QuoteAgeInMinutes>16</QuoteAgeInMinutes>
          <Price>85.73</Price>
          <Rank xsi:nil="true"/>
          <D1 xsi:nil="true"/>
          <D2 xsi:nil="true"/>
          <D3 xsi:nil="true"/>
          <Price2 xsi:nil="true"/>
          <DeeplinkUrl>http://partners.api.skyscanner.net/apiservices/deeplink/v2?_cje=50NFWwKqwhqrumQqUleN2%2fLpO0RuaosbZ9ToDdnt0UNSSsT%2fc%2fS5%2bluEpN8FQudC&amp;url=http%3a%2f%2fwww.apideeplink.com%2ftransport_deeplink%2f4.0%2fUK%2fen-gb%2fGBP%2flmuk%2f2%2f13542.17517.2016-11-01%2c17517.13542.2016-11-03%2fair%2ftrava%2fflights%3fitinerary%3dflight%7c-32356%7c5351%7c13542%7c2016-11-01T08%3a20%7c17517%7c2016-11-01T11%3a35%2cflight%7c-32356%7c5352%7c17517%7c2016-11-03T12%3a10%7c13542%7c2016-11-03T13%3a40%26carriers%3d-32356%26passengers%3d1%2c0%2c0%26channel%3ddataapi%26cabin_class%3deconomy%26facilitated%3dfalse%26ticket_price%3d85.73%26is_npt%3dfalse%26is_multipart%3dfalse%26client_id%3dskyscanner_b2b%26request_id%3df2a63db4-2b9f-4578-b116-c61bee45df20%26deeplink_ids%3deu-west-1.prod_4983575e609761f9ad1fcca8f001db55%26commercial_filters%3dfalse%26q_datetime_utc%3d2016-08-16T19%3a55%3a08</DeeplinkUrl>
        </PricingOptionApiDto>
      </PricingOptions>
      <BookingDetailsLink>
        <Uri>/apiservices/pricing/v1.0/7973f72c4c37493c9d4fddf626a2efc8_ecilpojl_96FF83C678D7E7C1F630A16E8F3068D6/booking</Uri>
        <Body>OutboundLegId=13542-1611010820-EZ-0-17517-1611011135&amp;InboundLegId=17517-1611031210-EZ-0-13542-1611031340</Body>
        <Method>PUT</Method>
      </BookingDetailsLink>
    </ItineraryApiDto>
`

	bookingdetailsLink = `
			<BookingDetailsLink>
        <Uri>/apiservices/pricing/v1.0/7973f72c4c37493c9d4fddf626a2efc8_ecilpojl_96FF83C678D7E7C1F630A16E8F3068D6/booking</Uri>
        <Body>OutboundLegId=13542-1611010820-EZ-0-17517-1611011135&amp;InboundLegId=17517-1611032045-EZ-0-13542-1611032215</Body>
        <Method>PUT</Method>
      </BookingDetailsLink>`

	liveQuery = `
  <Query>
    <Country>GB</Country>
    <Currency>GBP</Currency>
    <Locale>en-gb</Locale>
    <Adults>1</Adults>
    <Children>0</Children>
    <Infants>0</Infants>
    <OriginPlace>4698</OriginPlace>
    <DestinationPlace>8222</DestinationPlace>
    <OutboundDate>2016-11-01</OutboundDate>
    <InboundDate>2016-11-03</InboundDate>
    <LocationSchema>Default</LocationSchema>
    <CabinClass>Economy</CabinClass>
    <GroupPricing>false</GroupPricing>
  </Query>`
)

func TestParseLiveQueryDto(t *testing.T) {
	var e LiveQueryDto
	if err := xml.Unmarshal([]byte(liveQuery), &e); err != nil {
		panic(err)
	}

	assert.Equal(t, "GB", e.Country)
	assert.Equal(t, "GBP", e.Currency)
	assert.Equal(t, "en-gb", e.Locale)
	assert.Equal(t, 1, e.Adults)
	assert.Equal(t, 0, e.Children)
	assert.Equal(t, 0, e.Infants)
	assert.Equal(t, "4698", string(e.OriginPlace))
	assert.Equal(t, "8222", string(e.DestinationPlace))
	assert.Equal(t, "2016-11-01", e.OutboundDate)
	assert.Equal(t, "2016-11-03", e.InboundDate)
	assert.Equal(t, "Default", e.LocationSchema)
	assert.Equal(t, "Economy", e.CabinClass)
	assert.Equal(t, false, e.GroupPricing)
}

func TestParseItineraryApiDto(t *testing.T) {
	var e ItineraryApiDto
	if err := xml.Unmarshal([]byte(itineraryApi), &e); err != nil {
		panic(err)
	}

	assert.Equal(t, "13542-1611010820-EZ-0-17517-1611011135", e.OutboundLegId)
	assert.Equal(t, "17517-1611032045-EZ-0-13542-1611032215", e.InboundLegId)
	assert.Equal(t, 6, len(e.PricingOptions))
	assert.Equal(t, 2363321, e.PricingOptions[0].Agents[0])
	assert.Equal(t, "PUT", e.BookingDetailsLink.Method)
}

func TestParseBookingDetailsLinkDto(t *testing.T) {
	var e BookingDetailsLinkDto
	if err := xml.Unmarshal([]byte(bookingdetailsLink), &e); err != nil {
		panic(err)
	}

	assert.Equal(t, 108, len(e.Uri))
	assert.Equal(t, 104, len(e.Body))
	assert.Equal(t, "PUT", e.Method)

}
func TestParsePricingOptionApi(t *testing.T) {
	var e PricingOptionApiDto
	if err := xml.Unmarshal([]byte(pricingOptionApi), &e); err != nil {
		panic(err)
	}

	assert.Equal(t, 1, len(e.Agents))
	assert.Equal(t, 2363321, e.Agents[0])
	assert.Equal(t, 17, e.QuoteAgeInMinutes)
	assert.Equal(t, 67.06, e.Price)
	assert.Equal(t, 809, len(e.DeeplinkUrl))
}

func TestParseItineraryLegApidto(t *testing.T) {
	var e ItineraryLegApiDto
	if err := xml.Unmarshal([]byte(itineraryLegApi), &e); err != nil {
		panic(err)
	}

	assert.Equal(t, "13542-1611010820-EZ-0-17517-1611011135", e.Id)
	assert.Equal(t, 1, len(e.SegmentIds))
	assert.Equal(t, 1, e.SegmentIds[0])
	assert.Equal(t, 13542, e.OriginStation)
	assert.Equal(t, 17517, e.DestinationStation)
	assert.Equal(t, "2016-11-01T08:20:00", e.Departure)
	assert.Equal(t, "2016-11-01T11:35:00", e.Arrival)
	assert.Equal(t, 135, e.Duration)
	assert.Equal(t, "Flight", e.JourneyMode)
	assert.Equal(t, 0, len(e.Stops))
	assert.Equal(t, 1, len(e.Carriers))
	assert.Equal(t, 1050, e.Carriers[0])
	assert.Equal(t, 1, len(e.OperatingCarriers))
	assert.Equal(t, "Outbound", e.Directionality)
	assert.Equal(t, 1, len(e.FlightNumbers))
	assert.Equal(t, "5351", e.FlightNumbers[0].FlightNumber)
	assert.Equal(t, 1050, e.FlightNumbers[0].CarrierId)
}

func TestParseFlightNumberDto(t *testing.T) {
	var e FlightNumberDto
	if err := xml.Unmarshal([]byte(flightNumber), &e); err != nil {
		panic(err)
	}

	assert.Equal(t, "5351", e.FlightNumber)
	assert.Equal(t, 1050, e.CarrierId)
}

func TestParseSegmentApiDto(t *testing.T) {
	var e SegmentApiDto
	if err := xml.Unmarshal([]byte(segmentApi), &e); err != nil {
		panic(err)
	}

	assert.Equal(t, 303, e.Id)
	assert.Equal(t, 12585, e.OriginStation)
	assert.Equal(t, 13554, e.DestinationStation)
	assert.Equal(t, "2016-11-03T18:45:00", e.DepartureDateTime)
	assert.Equal(t, "2016-11-03T21:05:00", e.ArrivalDateTime)
	assert.Equal(t, 1755, e.Carrier)
	assert.Equal(t, 1755, e.OperatingCarrier)
	assert.Equal(t, 260, e.Duration)
	assert.Equal(t, "1983", e.FlightNumber)
	assert.Equal(t, "Flight", e.JourneyMode)
	assert.Equal(t, "Outbound", e.Directionality)
}
func TestParseCarrierApiDto(t *testing.T) {
	var e CarrierApiDto
	if err := xml.Unmarshal([]byte(carriersApi), &e); err != nil {
		panic(err)
	}

	assert.Equal(t, 1050, e.Id)
	assert.Equal(t, "U2", e.Code)
	assert.Equal(t, "easyJet", e.Name)
	assert.Equal(t, "http://s1.apideeplink.com/images/airlines/EZ.png", e.ImageUrl)
	assert.Equal(t, "EZY", e.DisplayCode)
}

func TestParseAgentApiDto(t *testing.T) {
	var e AgentApiDto
	if err := xml.Unmarshal([]byte(agentApi), &e); err != nil {
		panic(err)
	}

	assert.Equal(t, 2363321, e.Id)
	assert.Equal(t, "easyJet", e.Name)
	assert.Equal(t, "http://s1.apideeplink.com/images/websites/easy.png", e.ImageUrl)
	assert.Equal(t, "UpdatesComplete", e.Status)
	assert.Equal(t, true, e.OptimisedForMobile)
	assert.Equal(t, "08431045000", e.BookingNumber)
	assert.Equal(t, "Airline", e.Type)

}

func TestParseQuoteDto(t *testing.T) {
	var e QuoteDto
	if err := xml.Unmarshal([]byte(quote), &e); err != nil {
		panic(err)
	}

	assert.Equal(t, 1, e.QuoteId)
	assert.Equal(t, 326.0, e.MinPrice)
	assert.Equal(t, false, e.Direct)
	assert.Equal(t, 65698, e.OutboundLeg.OriginId)
	assert.Equal(t, 65698, e.InboundLeg.DestinationId)
	assert.Equal(t, "2016-07-20T16:54:00", e.QuoteDateTime)

}

func TestParsePlaceApiDto(t *testing.T) {
	var e PlaceApiDto
	if err := xml.Unmarshal([]byte(placeApi), &e); err != nil {
		panic(err)
	}

	assert.Equal(t, 13542, e.Id)
	assert.Equal(t, "4698", string(e.ParentId))
	assert.Equal(t, "LGW", e.Code)
	assert.Equal(t, "Airport", e.Type)
	assert.Equal(t, "London Gatwick", e.Name)
}

func TestParseCurrencyDto(t *testing.T) {
	var e CurrencyDto
	if err := xml.Unmarshal([]byte(currency), &e); err != nil {
		panic(err)
	}

	assert.Equal(t, "GBP", e.Code)
	//assert.Equal(t, "£", e.Symbol)
	assert.Equal(t, ",", e.ThousandsSeparator)
	assert.Equal(t, ".", e.DecimalSeparator)
	assert.Equal(t, true, e.SymbolOnLeft)
	assert.Equal(t, false, e.SpaceBetweenAmountAndSymbol)
	assert.Equal(t, 0, e.RoundingCoefficient)
	assert.Equal(t, 2, e.DecimalDigits)

}

func TestParseLeg(t *testing.T) {
	var e Leg
	xml.Unmarshal([]byte(leg), &e)
	assert.Equal(t, 65698, e.OriginId, "diff")

}
func TestPlace(t *testing.T) {
	var e PlaceDto
	if err := xml.Unmarshal([]byte(place), &e); err != nil {
		panic(err)
	}

	assert.Equal(t, 837, e.PlaceId)
	assert.Equal(t, "United Arab Emirates", e.Name)
	assert.Equal(t, "Country", e.Type)
	assert.Equal(t, "AE", e.SkyscannerCode)

}

func TestCarrier(t *testing.T) {
	var e CarriersDto
	if err := xml.Unmarshal([]byte(carrier), &e); err != nil {
		panic(err)
	}

	assert.Equal(t, 6, e.CarrierId)
	assert.Equal(t, "Thomson Airways", e.Name)
}

func TestAnywhere(t *testing.T) {

	var anywhere BrowseRoutesReply
	ParseFromXmlFile(AnywhereLocationXml, &anywhere)
	anywhere.PrintStats()

	assert.Equal(t, 1, len(anywhere.Currencies))
	assert.Equal(t, 312, len(anywhere.Quotes))
	assert.Equal(t, 222, len(anywhere.Routes))
	assert.Equal(t, 478, len(anywhere.Places))
	assert.Equal(t, 73, len(anywhere.Carriers))

}

func OpenOrPanic(fileName string) *os.File {
	file, err := os.Open(AnywhereLocationJson)
	if err != nil {
		panic(err)
	}
	return file
}

func ReadOrPanic(fileName string) []byte {

	file := OpenOrPanic(fileName)
	defer file.Close()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}
	return data

}

func GetFileReader(fileName string) io.Reader {
	return bytes.NewReader(ReadOrPanic(fileName))
}

func TestAnywhereJson(t *testing.T) {
	decoder := json.NewDecoder(GetFileReader(AnywhereLocationJson))
	var anywhere BrowseRoutesReply
	err := decoder.Decode(&anywhere)
	if err != nil {
		panic(err)
	}

	anywhere.PrintStats()

	assert.Equal(t, 1, len(anywhere.Currencies))
	assert.Equal(t, 182, len(anywhere.Quotes))
	assert.Equal(t, 199, len(anywhere.Routes))
	assert.Equal(t, 350, len(anywhere.Places))
	assert.Equal(t, 65, len(anywhere.Carriers))

	anywhere.GetPlacesByPrice2()

}
func TestRouteJson(t *testing.T) {
	var e RouteDto
	if err := json.Unmarshal([]byte(route), &e); err != nil {
		panic(err)
	}

	assert.Equal(t, 2, len(e.QuoteIds))

}

func TestFormat(t *testing.T) {

	assert.Equal(t, "2016-08-16", FormatDate("20160816"))
}

func TestLiveComplete(t *testing.T) {
	var reply LiveReply
	ParseFromXmlFile(LiveCompleteLocation, &reply)

	assert.Equal(t, UpdatesCompleteStatus, reply.Status)
	var expected = map[string]int{
		"Currencies": 10, "Places": 82, "Legs": 194, "Segments": 303, "Carriers": 33, "Agents": 33,
		"Itineraries": 631}
	assert.Equal(t, expected, reply.Stats())

}

func ParseFromXmlFile(fileName string, data interface{}) {
	file, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	decoder := xml.NewDecoder(file)
	err = decoder.Decode(data)
	if err != nil {
		panic(err)
	}
}

func ParseFromJsonFile(fileName string, data interface{}) {
	file, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(data)
	if err != nil {
		panic(err)
	}
}

func TestLiveCompleteJson(t *testing.T) {

	var reply LiveReply
	ParseFromJsonFile(LiveCompleteJsonLocation, &reply)
	assert.Equal(t, UpdatesCompleteStatus, reply.Status)
	var expected = map[string]int{
		"Currencies": 10, "Itineraries": 699, "Places": 82, "Legs": 201, "Segments": 311, "Carriers": 33, "Agents": 33}
	assert.Equal(t, expected, reply.Stats())

}
