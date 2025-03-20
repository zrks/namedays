package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
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

// Helper functions to reduce duplication
func createTestNamedayHandler() (*MemStore, *NamedayHandler) {
	store := NewMemStore()
	handler := NewNamedayHandler(store)
	return store, handler
}

func setupTestRequest(t *testing.T, method, path string, body []byte) (*httptest.ResponseRecorder, *http.Request) {
	req, err := http.NewRequest(method, path, bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf(errFailedToCreateRequest, err)
	}
	return httptest.NewRecorder(), req
}

func createTestDb(t *testing.T) (string, *sql.DB) {
	// Create a test database file
	tmpDB, err := os.CreateTemp("", "test-namedays-*.db")
	if err != nil {
		t.Fatal("Failed to create temporary database:", err)
	}
	tmpDBPath := tmpDB.Name()
	tmpDB.Close()

	// Initialize test database
	db, err := sql.Open("sqlite3", tmpDBPath)
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}

	// Create table
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS namedays (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		date TEXT NOT NULL,
		name TEXT NOT NULL
	);`)
	if err != nil {
		t.Fatal("Failed to create table:", err)
	}

	// Register cleanup
	t.Cleanup(func() {
		db.Close()
		os.Remove(tmpDBPath)
	})

	return tmpDBPath, db
}

func addTestNameday(store *MemStore, key string, name, date string) Nameday {
	nameday := Nameday{
		Name: name,
		Date: date,
	}
	store.Add(key, nameday)
	return nameday
}

func checkResponseStatus(t *testing.T, rr *httptest.ResponseRecorder, expected int) {
	if status := rr.Code; status != expected {
		t.Errorf(errWrongStatusCode, status, expected)
	}
}

func TestNamedayHandlerCreateNameday(t *testing.T) {
	store, handler := createTestNamedayHandler()

	// Test data
	nameday := Nameday{
		Name: testJohnSmith,
		Date: "04-12",
	}
	jsonData, _ := json.Marshal(nameday)

	// Create a request and record response
	rr, req := setupTestRequest(t, http.MethodPost, "/nameday", jsonData)
	handler.ServeHTTP(rr, req)

	// Check response
	checkResponseStatus(t, rr, http.StatusOK)

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
	store, handler := createTestNamedayHandler()

	// Add test data
	testNameday := addTestNameday(store, johnSmithKey, testJohnSmith, "04-12")

	// Create a request and record response
	rr, req := setupTestRequest(t, http.MethodGet, johnSmithPath, nil)
	handler.ServeHTTP(rr, req)

	// Check response
	checkResponseStatus(t, rr, http.StatusOK)

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
	_, handler := createTestNamedayHandler()

	// Create a request for non-existent nameday and record response
	rr, req := setupTestRequest(t, http.MethodGet, "/nameday/non-existent", nil)
	handler.ServeHTTP(rr, req)

	// Check response - should be 404
	checkResponseStatus(t, rr, http.StatusNotFound)
}

func TestNamedayHandlerUpdateNameday(t *testing.T) {
	store, handler := createTestNamedayHandler()

	// Add initial data
	addTestNameday(store, johnSmithKey, testJohnSmith, "04-12")

	// Updated data
	updatedNameday := Nameday{
		Name: testJohnSmith,
		Date: "05-15", // Changed date
	}
	jsonData, _ := json.Marshal(updatedNameday)

	// Create a request and record response
	rr, req := setupTestRequest(t, http.MethodPut, johnSmithPath, jsonData)
	handler.ServeHTTP(rr, req)

	// Check response
	checkResponseStatus(t, rr, http.StatusOK)

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
	store, handler := createTestNamedayHandler()

	// Add test data
	addTestNameday(store, johnSmithKey, testJohnSmith, "04-12")

	// Create a request and record response
	rr, req := setupTestRequest(t, http.MethodDelete, johnSmithPath, nil)
	handler.ServeHTTP(rr, req)

	// Check response
	checkResponseStatus(t, rr, http.StatusOK)

	// Verify data was deleted
	_, err := store.Get(johnSmithKey)
	if err == nil {
		t.Errorf("Nameday was not deleted as expected")
	}
}

func TestNamedayHandlerListNamedays(t *testing.T) {
	store, handler := createTestNamedayHandler()

	// Add test data
	addTestNameday(store, johnSmithKey, testJohnSmith, "04-12")
	addTestNameday(store, "jane-doe", "Jane Doe", "05-15")

	// Create a request and record response
	rr, req := setupTestRequest(t, http.MethodGet, "/nameday", nil)
	handler.ServeHTTP(rr, req)

	// Check response
	checkResponseStatus(t, rr, http.StatusOK)

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
	_, handler := createTestNamedayHandler()

	// Create a request with invalid method and record response
	rr, req := setupTestRequest(t, http.MethodPatch, johnSmithPath, nil)
	handler.ServeHTTP(rr, req)

	// Check response - should be 404 (not found for invalid method)
	checkResponseStatus(t, rr, http.StatusNotFound)
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

// TestHomeHandler tests the home page handler
func TestHomeHandler(t *testing.T) {
	tmpDBPath, db := createTestDb(t)

	// Insert test data with today's date
	today := time.Now().Format("01-02")
	_, err := db.Exec("INSERT INTO namedays (date, name) VALUES (?, ?)", today, "Test Name")
	if err != nil {
		t.Fatal("Failed to insert test data:", err)
	}

	// Create request and record response
	rr, req := setupTestRequest(t, http.MethodGet, "/", nil)

	// Use the new constructor with our test DB path
	handler := NewHomeHandler(tmpDBPath)

	// Call the handler
	handler.ServeHTTP(rr, req)

	// Check response code
	checkResponseStatus(t, rr, http.StatusOK)

	// Check that the response contains HTML
	if !bytes.Contains(rr.Body.Bytes(), []byte("<!DOCTYPE html>")) {
		t.Error("Response does not contain HTML")
	}
}

func TestGetNamedayDB(t *testing.T) {
	_, db := createTestDb(t)

	// Insert test data
	today := time.Now().Format("01-02")
	testName := "Today's Test Name"
	_, err := db.Exec("INSERT INTO namedays (date, name) VALUES (?, ?)", today, testName)
	if err != nil {
		t.Fatal("Failed to insert test data:", err)
	}

	// Call the function
	names, err := getNameday(db)
	if err != nil {
		t.Fatal("getNameday returned an error:", err)
	}

	// Verify the result
	if len(names) != 1 {
		t.Errorf("Expected 1 name, got %d", len(names))
	}
	if len(names) > 0 && names[0] != testName {
		t.Errorf("Expected name %s, got %s", testName, names[0])
	}
}
