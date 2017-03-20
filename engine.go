package sklib

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"github.com/boltdb/bolt"
)

type RequestEngine interface {
	Get(url string) (payload []byte, err error)
	PostAndPoll(url string, form url.Values) (payload []byte, err error)
}

type LiveEngine struct {
	Key string
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
	cacheURL := url + "?" + form.Encode()
	if cache := m.Cache.Get(cacheURL); cache != nil {
		return cache, nil
	}
	payload, err := m.Engine.PostAndPoll(url, form)
	if err != nil {
		return nil, err
	}
	return payload, m.Cache.Set(cacheURL, payload)
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

func (m *LiveEngine) PostAndPoll(url string, form url.Values) ([]byte, error) {
	form.Set("apiKey", m.Key)

	resp, err := http.PostForm(url, form)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	location := resp.Header.Get(locationKey)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", len(body))
	fmt.Println(location)
	fullUrl := location + formatKey(m.Key)
	return Poll(fullUrl)
}

func (m *LiveEngine) Get(url string) ([]byte, error) {

	fullUrl := url + formatKey(m.Key)
	resp, err := http.Get(fullUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)

}

func CreateEngine(key string, noCache bool) (RequestEngine, func() error) {
	engine := &LiveEngine{Key: key}
	db := CreateDB()
	var cache CacheStore = CreateCache(db)
	if noCache {
		cache = &WriteOnlyStore{Store: cache}
	}
	return &CachedEngine{engine, cache}, db.Close
}

func CreateCache(db *bolt.DB) *BoltStore {
	return &BoltStore{db, bucketName}
}
