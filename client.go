package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/boltdb/bolt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

const (
	browseRouteFormat  = "http://partners.api.skyscanner.net/apiservices/browseroutes/v1.0/%s/%s/%s/%s/%s"
	browseRouteExample = "http://partners.api.skyscanner.net/apiservices/browseroutes/v1.0/GB/GBP/en-GB/LON/anywhere/20160819/20160821"
	anywhere           = "anywhere"
	linkBase           = "https://www.skyscanner.net/transport/flights/%s/%s/%s/%s/"
	LocationKey        = "Location"
	ApiKeyTag          = "?apiKey="
)

type RequestEngine interface {
	Get(url string) (payload []byte, err error)
}

type LiveEngine struct {
	Key string
}

func FormatKey(key string) string {
	return ApiKeyTag + key
}

func (m *LiveEngine) Get(url string) ([]byte, error) {

	fullUrl := url + FormatKey(m.Key)
	resp, err := http.Get(fullUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)

}

type CacheStore interface {
	Get(key string) (data []byte)
	Set(key string, data []byte) error
}

type BoltStore struct {
	DB     *bolt.DB
	Bucket string
}

type CachedEngine struct {
	Engine    RequestEngine
	Cache     CacheStore
	WriteOnly bool
}

type SlowEngine struct {
	Engine RequestEngine
	Delay  time.Duration
}

func (m *SlowEngine) Get(url string) ([]byte, error) {
	time.Sleep(m.Delay)
	return m.Engine.Get(url)
}

func (m *CachedEngine) Get(url string) ([]byte, error) {

	var cache []byte
	if !m.WriteOnly {
		cache = m.Cache.Get(url)
	}
	if cache != nil {
		return cache, nil
	}
	payload, err := m.Engine.Get(url)
	if err != nil {
		return nil, err
	}
	err = m.Cache.Set(url, payload)
	if err != nil {
		panic(err)
	}

	return payload, err

}

type Local struct {
	Country  string
	Currency string
	Language string
}

func (m Local) SubUrl() string {
	return fmt.Sprintf("%s/%s/%s",
		m.Country,
		m.Currency,
		m.Language)
}

type BrowseRoutesRequest struct {
	Local         Local
	From          string
	To            string
	DepartureDate string
	ReturnDate    string
}

func CreateBrowseRouteRequest(local Local, from string, departureDate string, returnDate string) BrowseRoutesRequest {
	return BrowseRoutesRequest{local, from, anywhere, departureDate, returnDate}
}

func (m BrowseRoutesRequest) Url() string {
	return fmt.Sprintf(
		browseRouteFormat,
		m.Local.SubUrl(),
		m.From,
		m.To,
		m.DepartureDate,
		m.ReturnDate)
}

func Anywhere(m RequestEngine, r BrowseRoutesRequest) error {
	url := r.Url()
	payload, err := m.Get(url)
	if err != nil {
		return err
	}
	ParseBrowseRoutesReplyJson(payload)
	return nil
}

func GetLink(from string, to string, departureDate string, returnDate string) string {
	return fmt.Sprintf(linkBase, from, to, departureDate, returnDate)
}

func runRequest(engine RequestEngine, r BrowseRoutesRequest) (*BrowseRoutesReply, error) {
	url := r.Url()
	data, err := engine.Get(url)
	if err != nil {
		return nil, err
	} else {
		results := ParseBrowseRoutesReplyJson(data)
		return results, nil
	}

}

func (m *BoltStore) Get(key string) []byte {
	var value []byte
	m.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(m.Bucket))
		if b == nil {
			return nil
		}
		value = b.Get([]byte(key))
		return nil
	})
	return value
}

func (m *BoltStore) Set(key string, data []byte) error {
	var err error
	m.DB.Update(func(tx *bolt.Tx) error {
		var b *bolt.Bucket
		b, err = tx.CreateBucketIfNotExists([]byte(m.Bucket))
		if err != nil {
			return err
		}
		err = b.Put([]byte(key), data)
		return err
	})
	return err
}

func Test(cs CacheStore) {
	cs.Get("hello")
	cs.Get("hello2")
	cs.Set("hello", []byte("world"))
	cs.Get("hello")

}

func CreateCache() *BoltStore {
	var db *bolt.DB
	var err error
	db, err = bolt.Open("cache.db", 0600, nil)
	if err != nil {
		panic(err)
	}
	return &BoltStore{db, "cache"}
}

func GetLocalExample() Local {
	return Local{"GB", "GBP", "en-GB"}
}
func GetBrowseRoutesRequestExample() BrowseRoutesRequest {
	return CreateBrowseRouteRequest(GetLocalExample(), "LON", "20161101", "20161103")
}

func insertAll(from map[string]float64, to map[string]float64) {
	for k, v := range from {
		to[k] = v
	}
}

type LiveRequest struct {
	ApiKey        string
	Local         Local
	Origin        string
	Destination   string
	DepartureDate string
	ReturnDate    string
}

func (m *LiveRequest) Values() url.Values {
	return url.Values{
		"apiKey":           {m.ApiKey},
		"country":          {m.Local.Country},
		"currency":         {m.Local.Currency},
		"locale":           {m.Local.Language},
		"originplace":      {m.Origin},
		"destinationplace": {m.Destination},
		"outbounddate":     {FormatDate(m.DepartureDate)},
		"inbounddate":      {FormatDate(m.ReturnDate)},
		"locationschema":   {"Iata"}}
}

