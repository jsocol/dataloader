// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

type Book struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	AuthorIDs []string  `json:"authorIDs,omitempty"`
	Authors   []*Person `json:"authors,omitempty"`
	EditorIDs []string  `json:"editorIDs,omitempty"`
	Editors   []*Person `json:"editors,omitempty"`
}

type Person struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Query struct {
}