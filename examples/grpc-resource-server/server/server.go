package server

import (
	"context"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/jsocol/dataloader"
	"github.com/jsocol/dataloader/examples/grpc-resource-server/proto"
)

type bookLoader interface {
	Load(string) (*proto.Book, error)
}

type Server struct {
	proto.UnimplementedBookServiceServer
	Books bookLoader
}

func (s *Server) GetBook(ctx context.Context, in *proto.GetBookRequest) (*proto.Book, error) {
	book, err := s.Books.Load(in.Id)
	if err != nil {
		if err == dataloader.NotFound {
			return nil, status.Errorf(codes.NotFound, "book not found: %s", in.Id)
		}
		slog.ErrorContext(ctx, "error loading book", "book", in.Id, "error", err)
		return nil, status.Errorf(codes.Internal, "unknown error")
	}
	return book, nil
}
