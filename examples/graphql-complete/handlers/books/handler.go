package books

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/jsocol/rest-data-loader/examples/graphql-complete/handlers"
)

type Book struct {
	ID      string   `json:"id"`
	Title   string   `json:"title"`
	Authors []string `json:"authors"`
	Editors []string `json:"editors"`
}

//go:embed books.json
var bookData []byte

type Handler struct {
	books map[string]Book
}

func New() *Handler {
	var rawBooks []Book
	decoder := json.NewDecoder(bytes.NewReader(bookData))
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&rawBooks)
	if err != nil {
		panic(err)
	}

	books := make(map[string]Book, len(rawBooks))

	for _, book := range rawBooks {
		slog.Debug("loading book", "book", book)
		books[book.ID] = book
	}

	return &Handler{
		books: books,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", handlers.ContentTypeJSON)
	jw := json.NewEncoder(w)

	qv := r.URL.Query()
	ids, ok := qv["id"]
	if !ok {
		for k := range h.books {
			ids = append(ids, k)
		}
	}

	books := make([]*Book, 0, len(ids))
	for _, id := range ids {
		if b, ok := h.books[id]; ok {
			books = append(books, &b)
		}
	}

	err := jw.Encode(books)
	if err != nil {
		slog.ErrorContext(r.Context(), "error writing response", "error", err)
	}
}
