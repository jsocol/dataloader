package books

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/jsocol/dataloader/examples/graphql-complete/fetchers"
	"github.com/jsocol/dataloader/examples/graphql-complete/graph/model"
)

// define some internal types for books.fetcher

type book struct {
	ID      string
	Title   string
	Authors []string
}

type doer interface {
	Do(*http.Request) (*http.Response, error)
}

type fetcher struct {
	baseAddr string
	client   doer
}

// New creates a new fetcher struct with a Fetch function that can be passed to
// restdataloader.New.
func New(addr string) *fetcher {
	u := fetchers.MustParse(addr)
	u.Path = "books"

	return &fetcher{
		baseAddr: u.String(),
		client:   http.DefaultClient,
	}
}

// Fetch implements the restdataloader.Fetcher interface, pulling a set of
// resources by ID from the resource server and converting them into the
// internal (in this case GraphQL) types.
func (f *fetcher) Fetch(ids []string) (map[string]*model.Book, error) {
	ctx := context.TODO()

	slog.InfoContext(ctx, "fetching books", "ids", ids)

	qv := url.Values{}
	for _, id := range ids {
		qv.Add("id", id)
	}
	u := fetchers.MustParse(f.baseAddr)
	u.RawQuery = qv.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}

	bodyReader := io.LimitReader(resp.Body, 4<<20)
	defer resp.Body.Close()

	jr := json.NewDecoder(bodyReader)
	books := make([]book, 0, len(ids))
	err = jr.Decode(&books)
	if err != nil {
		return nil, err
	}

	ret := make(map[string]*model.Book, len(books))
	for _, b := range books {
		ret[b.ID] = &model.Book{
			ID:        b.ID,
			Title:     b.Title,
			AuthorIDs: b.Authors,
		}
	}

	return ret, nil
}
