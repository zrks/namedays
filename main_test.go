package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const (
	errFailedToCreateRequest = "Failed to create request: %v"
	testPersonKey            = "test-person"
	testJohnSmith            = "John Smith"
	johnSmithKey             = "john-smith"
	johnSmithPath            = "/nameday/" + johnSmithKey
	errWrongStatusCode       = "Handler returned wrong status code: got %v want %v"
)

func TestNamedayHandlerCreateNameday(t *testing.T) {
	store := NewMemStore()
	handler := NewNamedayHandler(store)

	// Test data
	nameday := Nameday{
		Name: testJohnSmith,
		Date: "04-12",
	}
	jsonData, _ := json.Marshal(nameday)

	// Create a request
	req, err := http.NewRequest(http.MethodPost, "/nameday", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf(errFailedToCreateRequest, err)
	}

	// Record the response
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Check response
	if status := rr.Code; status != http.StatusOK {
		t.Errorf(errWrongStatusCode, status, http.StatusOK)
	}

	// Verify data was stored correctly
	storedNameday, err := store.Get(johnSmithKey)
	if err != nil {
		t.Fatalf("Failed to retrieve created nameday: %v", err)
	}
	if storedNameday.Name != nameday.Name || storedNameday.Date != nameday.Date {
		t.Errorf("Stored nameday does not match input: got %v want %v", storedNameday, nameday)
	}
}

func TestNamedayHandlerGetNameday(t *testing.T) {
	store := NewMemStore()
	handler := NewNamedayHandler(store)

	// Add test data
	testNameday := Nameday{
		Name: testJohnSmith,
		Date: "04-12",
	}
	store.Add(johnSmithKey, testNameday)

	// Create a request
	req, err := http.NewRequest(http.MethodGet, johnSmithPath, nil)
	if err != nil {
		t.Fatalf(errFailedToCreateRequest, err)
	}

	// Record the response
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Check response
	if status := rr.Code; status != http.StatusOK {
		t.Errorf(errWrongStatusCode, status, http.StatusOK)
	}

	// Verify response data
	var responseNameday Nameday
	if err := json.Unmarshal(rr.Body.Bytes(), &responseNameday); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if responseNameday.Name != testNameday.Name || responseNameday.Date != testNameday.Date {
		t.Errorf("Response nameday does not match test data: got %v want %v", responseNameday, testNameday)
	}
}

func TestNamedayHandlerGetNamedayNotFound(t *testing.T) {
	store := NewMemStore()
	handler := NewNamedayHandler(store)

	// Create a request for non-existent nameday
	req, err := http.NewRequest(http.MethodGet, "/nameday/non-existent", nil)
	if err != nil {
		t.Fatalf(errFailedToCreateRequest, err)
	}

	// Record the response
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Check response - should be 404
	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf(errWrongStatusCode, status, http.StatusNotFound)
	}
}

func TestNamedayHandlerUpdateNameday(t *testing.T) {
	store := NewMemStore()
	handler := NewNamedayHandler(store)

	// Add initial data
	initialNameday := Nameday{
		Name: testJohnSmith,
		Date: "04-12",
	}
	store.Add(johnSmithKey, initialNameday)

	// Updated data
	updatedNameday := Nameday{
		Name: testJohnSmith,
		Date: "05-15", // Changed date
	}
	jsonData, _ := json.Marshal(updatedNameday)

	// Create a request
	req, err := http.NewRequest(http.MethodPut, johnSmithPath, bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf(errFailedToCreateRequest, err)
	}

	// Record the response
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Check response
	if status := rr.Code; status != http.StatusOK {
		t.Errorf(errWrongStatusCode, status, http.StatusOK)
	}

	// Verify data was updated correctly
	storedNameday, err := store.Get(johnSmithKey)
	if err != nil {
		t.Fatalf("Failed to retrieve updated nameday: %v", err)
	}
	if storedNameday.Date != updatedNameday.Date {
		t.Errorf("Updated nameday does not match: got %v want %v", storedNameday.Date, updatedNameday.Date)
	}
}

func TestNamedayHandlerDeleteNameday(t *testing.T) {
	store := NewMemStore()
	handler := NewNamedayHandler(store)

	// Add test data
	testNameday := Nameday{
		Name: testJohnSmith,
		Date: "04-12",
	}
	store.Add(johnSmithKey, testNameday)

	// Create a request
	req, err := http.NewRequest(http.MethodDelete, johnSmithPath, nil)
	if err != nil {
		t.Fatalf(errFailedToCreateRequest, err)
	}

	// Record the response
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Check response
	if status := rr.Code; status != http.StatusOK {
		t.Errorf(errWrongStatusCode, status, http.StatusOK)
	}

	// Verify data was deleted
	_, err = store.Get(johnSmithKey)
	if err == nil {
		t.Errorf("Nameday was not deleted as expected")
	}
}

func TestNamedayHandlerListNamedays(t *testing.T) {
	store := NewMemStore()
	handler := NewNamedayHandler(store)

	// Add test data
	testNamedays := []Nameday{
		{Name: testJohnSmith, Date: "04-12"},
		{Name: "Jane Doe", Date: "05-15"},
	}
	store.Add(johnSmithKey, testNamedays[0])
	store.Add("jane-doe", testNamedays[1])

	// Create a request
	req, err := http.NewRequest(http.MethodGet, "/nameday", nil)
	if err != nil {
		t.Fatalf(errFailedToCreateRequest, err)
	}

	// Record the response
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Check response
	if status := rr.Code; status != http.StatusOK {
		t.Errorf(errWrongStatusCode, status, http.StatusOK)
	}

	// Verify response data
	var responseNamedays map[string]Nameday
	if err := json.Unmarshal(rr.Body.Bytes(), &responseNamedays); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if len(responseNamedays) != 2 {
		t.Errorf("Expected 2 namedays, got %d", len(responseNamedays))
	}
	if _, exists := responseNamedays[johnSmithKey]; !exists {
		t.Errorf("Expected '%s' to be in the response", johnSmithKey)
	}
	if _, exists := responseNamedays["jane-doe"]; !exists {
		t.Errorf("Expected 'jane-doe' to be in the response")
	}
}

