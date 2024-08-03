package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	restdataloader "github.com/jsocol/rest-data-loader"
	"github.com/jsocol/shutdown"

	"github.com/jsocol/rest-data-loader/examples/graphql-complete/fetchers/books"
	"github.com/jsocol/rest-data-loader/examples/graphql-complete/fetchers/people"
	"github.com/jsocol/rest-data-loader/examples/graphql-complete/graph"
	"github.com/jsocol/rest-data-loader/examples/graphql-complete/shared/middleware"
)

const (
	defaultBindAddr     = "127.0.0.1"
	defaultPort         = "8080"
	defaultResourceAddr = "http://127.0.0.1:9090"
	defaultLogLevel     = "INFO"
)

func main() {
	var bindAddr string
	var port string
	var resourceAddr string
	var levelName string
	flag.StringVar(&bindAddr, "bind-addr", defaultBindAddr, "interface to bind to")
	flag.StringVar(&port, "port", defaultPort, "port to bind to")
	flag.StringVar(&resourceAddr, "resource-server-addr", defaultResourceAddr, "address of the resource server")
	flag.StringVar(&levelName, "log-level", defaultLogLevel, "change the log level")
	flag.Parse()

	var level slog.Level
	if err := level.UnmarshalText([]byte(levelName)); err != nil {
		slog.Warn("could not parse log level", "level", levelName)
	}
	slog.SetLogLoggerLevel(level)

	srvAddr := bindAddr + ":" + port

	peopleFetcher := people.New(resourceAddr)
	bookFetcher := books.New(resourceAddr)

	resolver := &graph.Resolver{
		People: restdataloader.New(peopleFetcher.Fetch),
		Books:  restdataloader.New(bookFetcher.Fetch),
	}

	gqlsrv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: resolver}))

	mux := http.NewServeMux()
	srv := http.Server{
		Handler: middleware.WithLogger(mux),
		Addr:    srvAddr,
	}

	mux.Handle("/", playground.Handler("GraphQL playground", "/query"))
	mux.Handle("/query", gqlsrv)

	shutdown.Listen(func(ctx context.Context) error {
		return srv.Shutdown(ctx)
	})

	slog.Info(fmt.Sprintf("connect to http://%s/ for GraphQL playground", srvAddr))
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		slog.Error("server error", "error", err)
	}
}
