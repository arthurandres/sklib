package sklib

import (
	"fmt"
	"github.com/boltdb/bolt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

const (
	cacheLocation      = "cache.db"
	bucketName         = "cache"
	browseRouteFormat  = "http://partners.api.skyscanner.net/apiservices/browseroutes/v1.0/%s/%s/%s/%s/%s"
	browseRouteExample = "http://partners.api.skyscanner.net/apiservices/browseroutes/v1.0/GB/GBP/en-GB/LON/anywhere/20160819/20160821"
	liveUrl            = "http://partners.api.skyscanner.net/apiservices/pricing/v1.0"
	anywhere           = "anywhere"
	linkBase           = "https://www.skyscanner.net/transport/flights/%s/%s/%s/%s/"
	LocationKey        = "Location"
	ApiKeyTag          = "?apiKey="
)

type RequestEngine interface {
	Get(url string) (payload []byte, err error)
	PostAndPoll(url string, form url.Values) (payload []byte, err error)
}

type LiveEngine struct {
	Key string
}

func FormatKey(key string) string {
	return ApiKeyTag + key
}

func (m *LiveEngine) PostAndPoll(url string, form url.Values) ([]byte, error) {
	form.Set("apiKey", m.Key)

	resp, err := http.PostForm(url, form)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	location := resp.Header.Get(LocationKey)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", len(body))
	fmt.Println(location)
	fullUrl := location + FormatKey(m.Key)
	return Poll(fullUrl)
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

type WriteOnlyStore struct {
	Store CacheStore
}

type CachedEngine struct {
	Engine RequestEngine
	Cache  CacheStore
}

type SlowEngine struct {
	Engine RequestEngine
}

func (m *SlowEngine) Get(url string) ([]byte, error) {
	m.Wait()
	return m.Engine.Get(url)
}

func (m *SlowEngine) Wait() {
	random := rand.Intn(5000)
	time.Sleep(time.Millisecond * time.Duration(random))
}

func (m *SlowEngine) PostAndPoll(url string, form url.Values) ([]byte, error) {
	m.Wait()
	return m.Engine.PostAndPoll(url, form)
}

func (m *CachedEngine) PostAndPoll(url string, form url.Values) ([]byte, error) {
	cacheUrl := url + "?" + form.Encode()
	if cache := m.Cache.Get(cacheUrl); cache != nil {
		return cache, nil
	}
	payload, err := m.Engine.PostAndPoll(url, form)
	if err != nil {
		return nil, err
	}
	return payload, m.Cache.Set(cacheUrl, payload)
}

func (m *CachedEngine) Get(url string) ([]byte, error) {

	cache := m.Cache.Get(url)
	if cache != nil {
		return cache, nil
	}
	payload, err := m.Engine.Get(url)
	if err != nil {
		return nil, err
	}
	return payload, m.Cache.Set(url, payload)
}

type Localisation struct {
	Country  string
	Currency string
	Language string
}

func (m *Localisation) SubUrl() string {
	return fmt.Sprintf("%s/%s/%s",
		m.Country,
		m.Currency,
		m.Language)
}

type BrowseRoutesRequest struct {
	Localisation  Localisation
	Origin        string
	Destination   string
	DepartureDate string
	ReturnDate    string
}

func CreateBrowseRouteRequest(local Localisation, from string, departureDate string, returnDate string) BrowseRoutesRequest {
	return BrowseRoutesRequest{local, from, anywhere, departureDate, returnDate}
}

func (m BrowseRoutesRequest) Url() string {
	return fmt.Sprintf(
		browseRouteFormat,
		m.Localisation.SubUrl(),
		m.Origin,
		m.Destination,
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

func (m *WriteOnlyStore) Get(key string) []byte {
	return nil
}

func (m *WriteOnlyStore) Set(key string, data []byte) error {
	return m.Store.Set(key, data)
}

func Test(cs CacheStore) {
	cs.Get("hello")
	cs.Get("hello2")
	cs.Set("hello", []byte("world"))
	cs.Get("hello")

}

func CreateDB() *bolt.DB {
	var db *bolt.DB
	var err error
	db, err = bolt.Open(cacheLocation, 0600, nil)
	if err != nil {
		panic(err)
	}
	return db
}

func CreateCache(db *bolt.DB) *BoltStore {
	return &BoltStore{db, bucketName}
}

func insertAll(from map[string]float64, to map[string]float64) {
	for k, v := range from {
		to[k] = v
	}
}

type LiveRequest struct {
	Localisation  Localisation
	Origin        string
	Destination   string
	DepartureDate string
	ReturnDate    string
}

func (m *LiveRequest) Values() url.Values {
	return url.Values{
		"country":          {m.Localisation.Country},
		"currency":         {m.Localisation.Currency},
		"locale":           {m.Localisation.Language},
		"originplace":      {m.Origin},
		"destinationplace": {m.Destination},
		"outbounddate":     {FormatDate(m.DepartureDate)},
		"inbounddate":      {FormatDate(m.ReturnDate)},
		"locationschema":   {"Iata"}}
}

func (m *LiveRequest) Encode() string {
	return m.Values().Encode()
}

func CreateLiveRequest(
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

func Poll(url string) ([]byte, error) {
	for {
		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		if len(data) == 0 {
			fmt.Println("Empty")
			time.Sleep(1 * time.Second)
			continue
		}
		if err != nil {
			panic(err)
		}
		var reply LiveReply
		err = ParseJson(data, &reply)
		if err != nil {
			fmt.Println(string(data))
			panic(err)
		}
		fmt.Println(reply.Status)
		if reply.Status == UpdatesCompleteStatus {
			return data, nil
		}
		time.Sleep(1 * time.Second)
	}

}

func AppendTimeFilter(cp CompositeFilter, time *time.Duration, before, departure bool) CompositeFilter {
	if time != nil {
		cp = append(cp, &DepartureAfterFilter{
			Limit:     *time,
			Before:    before,
			Departure: departure})

	}
	return cp
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

type RequestResults struct {
	Error error
	Data  *BrowseRoutesReply
}

func runAndPost(request BrowseRoutesRequest, engine RequestEngine, channel chan RequestResults) {
	data, err := runRequest(engine, request)
	channel <- RequestResults{err, data}
}

func LookForCountries(request BrowseRoutesRequest, countries []PlaceDto, engine RequestEngine) (FullQuotes, error) {

	countriesCount := len(countries)
	results := make(FullQuotes, 0, countriesCount)
	channel := make(chan RequestResults, countriesCount)

	for _, place := range countries {
		subRequest := BrowseRoutesRequest{request.Localisation, request.Origin, place.SkyscannerCode, request.DepartureDate, request.ReturnDate}
		go runAndPost(subRequest, engine, channel)
	}

	for i := 0; i < countriesCount; i++ {
		fmt.Printf("\rReceiving %d/%d", i, countriesCount)
		subResults := <-channel
		if subResults.Error != nil {
			return make(FullQuotes, 0, 0), subResults.Error
		}
		results = append(results, subResults.Data.GetBestQuotes()...)
	}
	fmt.Printf("\n")

	return results, nil
}

func GetTowns(quotes FullQuotes) []PlaceDto {

	townsMap := make(map[string]PlaceDto)
	for _, v := range quotes {
		townsMap[v.Destination.SkyscannerCode] = v.Destination
	}
	towns := make([]PlaceDto, 0, 0)
	for _, v := range townsMap {
		towns = append(towns, v)
	}
	return towns
}

type SearchRequest struct {
	Localisation  Localisation
	Origin        string
	Destinations  []string
	DepartureDate string
	ReturnDate    string
}

func Search(engine RequestEngine, arguments SearchRequest) (Itineraries, error) {
	results := make(Itineraries, 0)
	for _, destination := range arguments.Destinations {
		fmt.Println("Searching", destination)
		liveRequest := CreateLiveRequest(
			arguments.Localisation,
			arguments.Origin,
			destination,
			arguments.DepartureDate,
			arguments.ReturnDate)
		data, err := engine.PostAndPoll(liveUrl, liveRequest.Values())
		if err != nil {
			return nil, err
		}
		var reply LiveReply
		err = ParseJson(data, &reply)
		if err != nil {
			return nil, err
		}
		fmt.Println("Results", reply.Stats())
		flightsData, err := ReadLiveReply(&reply)
		if err != nil {
			return nil, err
		}
		results = append(results, flightsData.Itineraries...)
	}

	return results, nil
}

func Browse(engine RequestEngine, arguments BrowseRoutesRequest) (FullQuotes, error) {

	request := CreateBrowseRouteRequest(arguments.Localisation, arguments.Origin, arguments.DepartureDate, arguments.ReturnDate)
	fmt.Println("Searching countries...")
	reply, err := runRequest(engine, request)
	if err != nil {
		panic(err)
	}

	results, err := LookForCountries(request, reply.GetCountries(), engine)
	if err != nil {
		panic(err)
	}
	sort.Sort(results)
	towns := GetTowns(results)

	fmt.Println("Searching towns...")
	return LookForCountries(request, towns, engine)
}

func FilterDirects(quotes FullQuotes) FullQuotes {
	results := make(FullQuotes, 0, len(quotes))
	for _, quote := range quotes {
		if quote.Quote.Direct {
			results = append(results, quote)
		}
	}
	return results
}
