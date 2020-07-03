package main

import (

	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"

)

type SuggestionRequest struct {
	Id     string `json:"id"`
	Email  string `json:"email"`
	Detail string `json:"detail"`
	Date   string `json:"date"`
}

type SuggestionResponse struct {
	Id     string `json:"id"`
	Email  string `json:"email"`
	Detail string `json:"detail"`
	Date   string `json:"date"`
}

type Error struct {
	Message string `json:"message"`
}

func main() {

	router := mux.NewRouter()

	router.HandleFunc("/v1/feedback/createFeedback", insertHandler).Methods("POST")

	// [START setting_port]
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal(err)
	}
	// [END setting_port]

}

func insertHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Println("I am in Insert Handler")

	var indiSuggestReg SuggestionRequest
    var indiSuggestRes SuggestionResponse
    var error Error

    err := json.NewDecoder(r.Body).Decode(&indiSuggestReg)

    if err != nil {
        error.Message = "Bad data"
        responseWithError(w, http.StatusBadRequest, error)
        return
    }

    indiSuggestRes.Email = indiSuggestReg.Email
    indiSuggestRes.Detail = indiSuggestReg.Detail
    indiSuggestRes.Date = indiSuggestRes.Date
    indiSuggestRes.Id = "1"

    w.Header().Set("content-type", "application/json")

    json.NewEncoder(w).Encode(indiSuggestRes)
}


func responseWithError(w http.ResponseWriter, status int, error Error) {
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(error)
}