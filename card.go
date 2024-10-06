package main

import (
	"encoding/json"
	"fmt"
)

type Card struct {
	Rank  int    `json:"rank"`
	Suit  int    `json:"suit"`
	Code  string `json:"code"`
	Image string `json:"image"`
}

func (c Card) MarshalJSON() ([]byte, error) {
	suitMap := map[int]string{
		1: "h",
		2: "d",
		3: "c",
		4: "s",
	}

	rankMap := map[int]string{
		0:  "x",
		11: "j",
		12: "q",
		13: "k",
	}

	if rankMap[c.Rank] == "" {
		rankMap[c.Rank] = fmt.Sprintf("%d", c.Rank)
	}

	code := fmt.Sprintf("%s%s", rankMap[c.Rank], suitMap[c.Suit])
	image := fmt.Sprintf("/static/%s.png", code)

	rankValue := fmt.Sprintf("%d", c.Rank)
	if mappedRank, found := rankMap[c.Rank]; found {
		rankValue = mappedRank
	}

	type Alias Card

	return json.Marshal(&struct {
		Alias
		Rank  string `json:"rank"`
		Suit  string `json:"suit"`
		Code  string `json:"code"`
		Image string `json:"image"`
	}{
		Alias: Alias(c),
		Rank:  rankValue,
		Suit:  suitMap[c.Suit],
		Code:  code,
		Image: image,
	})
}
