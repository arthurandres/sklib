package main

import (
	"fmt"
	"net/http"
)

type Local struct {
	Country  string
	Currency string
	Language string
}

type AnywhereRequest struct {
	Local         Local
	DepartureFrom string
	DepartureDate string
	ReturnDate    string
}

func (m AnywhereRequest) Query() string {
	return fmt.Sprintf(
		anywhereQueryFormat,
		m.Local.Country,
		m.Local.Currency,
		m.Local.Language,
		m.DepartureFrom,
		m.DepartureDate,
		m.ReturnDate)
}

const key = "TBD"
const anywhereQueryFormat = "http://partners.api.skyscanner.net/apiservices/browseroutes/v1.0/%s/%s/%s/%s/%s/%s/%s?apiKey=" + key
const anywhereQueryExample = "http://partners.api.skyscanner.net/apiservices/browseroutes/v1.0/GB/GBP/en-GB/LON/anywhere/20160819/20160821?apiKey=" + key

func runQuery() {
	const url = anywhereQueryExample
	fmt.Printf(url)

	resp, err := http.Get(anywhereQueryExample)

	fmt.Println(resp)
	fmt.Println(err)
}

func main() {
	var request AnywhereRequest

	runQuery()
}
