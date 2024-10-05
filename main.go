package main

import (
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func deckRouter(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if parts[1] == "new" {
		newDeckHandler(w, r)
	} else if deckExists(parts[1]) {
		deckIdHandler(w, r)
	} else {
		http.Error(w, "Invalid request", http.StatusBadRequest)
	}
}

func deckIdHandler(w http.ResponseWriter, r *http.Request) {

}

func deckExists(id string) bool {
	return false
}

func newDeckHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := strings.Split(path, "/")
	includeJokers := false
	nbrPack := 1

	if len(parts) == 2 {
		createNewDeck(nbrPack, includeJokers)
	} else if _, err := strconv.Atoi(parts[3]); err == nil {
		nbrPack, _ = strconv.Atoi(parts[3])
		if len(parts) == 4 {
			switch parts[3] {
			case "includeJokers":
				includeJokers = true
			default:
				http.Error(w, "Invalid request", http.StatusBadRequest)
			}
		}
		createNewDeck(nbrPack, includeJokers)
	}
}

func createNewDeck(nbrPack int, includeJokers bool) {

}

func main() {
	var err error
	db, err = initDB()
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/deck", deckRouter)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	fmt.Println("Server running on port 8080")
	http.ListenAndServe(":8080", nil)
}
