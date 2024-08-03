# Example: GraphQL and Resource Servers

In this example, restdataloader is used to reduce the risk of duplicative or N+1
queries between a GraphQL server (built from [gqlgen][gqlgen]) and a toy
resource server that implements list endpoints that support multiple IDs as
inputs.

Other than data storage, which is simple in-memory data populated by JSON files,
I have tried to make this example as "real" as possible. There are comments
scattered throughout the code to help understand what's going on.

## Run the Example

After cloning this repo, open two terminal windows in this directory. In one
window, run the resource server:

```sh
go run cmd/resource-server/main.go
```

In the other terminal, run the GraphQL server:

```sh
go run cmd/graph-server/main.go
```

You can add the `-help` flag to see the available flags for either command.

If you're using the defaults, open a browser to `http://localhost:8080/` and you
should be presented with the GraphiQL playground. Introspection is on so you can
use the playground to explore the queries available.

### Collapsing Requests

Naive implementations of GraphQL resolvers can end up making separate requests
to resource servers for each, potentially nested or duplicated, resource
returned by the query. To see how restdataloader collapses these individual
requests, try this query:

```graphql
{
  book(id:"urn:isbn:978-1942788331"){
    title
    authors {
      name
    }
  }
}
```

The resulting book has three authors, and we're fetching their names. However,
if you look at the logging output in the resource server terminal, you will see
two requests:

```http
GET /books?id=urn%3Aisbn%3A978-1942788331
GET /people?id=urn%3Aperson%3Agene-kim&id=urn%3Aperson%3Anicole-forsgren&id=urn%3Aperson%3Ajez-humble
```

The first request, to the "books" resource server, fetches a single result by
its ID. The second request, to the "people" resource server, fetches all three
authors in one request.

### Collapsing Across Queries

One risk of GraphQL is that a single graph query can result in several backend
lookups due to the structure of the query. In this example, constructing a query
that looks up two books results in two concurrent calls to the Book resolver,
which typically means two requests to the book resource server. Using
restdataloader, we're able to collapse concurrent requests _and_ the nested
resource requests.

```graphql
{
  accel: book(id: "urn:isbn:978-1942788331") {
    title
    authors {
      id
      name
    }
  }
  staff: book(id: "urn:isbn:978-1736417911") {
    id
    title
    authors {
      name
    }
  }
}
```

The logs again show two requests, one to each resource server:

```http
GET /books?id=urn%3Aisbn%3A978-1942788331&id=urn%3Aisbn%3A978-1736417911
GET /people?id=urn%3Aperson%3Anicole-forsgren&id=urn%3Aperson%3Ajez-humble&id=urn%3Aperson%3Agene-kim&id=urn%3Aperson%3Atanya-reilly&id=urn%3Aperson%3Awill-larson
```

restdataloader turns up to 7 separate requests (2 for books and 5 for authors)
into one request per resource type.

### Deduplication

A consequence of how restdataloader collapses requests is that individual
resources can be reused across concurrent requests (even potentially across
concurrent users)! In this example, we have three books with a total of two
authors across them.

```graphql
{
  execs: book(id: "urn:isbn:978-1098149482") {
    title
    authors {
      id
      name
    }
  }
  staff: book(id: "urn:isbn:978-1736417911") {
    id
    title
    authors {
      name
    }
  }
  staffPath: book(id:"urn:isbn:978-1098118730") {
    title
    authorIDs
    authors {
      name
    }
  }
}
```

In the logs we again see just two requests, and in particular each author is
only requested once and used to satisfy multiple concurrent lookups:

```http
GET /books?id=urn%3Aisbn%3A978-1098149482&id=urn%3Aisbn%3A978-1736417911&id=urn%3Aisbn%3A978-1098118730
GET /people?id=urn%3Aperson%3Atanya-reilly&id=urn%3Aperson%3Awill-larson
```

### Partial Failures

GraphQL is able to return partial errors for a given query if there are
failures with some resources but not all. Similarly, restdataloader can handle
the case of missing results for a given lookup batch.

```graphql
{
  fake: book(id: "urn:isbn:978-notreal") {
    title
    authors {
      name
    }
  }
  phb: book(id: "urn:isbn:978-0786965601") {
    id
    title
    authorIDs
    authors {
      name
    }
  }
}
```

## Directory Walkthrough

This example contains the source code for both the GraphQL and resource servers,
but only code in `shared/` is used by both.

### `shared/middleware/`

Since both servers use HTTP, they both use the same request logging middleware,
defined here.

### GraphQL Server

The GraphQL server starts in `cmd/graph-server/main.go` and largely follows the
default layout of a [gqlgen][gqlgen] project.

- `graph/` contains the generated code and resolver implementations. The **most
  important thing to note** is that the resolvers in `graph/schema.resolvers.go`
  use restdataloader to load data _in parallel_. Serial requests cannot be
  collapsed.
- `schema/` contains the GraphQL schema definition used to generate the server
  code. If you make any changes in `schema`, re-run `go run
  github.com/99designs/gqlgen generate` in the root of this example.
- `fetchers/` contains the required fetching implementation that, given a set of
  unique IDs provided by restdataloader, makes a single request to the
  appropriate resource server.
- `tools.go` ensures that the correct version of gqlgen is used.
- `gqlgen.yml` is the configuration for gqlgen.

### Resource Server

The Resource server is an REST-ish HTTP server that exposes two "list"
endpoints, one at `/books` and one at `/people`, that support filtering by
multiple IDs.

The Resource server starts in `cmd/resource-server/main.go` and is primarily
implemented within `handlers/`.

Both `handlers/books` and `handlers/people` follow a similar pattern:

- `New()` loads a set of data from JSON files stored in the directory and
  embedded into the binary. It constructs a local key-value store and returns an
  `http.Handler`.
- `Handler.ServeHTTP` looks at any `id=` query string parameters, pulls the
  requisite IDs out of the key-value store, and builds a JSON response.

[gqlgen]: https://gqlgen.com
