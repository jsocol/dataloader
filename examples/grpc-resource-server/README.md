# gRPC Resource Server

In this example, a gRPC server backed by a SQL database handles concurrent
requests from different clients, and is able to batch and deduplicate queries to
reduce load on the database.

The underlying gRPC service is defined in `proto/books.proto`. If you make
changes to that file, run `make`. This requires the Go gRPC [compiler
toolchain][1].

## Running the Example

Open two terminal windows to this directory. In the first window, start the
server:

```sh
go run cmd/grpc-server/main.go -log-level=DEBUG -seed
```

Note that the `-seed` argument should only be used once, or after erasing the
database.

Setting the log level to DEBUG will show you the exact SQL that is executed.

In the second window, run the client:

```sh
go run cmd/grpc-client/main.go
```

You should see a log line that looks like this:

```
DEBUG SELECT ids="[urn:isbn:978-1098118730 urn:isbn:fake-book urn:isbn:978-0786965601 urn:isbn:978-1491973899 urn:isbn:978-1098149482]" query="SELECT id, title FROM books WHERE id IN (?,?,?,?,?)" args="[urn:isbn:978-1098118730 urn:isbn:fake-book urn:isbn:978-0786965601 urn:isbn:978-1491973899 urn:isbn:978-1098149482]"
```

The order of the IDs and args may be different.

Both commands accept the `-help` flag to see how they can be configured.

## The Server

The server starts in `cmd/grpc-server/main.go`, but the implementation of the
gRPC server is in `server/server.go`, and the database is implemented in
`fetcher/fetcher.go`.

The server implements a single RPC, `GetBook`, which is a single-resource
lookup. This follows [Resource-Oriented Design][2] for gRPC and is a common
pattern for both gRPC and REST services.

All requests to this server are concurrent: they do not share anything and may
be initiated by different clients. restdataloader is able to collapse these
concurrent requests and deduplicate unique IDs before passing them into the
fetcher, which uses a SQL `SELECT ... WHERE id IN (?...)` query to gather
multiple results.

## The Client

The client starts in `cmd/grpc-client/main.go`. It simulates several clients by
making several parallel, independent requests, including multiple requests for
the same resource. Other than the protocol definitions in `proto/`, the client
is self-contained.

[1]: https://grpc.io/docs/languages/go/quickstart/
[2]: https://google.aip.dev/100
