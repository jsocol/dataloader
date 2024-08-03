package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/jsocol/shutdown"

	"github.com/jsocol/rest-data-loader/examples/graphql-complete/handlers/books"
	"github.com/jsocol/rest-data-loader/examples/graphql-complete/handlers/people"
	"github.com/jsocol/rest-data-loader/examples/graphql-complete/shared/middleware"
)

const (
	defaultBindAddr = "127.0.0.1"
	defaultPort     = "9090"
	defaultLogLevel = "INFO"
)

func main() {
	var bindAddr string
	var port string
	var levelName string
	flag.StringVar(&bindAddr, "bind-addr", defaultBindAddr, "interface to bind to")
	flag.StringVar(&port, "port", defaultPort, "port to bind to")
	flag.StringVar(&levelName, "log-level", defaultLogLevel, "change the log level")
	flag.Parse()

	var level slog.Level
	if err := level.UnmarshalText([]byte(levelName)); err != nil {
		slog.Warn("could not parse log level", "level", levelName)
	}
	slog.SetLogLoggerLevel(level)

	srvAddr := bindAddr + ":" + port

	mux := http.NewServeMux()
	mux.Handle("/books", books.New())
	mux.Handle("/people", people.New())

	srv := http.Server{
		Handler: middleware.WithLogger(mux),
		Addr:    srvAddr,
	}

	shutdown.Listen(func(ctx context.Context) error {
		return srv.Shutdown(ctx)
	})

	slog.Info(fmt.Sprintf("connect to http://%s/ for resource server", srvAddr))
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		slog.Error("server error", "error", err)
	}
}
