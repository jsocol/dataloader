package fetchers

import "net/url"

func MustParse(addr string) *url.URL {
	u, err := url.Parse(addr)
	if err != nil {
		panic(err)
	}
	return u
}