func CreateLiveRequestExample(key string, origin string, destination string, departureDate string, returnDate string) LiveRequest {

	return LiveRequest{
		key,
		GetLocalExample(),
		origin,
		destination,
		departureDate,
		returnDate}
}

func Poll(url string) {
	for {
		resp, err := http.Get(url)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		decoder := json.NewDecoder(bytes.NewReader(data))
		fmt.Println(string(data))
		var reply LiveReply
		err = decoder.Decode(&reply)
		if err != nil {
			panic(err)
		}
		fmt.Println(reply.Status)
		if reply.Status == UpdatesCompleteStatus {
			break
		}
		time.Sleep(1 * time.Second)
	}

}

func PostAndPoll(request LiveRequest) {
	url := "http://partners.api.skyscanner.net/apiservices/pricing/v1.0"

	resp, err := http.PostForm(url, request.Values())
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	location := resp.Header.Get(LocationKey)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))

	fmt.Println(location)
	fullUrl := location + FormatKey(request.ApiKey)
	Poll(fullUrl)
}

var (
	keyFile       = flag.String("keyFile", "key", "API key provided by skyscanner")
	origin        = flag.String("from", "LON", "Origin Town/Airport")
	departureDate = flag.String("out", "20161101", "Date of departure/outbound flight")
	returnDate    = flag.String("in", "20161103", "Date of return/inbound flight")
	noCache       = flag.Bool("noCache", false, "Do not read from cache")
)

type ApplicationParameters struct {
	KeyFile       string
	Key           string
	Origin        string
	DepartureDate string
	ReturnDate    string
	NoCache       bool
}

func ReadArguments() ApplicationParameters {
	flag.Parse()
	key := ReadKey(*keyFile)
	return ApplicationParameters{
		KeyFile: *keyFile, Key: key, Origin: *origin, DepartureDate: *departureDate, ReturnDate: *returnDate, NoCache: *noCache}
}

func main1() {
	arguments := ReadArguments()
	lr := CreateLiveRequestExample(arguments.Key, arguments.Origin, "VIE", arguments.DepartureDate, arguments.ReturnDate)
	PostAndPoll(lr)
}

func ReadFromFile(fileName string) (string, error) {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func ReadKey(fileName string) string {
	data, err := ReadFromFile(fileName)
	if err != nil {
		panic(err)
	}
	return strings.TrimSpace(data)
}

func main() {
	main2()
}

type RequestResults struct {
	Error error
	Data  *BrowseRoutesReply
}

func runAndPost(request BrowseRoutesRequest, engine RequestEngine, channel chan RequestResults) {
	data, err := runRequest(engine, request)
	channel <- RequestResults{err, data}
}

func DisplayResults(request BrowseRoutesRequest, quotes FullQuotes) {

	sort.Sort(sort.Reverse(quotes))
	for _, v := range quotes {
		link := GetLink(request.From, v.Destination.SkyscannerCode, request.DepartureDate, request.ReturnDate)
		fmt.Printf("%s %.0f %s %s\n", v.Destination.SkyscannerCode, v.Quote.MinPrice, v.Destination.Name, link)
	}
	fmt.Printf("%d results\n", len(quotes))
}

func LookForCountries(request BrowseRoutesRequest, countries []PlaceDto, engine RequestEngine) (FullQuotes, error) {

	countriesCount := len(countries)
	results := make(FullQuotes, 0, countriesCount)
	channel := make(chan RequestResults, countriesCount)

	for _, place := range countries {
		fmt.Printf("searching %s\n", place.Name)
		subRequest := BrowseRoutesRequest{request.Local, request.From, place.SkyscannerCode, request.DepartureDate, request.ReturnDate}
		go runAndPost(subRequest, engine, channel)
	}

	for i := 0; i < countriesCount; i++ {
		subResults := <-channel
		if subResults.Error != nil {
			return make(FullQuotes, 0, 0), subResults.Error
		}
		results = append(results, subResults.Data.GetBestQuotes()...)
	}

	return results, nil
}

func main2() {
	arguments := ReadArguments()
	engine := &LiveEngine{Key: arguments.Key}
	cache := CreateCache()
	ce := &CachedEngine{engine, cache, *noCache}
	sce := &SlowEngine{ce, time.Millisecond * 500}

	request := CreateBrowseRouteRequest(GetLocalExample(), arguments.Origin, arguments.DepartureDate, arguments.ReturnDate)
	reply, err := runRequest(sce, request)
	if err != nil {
		panic(err)
	}

	results, err := LookForCountries(request, reply.GetCountries(), sce)
	if err != nil {
		panic(err)
	}
	DisplayResults(request, results)
	sort.Sort(results)

	townsMap := make(map[string]PlaceDto)
	for _, v := range results {
		townsMap[v.Destination.SkyscannerCode] = v.Destination
	}
	towns := make([]PlaceDto, 0, 0)
	for k, v := range townsMap {
		fmt.Println(k, v)
		towns = append(towns, v)
	}

	results2, err := LookForCountries(request, towns, sce)
	DisplayResults(request, results2)

}