func TestNamedayHandlerInvalidMethod(t *testing.T) {
	store := NewMemStore()
	handler := NewNamedayHandler(store)

	// Create a request with invalid method
	req, err := http.NewRequest(http.MethodPatch, johnSmithPath, nil)
	if err != nil {
		t.Fatalf(errFailedToCreateRequest, err)
	}

	// Record the response
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Check response - should be 404 (not found for invalid method)
	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf(errWrongStatusCode, status, http.StatusNotFound)
	}
}

func TestReadJSONFromURL(t *testing.T) {
	// Mock HTTP server to return test JSON data
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockJSON := `{"01-01":["New Year"],"03-15":["Test Day"]}`
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(mockJSON))
	}))
	defer testServer.Close()

	// Test the function
	result, err := ReadJSONFromURL(testServer.URL)
	if err != nil {
		t.Fatalf("ReadJSONFromURL returned an error: %v", err)
	}

	// Verify results
	if len(result) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(result))
	}
	if names, exists := result["01-01"]; !exists || names[0] != "New Year" {
		t.Errorf("Expected '01-01' to contain 'New Year', got %v", names)
	}
}

func TestFilterNamedaysByMonth(t *testing.T) {
	testData := map[string][]string{
		"01-01": {"New Year"},
		"01-15": {"Mid January"},
		"02-01": {"First February"},
	}

	// Test filtering for January (01)
	januaryResults := FilterNamedaysByMonth(testData, "01")
	if len(januaryResults) != 2 {
		t.Errorf("Expected 2 namedays in January, got %d", len(januaryResults))
	}

	// Test filtering for February (02)
	februaryResults := FilterNamedaysByMonth(testData, "02")
	if len(februaryResults) != 1 {
		t.Errorf("Expected 1 nameday in February, got %d", len(februaryResults))
	}

	// Test filtering for March (03) - should be empty
	marchResults := FilterNamedaysByMonth(testData, "03")
	if len(marchResults) != 0 {
		t.Errorf("Expected 0 namedays in March, got %d", len(marchResults))
	}
}

func TestGetCurrentMonth(t *testing.T) {
	// Current month should match what time.Now() returns
	expected := time.Now().Format("01")
	result := GetCurrentMonth()
	if result != expected {
		t.Errorf("GetCurrentMonth() returned %s, expected %s", result, expected)
	}
}

func TestGetCurrentMonthDate(t *testing.T) {
	// Current month-date should match what time.Now() returns
	expected := time.Now().Format("01-02")
	result := GetCurrentMonthDate()
	if result != expected {
		t.Errorf("GetCurrentMonthDate() returned %s, expected %s", result, expected)
	}
}

func TestRenderHTMLList(t *testing.T) {
	testItems := []string{"Item 1", "Item 2"}
	result := RenderHTMLList(testItems)

	// Check that result contains the test items
	for _, item := range testItems {
		if !bytes.Contains([]byte(result), []byte(item)) {
			t.Errorf("RenderHTMLList output does not contain expected item: %s", item)
		}
	}

	// Check that it has HTML structure
	if !bytes.Contains([]byte(result), []byte("<!DOCTYPE html>")) {
		t.Errorf("RenderHTMLList output does not contain DOCTYPE declaration")
	}
	if !bytes.Contains([]byte(result), []byte("<html")) {
		t.Errorf("RenderHTMLList output does not contain html tag")
	}
	if !bytes.Contains([]byte(result), []byte("<li>")) {
		t.Errorf("RenderHTMLList output does not contain list items")
	}
}

func TestMemStore(t *testing.T) {
	store := NewMemStore()

	// Test Add and Get
	nameday := Nameday{Name: "Test Person", Date: "05-05"}
	err := store.Add(testPersonKey, nameday)
	if err != nil {
		t.Fatalf("Failed to add nameday: %v", err)
	}

	retrieved, err := store.Get(testPersonKey)
	if err != nil {
		t.Fatalf("Failed to get nameday: %v", err)
	}
	if retrieved.Name != nameday.Name || retrieved.Date != nameday.Date {
		t.Errorf("Retrieved nameday does not match: got %v want %v", retrieved, nameday)
	}

	// Test List
	namedaysList, err := store.List()
	if err != nil {
		t.Fatalf("Failed to list namedays: %v", err)
	}
	if len(namedaysList) != 1 {
		t.Errorf("Expected 1 nameday in list, got %d", len(namedaysList))
	}

	// Test Update
	updatedNameday := Nameday{Name: "Test Person Updated", Date: "06-06"}
	err = store.Update(testPersonKey, updatedNameday)
	if err != nil {
		t.Fatalf("Failed to update nameday: %v", err)
	}

	retrieved, err = store.Get(testPersonKey)
	if err != nil {
		t.Fatalf("Failed to get updated nameday: %v", err)
	}
	if retrieved.Name != updatedNameday.Name || retrieved.Date != updatedNameday.Date {
		t.Errorf("Updated nameday does not match: got %v want %v", retrieved, updatedNameday)
	}

	// Test Remove
	err = store.Remove(testPersonKey)
	if err != nil {
		t.Fatalf("Failed to remove nameday: %v", err)
	}

	_, err = store.Get(testPersonKey)
	if err == nil {
		t.Errorf("Expected error when getting removed nameday")
	}
}
