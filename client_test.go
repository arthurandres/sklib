package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	TestKeyFile  = "testdata/key"
	TestKeyValue = "HELLOKEY"
)

func TestKey(t *testing.T) {
	key := ReadKey(TestKeyFile)

	assert.Equal(t, TestKeyValue, key)

}
