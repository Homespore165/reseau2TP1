package main

type Card struct {
	Rank  int    `json:"rank"`
	Suit  string `json:"suit"`
	Code  string `json:"code"`
	Image string `json:"image"`
}
