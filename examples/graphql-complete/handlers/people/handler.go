package people

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/jsocol/rest-data-loader/examples/graphql-complete/handlers"
)

//go:embed people.json
var peopleData []byte

type Person struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Handler struct {
	people map[string]Person
}

func New() *Handler {
	var rawPeople []Person
	decoder := json.NewDecoder(bytes.NewReader(peopleData))
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&rawPeople)
	if err != nil {
		panic(err)
	}

	people := make(map[string]Person, len(rawPeople))

	for _, person := range rawPeople {
		slog.Debug("loading person", "person", person)
		people[person.ID] = person
	}

	return &Handler{
		people: people,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", handlers.ContentTypeJSON)
	jw := json.NewEncoder(w)

	qv := r.URL.Query()
	ids, ok := qv["id"]
	if !ok {
		for k := range h.people {
			ids = append(ids, k)
		}
	}

	books := make([]*Person, 0, len(ids))
	for _, id := range ids {
		if b, ok := h.people[id]; ok {
			books = append(books, &b)
		}
	}

	err := jw.Encode(books)
	if err != nil {
		slog.ErrorContext(r.Context(), "error writing response", "error", err)
	}
}
