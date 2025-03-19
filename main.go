package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gosimple/slug"
	_ "github.com/mattn/go-sqlite3"
)

var (
	NamedayRe       = regexp.MustCompile(`^/nameday/*$`)
	NamedayReWithID = regexp.MustCompile(`^/nameday/([a-z0-9]+(?:-[a-z0-9]+)+)$`)
)

func main() {
	store := NewMemStore()
	namedayHandler := NewNamedayHandler(store)
	mux := http.NewServeMux()

	mux.Handle("/", &homeHandler{})
	mux.Handle("/nameday", namedayHandler)
	mux.Handle("/nameday/", namedayHandler)

	http.ListenAndServe(":8080", mux)
}

type homeHandler struct{}

func (h *homeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Open the database connection
	db, err := sql.Open("sqlite3", "./namedays.db")
	if err != nil {
		InternalServerErrorHandler(w, r)
		return
	}
	defer db.Close()

	// Get today's namedays
	names, err := getNameday(db)
	if err != nil {
		InternalServerErrorHandler(w, r)
		return
	}

	// Create HTML response
	var sb strings.Builder
	sb.WriteString("<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n")
	sb.WriteString("  <meta charset=\"UTF-8\">\n  <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n")
	sb.WriteString("  <title>Today's Namedays</title>\n</head>\n<body>\n")
	sb.WriteString(fmt.Sprintf("  <h1>Namedays for %s</h1>\n", time.Now().Format("January 2")))

	if len(names) == 0 {
		sb.WriteString("  <p>No namedays found for today</p>\n")
	} else {
		sb.WriteString("  <ul>\n")
		for _, name := range names {
			sb.WriteString(fmt.Sprintf("    <li>%s</li>\n", name))
		}
		sb.WriteString("  </ul>\n")
	}

	sb.WriteString("</body>\n</html>")
	w.Write([]byte(sb.String()))
}

func getNameday(db *sql.DB) ([]string, error) {
	// Get today's date in the format "MM-DD"
	today := time.Now().Format("01-02")

	// Query the database for names on today's date
	rows, err := db.Query("SELECT name FROM namedays WHERE date = ?", today)
	if err != nil {
		return nil, fmt.Errorf("error querying database: %w", err)
	}
	defer rows.Close()

	// Collect all names for today
	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		names = append(names, name)
	}

	// Check for any errors during iteration
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return names, nil
}

func ReadJSONFromURL(url string) (map[string][]string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-OK HTTP status: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	var result map[string][]string
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}

	return result, nil
}

func FilterNamedaysByMonth(namedays map[string][]string, month string) []string {
	var result []string
	for date, names := range namedays {
		if strings.HasPrefix(date, month) {
			result = append(result, fmt.Sprintf("%s: %s", date, strings.Join(names, ", ")))
		}
	}
	return result
}

func RenderHTMLList(items []string) string {
	var sb strings.Builder
	sb.WriteString("<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n")
	sb.WriteString("  <meta charset=\"UTF-8\">\n  <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n")
	sb.WriteString("  <title>Namedays</title>\n</head>\n<body>\n  <ul>\n")

	for _, item := range items {
		sb.WriteString(fmt.Sprintf("    <li>%s</li>\n", item))
	}

	sb.WriteString("  </ul>\n</body>\n</html>")
	return sb.String()
}

func GetCurrentMonth() string {
	return time.Now().Format("01")
}

type namedayStore interface {
	Add(name string, nameday Nameday) error
	Get(name string) (Nameday, error)
	List() (map[string]Nameday, error)
	Update(name string, nameday Nameday) error
	Remove(name string) error
}

type Nameday struct {
	Name string `json:"name"`
	Date string `json:"date"`
}

type NamedayHandler struct {
	store namedayStore
}

func NewMemStore() *MemStore {
	return &MemStore{
		data: make(map[string]Nameday),
	}
}

type MemStore struct {
	data map[string]Nameday
}

func (m *MemStore) Add(name string, nameday Nameday) error {
	m.data[name] = nameday
	return nil
}

func (m *MemStore) Get(name string) (Nameday, error) {
	nameday, exists := m.data[name]
	if !exists {
		return Nameday{}, fmt.Errorf("nameday not found")
	}
	return nameday, nil
}

func (m *MemStore) List() (map[string]Nameday, error) {
	return m.data, nil
}

func (m *MemStore) Update(name string, nameday Nameday) error {
	m.data[name] = nameday
	return nil
}

func (m *MemStore) Remove(name string) error {
	delete(m.data, name)
	return nil
}

func NewNamedayHandler(s namedayStore) *NamedayHandler {
	return &NamedayHandler{store: s}
}

func (h *NamedayHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodPost && NamedayRe.MatchString(r.URL.Path):
		h.CreateNameday(w, r)
	case r.Method == http.MethodGet && NamedayRe.MatchString(r.URL.Path):
		h.ListNamedays(w, r)
	case r.Method == http.MethodGet && NamedayReWithID.MatchString(r.URL.Path):
		h.GetNameday(w, r)
	case r.Method == http.MethodPut && NamedayReWithID.MatchString(r.URL.Path):
		h.UpdateNameday(w, r)
	case r.Method == http.MethodDelete && NamedayReWithID.MatchString(r.URL.Path):
		h.DeleteNameday(w, r)
	default:
		NotFoundHandler(w, r)
	}
}

func (h *NamedayHandler) GetNameday(w http.ResponseWriter, r *http.Request) {
	matches := NamedayReWithID.FindStringSubmatch(r.URL.Path)
	if len(matches) < 2 {
		InternalServerErrorHandler(w, r)
		return
	}

	nameday, err := h.store.Get(matches[1])
	if err != nil {
		NotFoundHandler(w, r)
		return
	}

	jsonBytes, err := json.Marshal(nameday)
	if err != nil {
		InternalServerErrorHandler(w, r)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func (h *NamedayHandler) CreateNameday(w http.ResponseWriter, r *http.Request) {
	var nameday Nameday
	if err := json.NewDecoder(r.Body).Decode(&nameday); err != nil {
		InternalServerErrorHandler(w, r)
		return
	}

	resourceID := slug.Make(nameday.Name)
	if err := h.store.Add(resourceID, nameday); err != nil {
		InternalServerErrorHandler(w, r)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *NamedayHandler) UpdateNameday(w http.ResponseWriter, r *http.Request) {
	matches := NamedayReWithID.FindStringSubmatch(r.URL.Path)
	if len(matches) < 2 {
		InternalServerErrorHandler(w, r)
		return
	}

	var nameday Nameday
	if err := json.NewDecoder(r.Body).Decode(&nameday); err != nil {
		InternalServerErrorHandler(w, r)
		return
	}

	if err := h.store.Update(matches[1], nameday); err != nil {
		InternalServerErrorHandler(w, r)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *NamedayHandler) ListNamedays(w http.ResponseWriter, r *http.Request) {
	namedaysList, err := h.store.List()
	if err != nil {
		InternalServerErrorHandler(w, r)
		return
	}

	jsonBytes, _ := json.Marshal(namedaysList)
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func (h *NamedayHandler) DeleteNameday(w http.ResponseWriter, r *http.Request) {
	matches := NamedayReWithID.FindStringSubmatch(r.URL.Path)
	if len(matches) < 2 {
		InternalServerErrorHandler(w, r)
		return
	}

	if err := h.store.Remove(matches[1]); err != nil {
		InternalServerErrorHandler(w, r)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func InternalServerErrorHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("500 Internal Server Error"))
}

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("404 Not Found"))
}

func GetCurrentMonthDate() string {
	return time.Now().Format("01-02")
}
