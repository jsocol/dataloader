package graph

import "github.com/jsocol/dataloader/examples/graphql-complete/graph/model"

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type personLoader interface {
	Load(string) (*model.Person, error)
}

type bookLoader interface {
	Load(string) (*model.Book, error)
}

type Resolver struct {
	People personLoader
	Books  bookLoader
}
