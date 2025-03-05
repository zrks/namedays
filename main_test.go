package main

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"net/http"
	"net/http/httptest"
)

func TestNamedayHandler_Integration(t *testing.T) {
	// Create a MemStore and Nameday Handler
	store := NewMemStore()
	namedayHandler := NewNamedayHandler(store)

	// Test data for nameday
	namedayData := readTestData(t, "nameday_data.json")
	namedayReader := bytes.NewReader(namedayData)

	// CREATE - add a new nameday
	req := httptest.NewRequest(http.MethodPost, "/nameday", namedayReader)
	w := httptest.NewRecorder()
	namedayHandler.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()
	assert.Equal(t, 200, res.StatusCode)

	saved, _ := store.List()
	assert.Len(t, saved, 1)

	// GET - find the nameday you just added
	req = httptest.NewRequest(http.MethodGet, "/nameday/john", nil)
	w = httptest.NewRecorder()
	namedayHandler.ServeHTTP(w, req)

	res = w.Result()
	defer res.Body.Close()
	assert.Equal(t, 200, res.StatusCode)

	data, err := io.ReadAll(res.Body)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	assert.JSONEq(t, string(namedayData), string(data))

	// UPDATE - update nameday data
	updatedNamedayData := readTestData(t, "updated_nameday_data.json")
	updatedNamedayReader := bytes.NewReader(updatedNamedayData)

	req = httptest.NewRequest(http.MethodPut, "/nameday/john", updatedNamedayReader)
	w = httptest.NewRecorder()
	namedayHandler.ServeHTTP(w, req)

	res = w.Result()
	defer res.Body.Close()
	assert.Equal(t, 200, res.StatusCode)

	updatedNameday, err := store.Get("john")
	assert.NoError(t, err)

	updatedNamedayDataBytes, _ := json.Marshal(updatedNameday)
	assert.JSONEq(t, string(updatedNamedayData), string(updatedNamedayDataBytes))

	// DELETE - remove the nameday
	req = httptest.NewRequest(http.MethodDelete, "/nameday/john", nil)
	w = httptest.NewRecorder()
	namedayHandler.ServeHTTP(w, req)

	res = w.Result()
	defer res.Body.Close()
	assert.Equal(t, 200, res.StatusCode)

	saved, _ = store.List()
	assert.Len(t, saved, 0)
}

func readTestData(t *testing.T, name string) []byte {
	t.Helper()
	content, err := os.ReadFile(os.Getenv("PWD") + "/testdata/" + name)
	if err != nil {
		t.Errorf("Could not read %v", name)
	}

	return content
}