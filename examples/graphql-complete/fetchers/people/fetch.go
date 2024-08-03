package people

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/jsocol/rest-data-loader/examples/graphql-complete/fetchers"
	"github.com/jsocol/rest-data-loader/examples/graphql-complete/graph/model"
)

type person struct {
	ID   string
	Name string
}

type doer interface {
	Do(*http.Request) (*http.Response, error)
}

type fetcher struct {
	baseAddr string
	client   doer
}

func New(addr string) *fetcher {
	u := fetchers.MustParse(addr)
	u.Path = "people"

	return &fetcher{
		baseAddr: u.String(),
		client:   http.DefaultClient,
	}
}

func (f *fetcher) Fetch(ids []string) (map[string]*model.Person, error) {
	ctx := context.TODO()

	slog.DebugContext(ctx, "fetching people", "ids", ids)

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
	people := make([]person, 0, len(ids))
	err = jr.Decode(&people)
	if err != nil {
		return nil, err
	}

	ret := make(map[string]*model.Person, len(people))
	for _, p := range people {
		ret[p.ID] = &model.Person{
			ID:   p.ID,
			Name: p.Name,
		}
	}

	return ret, nil
}
