package main

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/url"
	"os"
	"testing"
	"testutil"
)

func TestMain(m *testing.M) {
	// Create http.Client or http.Transport if necessary

	os.Exit(m.Run())
}
