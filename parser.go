package sklib

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
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

func ParseBrowseRoutesReplyJson(data []byte) *BrowseRoutesReply {
	anywhere := &BrowseRoutesReply{}
	err := ParseJson(data, anywhere)
	if err != nil {
		panic(err)
	}
	return anywhere
}

func FormatDateToForm(input string) string {
	date, err := ParseUrlDate(input)
	if err != nil {
		panic(err)
	}
	return date.Format(DateFormatForm)
}

func ParseUrlDate(input string) (time.Time, error) {
	return time.Parse(DateFormatUrl, input)
}

func ParseUrlDateOP(input string) time.Time {
	r, e := ParseUrlDate(input)
	if e != nil {
		panic(e)
	}
	return r

}

func ParseXml(data []byte, output interface{}) error {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	return decoder.Decode(output)
}

func ParseJson(data []byte, output interface{}) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	return decoder.Decode(output)
}
