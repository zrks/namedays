package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"k8s/pkg/recipes"

	"github.com/gosimple/slug"
)

var (
	RecipeRe       = regexp.MustCompile(`^/recipes/*$`)
	RecipeReWithID = regexp.MustCompile(`^/recipes/([a-z0-9]+(?:-[a-z0-9]+)+)$`)
)

func main() {
	store := recipes.NewMemStore()
	recipesHandler := NewRecipesHandler(store)
	mux := http.NewServeMux()

	mux.Handle("/", &homeHandler{})
	mux.Handle("/recipes", recipesHandler)
	mux.Handle("/recipes/", recipesHandler)

	http.ListenAndServe(":8080", mux)
}

type homeHandler struct{}

func (h *homeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	url := "https://gist.githubusercontent.com/laacz/5cccb056a533dffb2165/raw/5af9c97ef0b7c0256cbbf393bc45822aeb9ceba9/namedays-extended.json"
	parsedData, err := ReadJSONFromURL(url)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	htmlOutput := RenderHTMLList(parsedData[GetCurrentMonthDate()])
	w.Write([]byte(htmlOutput))
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

type recipeStore interface {
	Add(name string, recipes recipes.Recipe) error
	Get(name string) (recipes.Recipe, error)
	List() (map[string]recipes.Recipe, error)
	Update(name string, recipe recipes.Recipe) error
	Remove(name string) error
}

type RecipesHandler struct {
	store recipeStore
}

func NewRecipesHandler(s recipeStore) *RecipesHandler {
	return &RecipesHandler{store: s}
}

func (h *RecipesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodPost && RecipeRe.MatchString(r.URL.Path):
		h.CreateRecipe(w, r)
	case r.Method == http.MethodGet && RecipeRe.MatchString(r.URL.Path):
		h.ListRecipes(w, r)
	case r.Method == http.MethodGet && RecipeReWithID.MatchString(r.URL.Path):
		h.GetRecipe(w, r)
	case r.Method == http.MethodPut && RecipeReWithID.MatchString(r.URL.Path):
		h.UpdateRecipe(w, r)
	case r.Method == http.MethodDelete && RecipeReWithID.MatchString(r.URL.Path):
		h.DeleteRecipe(w, r)
	default:
		NotFoundHandler(w, r)
	}
}

func (h *RecipesHandler) GetRecipe(w http.ResponseWriter, r *http.Request) {
	matches := RecipeReWithID.FindStringSubmatch(r.URL.Path)
	if len(matches) < 2 {
		InternalServerErrorHandler(w, r)
		return
	}

	recipe, err := h.store.Get(matches[1])
	if err != nil {
		if err == recipes.NotFoundErr {
			NotFoundHandler(w, r)
			return
		}

		InternalServerErrorHandler(w, r)
		return
	}

	jsonBytes, err := json.Marshal(recipe)
	if err != nil {
		InternalServerErrorHandler(w, r)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func (h *RecipesHandler) CreateRecipe(w http.ResponseWriter, r *http.Request) {
	var recipe recipes.Recipe
	if err := json.NewDecoder(r.Body).Decode(&recipe); err != nil {
		InternalServerErrorHandler(w, r)
		return
	}

	resourceID := slug.Make(recipe.Name)
	if err := h.store.Add(resourceID, recipe); err != nil {
		fmt.Println(err)
		InternalServerErrorHandler(w, r)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *RecipesHandler) UpdateRecipe(w http.ResponseWriter, r *http.Request) {
	matches := RecipeReWithID.FindStringSubmatch(r.URL.Path)
	if len(matches) < 2 {
		InternalServerErrorHandler(w, r)
		return
	}

	// Recipe object that will be populated from JSON payload
	var recipe recipes.Recipe
	if err := json.NewDecoder(r.Body).Decode(&recipe); err != nil {
		InternalServerErrorHandler(w, r)
		return
	}

	if err := h.store.Update(matches[1], recipe); err != nil {
		if err == recipes.NotFoundErr {
			NotFoundHandler(w, r)
			return
		}
		InternalServerErrorHandler(w, r)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *RecipesHandler) ListRecipes(w http.ResponseWriter, r *http.Request) {
	recipesList, err := h.store.List()
	if err != nil {
		InternalServerErrorHandler(w, r)
		return
	}

	jsonBytes, _ := json.Marshal(recipesList)
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func (h *RecipesHandler) DeleteRecipe(w http.ResponseWriter, r *http.Request) {
	matches := RecipeReWithID.FindStringSubmatch(r.URL.Path)
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

func RenderHTMLList(items []string) string {
	var sb strings.Builder
	sb.WriteString("<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n")
	sb.WriteString("  <meta charset=\"UTF-8\">\n  <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n")
	sb.WriteString("  <title>Generated List</title>\n</head>\n<body>\n  <ul>\n")

	for _, item := range items {
		sb.WriteString(fmt.Sprintf("    <li>%s</li>\n", item))
	}

	sb.WriteString("  </ul>\n</body>\n</html>")
	return sb.String()
}

func GetCurrentMonthDate() string {
	return time.Now().Format("01-02")
}
