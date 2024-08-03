# restdataloader

`restdataloader` helps avoid N+1 lookups by collapsing a set of individual
lookups by key into a single list lookup. For example, when loading several
nested resources of the same type, or when loading nested resources in a list.
It is inspired by the "Data Loader" pattern in GraphQL, but intended to be used
in a wider range of situations.

## Quickstart

```go
type Author struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// all values in ids will be unique
func fetchAuthors(ids []string) (map[string]Author, error) {
	var qv url.Values
	qv.Add("id", strings.Join(ids, ","))

	u := url.Parse("https://resource.server/authors")
	u.RawQuery = qv.Encode()

	res, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	authors := make([]Author, 0, len(ids))
	err := json.Unmarshal(body, &authors)
	if err != nil {
		return nil, err
	}

	ret := make(map[string]Author, len(authors))

	for _, author := range authors {
		ret[author.ID] = author
	}

	return ret, nil
}

func OnePlusN() {
	authorLoader := restdataloader.New(fetchAuthors)

	posts := GetRecentPosts()

	var wg sync.WaitGroup
	// Fetch all authors, typically an N+1 problem with duplication
	// We'll do it in parallel for speed!
	for _, post := range posts {
		wg.Add(1)
		go func() {
			defer wg.Done()
			auth, _ := authorLoader.Load(post.AuthorID)
			post.Author = auth
		}()
	}

	wg.Wait()

	for _, post := range posts {
		fmt.Printf("%s by %s\n", post.Title, post.Author.Name)
	}
}
```
