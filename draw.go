package main

type draw struct {
	DeckID    string `json:"deck_id"`
	Cards     []Card `json:"cards,omitempty"`
	Remaining int    `json:"remaining"`
	Error     string `json:"error,omitempty"`
}
