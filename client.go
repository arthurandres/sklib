package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

type QueryEngine interface {
	Get(url string) (payload []byte, err error)
}

type LiveEngine struct {
	Key string
}

func (m LiveEngine) Get(url string) ([]byte, error) {

	fullUrl := url + "?apiKey=" + m.Key
	resp, err := http.Get(fullUrl)
	fmt.Println(fullUrl)
	if err != nil {
		return nil, err
	}
	fmt.Println(resp)
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)

}

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

const anywhereQueryFormat = "http://partners.api.skyscanner.net/apiservices/browseroutes/v1.0/%s/%s/%s/%s/%s/%s/%s"
const anywhereQueryExample = "http://partners.api.skyscanner.net/apiservices/browseroutes/v1.0/GB/GBP/en-GB/LON/anywhere/20160819/20160821"

func runQuery(engine QueryEngine) {
	const url = anywhereQueryExample
	data, err := engine.Get(url)
	fmt.Println(string(data))
	if err != nil {
		panic(err)
	} else {
		results := ParseAnywhereQuery(data)
		results.PrintStats()
	}

}

func main() {

	engine := LiveEngine{Key: "ar926739961631567929873917891697"}
	runQuery(engine)
}
