package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/jsocol/dataloader/examples/grpc-resource-server/proto"
)

const defaultAddr = "localhost:50051"

type book struct {
	ID    string
	Title string
}

func main() {
	var addr string
	flag.StringVar(&addr, "server-addr", defaultAddr, "address of the server")
	flag.Parse()

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("could not create connection: %v", err)
	}
	defer conn.Close()

	c := proto.NewBookServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Several book IDs, with duplicates
	ids := []string{
		"urn:isbn:978-1098149482",
		"urn:isbn:978-1098118730",
		"urn:isbn:978-1491973899",
		"urn:isbn:978-0786965601",
		"urn:isbn:978-1098118730",
		"urn:isbn:978-1491973899",
		"urn:isbn:fake-book",
	}

	results := make([]book, len(ids))

	var wg sync.WaitGroup
	wg.Add(len(ids))

	for i, id := range ids {
		go func(i int, id string) {
			defer wg.Done()
			b, err := c.GetBook(ctx, &proto.GetBookRequest{Id: id})
			if err != nil {
				slog.WarnContext(ctx, "error fetching book", "book", id)
				return
			}
			results[i] = book{
				ID:    b.Id,
				Title: b.Title,
			}
		}(i, id)
	}

	wg.Wait()

	for _, b := range results {
		fmt.Printf("%s (%s)\n", b.Title, b.ID)
	}
}
