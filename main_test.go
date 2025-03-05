package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNamedayHandler_Integration(t *testing.T) {
	store := NewMemStore()
	namedayHandler := NewNamedayHandler(store)
	namedayData := readTestData(t, "nameday_data.json")
	namedayReader := bytes.NewReader(namedayData)

	req := httptest.NewRequest(http.MethodPost, "/nameday", namedayReader)
	w := httptest.NewRecorder()
	namedayHandler.ServeHTTP(w, req)
	res := w.Result()
	defer res.Body.Close()
	assert.Equal(t, 200, res.StatusCode, "Expected status code 200 for POST request")

	saved, _ := store.List()
	assert.Len(t, saved, 1, "Expected store to have 1 item after POST request")

	req = httptest.NewRequest(http.MethodGet, "/nameday/john", nil)
	w = httptest.NewRecorder()
	namedayHandler.ServeHTTP(w, req)
	res = w.Result()
	defer res.Body.Close()
	assert.Equal(t, 200, res.StatusCode, "Expected status code 200 for GET request")

	data, err := io.ReadAll(res.Body)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	assert.JSONEq(t, string(namedayData), string(data), "Expected JSON data to match")

	updatedNamedayData := readTestData(t, "nameday_data.json")
	updatedNamedayReader := bytes.NewReader(updatedNamedayData)
	req = httptest.NewRequest(http.MethodPut, "/nameday/john", updatedNamedayReader)
	w = httptest.NewRecorder()
	namedayHandler.ServeHTTP(w, req)
	res = w.Result()
	defer res.Body.Close()
	assert.Equal(t, 200, res.StatusCode, "Expected status code 200 for PUT request")

	updatedNameday, err := store.Get("john")
	assert.NoError(t, err, "Expected no error when getting updated nameday")
	updatedNamedayDataBytes, _ := json.Marshal(updatedNameday)
	assert.JSONEq(t, string(updatedNamedayData), string(updatedNamedayDataBytes), "Expected JSON data to match")

	req = httptest.NewRequest(http.MethodDelete, "/nameday/john", nil)
	w = httptest.NewRecorder()
	namedayHandler.ServeHTTP(w, req)
	res = w.Result()
	defer res.Body.Close()
	assert.Equal(t, 200, res.StatusCode, "Expected status code 200 for DELETE request")

	saved, _ = store.List()
	assert.Len(t, saved, 0, "Expected store to be empty after DELETE request")
}

func readTestData(t *testing.T, name string) []byte {
	t.Helper()
	content, err := os.ReadFile(os.Getenv("PWD") + "/testdata/" + name)
	if err != nil {
		t.Fatalf("Could not read %v: %v", name, err)
	}
	return content
}

