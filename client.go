package sklib

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"time"
)

const (
	browseRouteFormat  = "http://partners.api.skyscanner.net/apiservices/browseroutes/v1.0/%s/%s/%s/%s/%s"
	browseRouteExample = "http://partners.api.skyscanner.net/apiservices/browseroutes/v1.0/GB/GBP/en-GB/LON/anywhere/20160819/20160821"
	liveURL            = "http://partners.api.skyscanner.net/apiservices/pricing/v1.0"
	anywhere           = "anywhere"
	linkBase           = "https://www.skyscanner.net/transport/flights/%s/%s/%s/%s/"
	locationKey        = "Location"
	apiKeyTag          = "?apiKey="
)

type RequestResults struct {
	Error error
	Data  *BrowseRoutesReply
}

func RunRequest(engine RequestEngine, r BrowseRoutesRequest) (*BrowseRoutesReply, error) {
	url := r.Url()
	fmt.Println(url)
	data, err := engine.Get(url)
	if err != nil {
		return nil, err
	} else {
		results := ParseBrowseRoutesReplyJson(data)
		return results, nil
	}

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

func runAndPost(request BrowseRoutesRequest, engine RequestEngine, channel chan RequestResults) {
	data, err := RunRequest(engine, request)
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

func Search(engine RequestEngine, arguments SearchRequest) (Itineraries, error) {
	results := make(Itineraries, 0)
	for _, destination := range arguments.Destinations {
		fmt.Println("Searching", destination)
		liveRequest := NewLiveRequest(
			arguments.Localisation,
			arguments.Origin,
			destination,
			arguments.DepartureDate,
			arguments.ReturnDate)
		data, err := engine.PostAndPoll(liveURL, liveRequest.Values())
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

	request := NewBrowseRouteRequest(arguments.Localisation, arguments.Origin, arguments.DepartureDate, arguments.ReturnDate)
	fmt.Println("Searching countries...")
	reply, err := RunRequest(engine, request)
	if err != nil {
		return nil, err
	}

	results, err := LookForCountries(request, reply.GetCountries(), engine)
	if err != nil {
		return nil, err
	}
	sort.Sort(results)
	towns := results.GetTowns()

	fmt.Println("Searching towns...")
	return LookForCountries(request, towns, engine)
}
