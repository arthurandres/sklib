package main

import (
	"flag"
	"fmt"
	"github.com/boltdb/bolt"
	"io/ioutil"
	"net/http"
	"sort"
)

const (
	browseRouteFormat  = "http://partners.api.skyscanner.net/apiservices/browseroutes/v1.0/%s/%s/%s/%s/%s"
	browseRouteExample = "http://partners.api.skyscanner.net/apiservices/browseroutes/v1.0/GB/GBP/en-GB/LON/anywhere/20160819/20160821"
	anywhere           = "anywhere"
	linkBase           = "https://www.skyscanner.net/transport/flights/%s/%s/%s/%s/"
)

type RequestEngine interface {
	Get(url string) (payload []byte, err error)
}

type LiveEngine struct {
	Key string
}

func (m *LiveEngine) Get(url string) ([]byte, error) {

	fullUrl := url + "?apiKey=" + m.Key
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
	Engine RequestEngine
	Cache  CacheStore
}

func (m CachedEngine) Get(url string) ([]byte, error) {
	cache := m.Cache.Get(url)
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

func CreateAnywhereQuery(local Local, from string, departureDate string, returnDate string) BrowseRoutesRequest {
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

func GetLink(from string, to string, outbound string, inbound string) string {
	return fmt.Sprintf(linkBase, from, to, outbound, inbound)
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
	return CreateAnywhereQuery(GetLocalExample(), "LON", "20161101", "20161103")
}

func insertAll(from map[string]float64, to map[string]float64) {
	for k, v := range from {
		to[k] = v
	}
}

var (
	key          = flag.String("key", "ar926739961631567929873917891697", "API key provided by skyscanner")
	origin       = flag.String("from", "LON", "Origin Town/Airport")
	outboundDate = flag.String("out", "20161101", "Date of departure/outbound flight")
	inboundDate  = flag.String("in", "20161103", "Date of return/inbound flight")
)

func main() {
	flag.Parse()

	engine := &LiveEngine{Key: *key}
	cache := CreateCache()
	ce := CachedEngine{engine, cache}
	request := CreateAnywhereQuery(GetLocalExample(), *origin, *outboundDate, *inboundDate)
	reply, err := runRequest(ce, request)
	if err != nil {
		panic(err)
	}
	results := make(FullQuotes, 0, 100)
	for _, place := range reply.GetCountries() {
		fmt.Printf("searching %s\n", place.Name)
		subRequest := BrowseRoutesRequest{request.Local, request.From, place.SkyscannerCode, request.DepartureDate, request.ReturnDate}
		subReply, err := runRequest(ce, subRequest)
		if err != nil {
			panic(err)
		}
		results = append(results, subReply.GetFullQuotes()...)
	}

	sort.Sort(results)
	sort.Sort(sort.Reverse(results))
	for _, v := range results {
		link := GetLink(request.From, v.Destination.SkyscannerCode, request.DepartureDate, request.ReturnDate)
		fmt.Printf("%s %.0f %s %s\n", v.Destination.SkyscannerCode, v.Quote.MinPrice, v.Destination.Name, link)
	}
	fmt.Printf("%d results\n", len(results))
}
