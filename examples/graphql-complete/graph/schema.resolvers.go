package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.49

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/jsocol/rest-data-loader/examples/graphql-complete/graph/model"
)

// Authors is the resolver for the authors field.
func (r *bookResolver) Authors(ctx context.Context, obj *model.Book) ([]*model.Person, error) {
	if len(obj.Authors) > 0 {
		return obj.Authors, nil
	}

	authors := make([]*model.Person, len(obj.AuthorIDs))

	// For performance reasons, we should try to fetch these Authors in
	// parallel. Normally this would create N queries, one for each author.
	// restdataloader collapses these requests, because they are made in rapid
	// succession.
	var wg sync.WaitGroup
	for i, id := range obj.AuthorIDs {
		wg.Add(1)
		go func(i int, id string) {
			defer wg.Done()

			author, err := r.Resolver.People.Load(id)
			if err != nil {
				slog.ErrorContext(ctx, "error loading author", "id", id, "error", err)
			}

			authors[i] = author
		}(i, id)
	}

	wg.Wait()

	return authors, nil
}

// Editors is the resolver for the editors field.
func (r *bookResolver) Editors(ctx context.Context, obj *model.Book) ([]*model.Person, error) {
	panic(fmt.Errorf("not implemented: Editors - editors"))
}

// Person is the resolver for the person field.
func (r *queryResolver) Person(ctx context.Context, id string) (*model.Person, error) {
	return r.People.Load(id)
}

// Book is the resolver for the book field.
func (r *queryResolver) Book(ctx context.Context, id string) (*model.Book, error) {
	return r.Books.Load(id)
}

// Book returns BookResolver implementation.
func (r *Resolver) Book() BookResolver { return &bookResolver{r} }

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

type (
	bookResolver  struct{ *Resolver }
	queryResolver struct{ *Resolver }
)
