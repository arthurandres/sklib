package sklib

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMapCarriers(t *testing.T) {
	carriers := GetTestLiveReply().Carriers
	mapped, err := MapCarriers(carriers)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, len(carriers), len(mapped))
}

func TestMapPlaces(t *testing.T) {
	reply := GetTestLiveReply()
	places, err := MapPlaces(reply.Places)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, len(reply.Places), len(places))
	assert.Equal(t, "LGW", places[13542].Code)
	parent := getParentPlace("13542", places)
	assert.Equal(t, "LGW", parent.Code)
	assert.Equal(t, "LON", places[13542].Parent.Code)
	assert.Equal(t, "GB", places[13542].Parent.Parent.Code)
}

func TestReadLiveReply(t *testing.T) {

	reply := GetTestLiveReply()
	data, err := ReadLiveReply(reply)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, len(reply.Itineraries), len(data.Itineraries))

}

func TestReadSegment(t *testing.T) {
	reply := GetTestLiveReply()
	places, err := MapPlaces(reply.Places)
	if err != nil {
		panic(err)
	}
	carriers, err := MapCarriers(reply.Carriers)
	if err != nil {
		panic(err)
	}
	segment, err := ReadSegment(reply.Segments[0], places, carriers)
	if err != nil {
		panic(err)
	}

	assert.Equal(t, "LGW", segment.Origin.Code)
	fmt.Println(segment)
}

func TestReadSegments(t *testing.T) {
	reply := GetTestLiveReply()
	places, err := MapPlaces(reply.Places)
	carriers, err := MapCarriers(reply.Carriers)
	segments, err := ReadSegments(reply.Segments, places, carriers)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, len(reply.Segments), len(segments))
	fmt.Println(segments)
}
