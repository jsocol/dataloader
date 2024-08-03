package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"log/slog"
	"net"

	restdataloader "github.com/jsocol/rest-data-loader"
	"github.com/jsocol/shutdown"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	_ "modernc.org/sqlite"

	"github.com/jsocol/rest-data-loader/examples/grpc-resource-server/fetcher"
	"github.com/jsocol/rest-data-loader/examples/grpc-resource-server/proto"
	"github.com/jsocol/rest-data-loader/examples/grpc-resource-server/server"
)

const (
	defaultBindAddr = "127.0.0.1"
	defaultPort     = "50051"
	defaultLogLevel = "INFO"
	defaultDBFile   = "./database.sqlite"
	defaultSeed     = false
)

func main() {
	var bindAddr string
	var port string
	var dbFile string
	var seed bool
	var levelName string
	flag.StringVar(&bindAddr, "bind-addr", defaultBindAddr, "interface to bind to")
	flag.StringVar(&port, "port", defaultPort, "port to bind to")
	flag.StringVar(&dbFile, "db-file", defaultDBFile, "path to a database file")
	flag.StringVar(&levelName, "log-level", defaultLogLevel, "change the log level")
	flag.BoolVar(&seed, "seed", defaultSeed, "populate the database at startup")
	flag.Parse()

	var level slog.Level
	if err := level.UnmarshalText([]byte(levelName)); err != nil {
		slog.Warn("could not parse log level", "level", levelName)
	}
	slog.SetLogLoggerLevel(level)

	srvAddr := bindAddr + ":" + port
	lis, err := net.Listen("tcp", srvAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}

	if err = fetcher.EnsureDBSchema(db); err != nil {
		log.Fatalf("failed to ensure database schema: %v", err)
	}

	if seed {
		err = fetcher.SeedDB(db)
		if err != nil {
			log.Fatalf("failed to seed database: %v", err)
		}
	}

	bookFetcher := fetcher.New(db)

	s := grpc.NewServer(grpc.ChainUnaryInterceptor())
	srv := &server.Server{
		Books: restdataloader.New(bookFetcher.Fetch),
	}
	proto.RegisterBookServiceServer(s, srv)

	reflection.Register(s)

	shutdown.Listen(func(context.Context) error {
		s.GracefulStop()
		return nil
	})

	slog.Info("server starting", "address", srvAddr)
	if err := s.Serve(lis); err != nil {
		slog.Error("error running server", "error", err)
	}
}
