package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/lib/pq"
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

var db *sql.DB

func main() {

	// driver - Folder
	pgUrl, err := pq.ParseURL("postgres://??")

	if err != nil {
		log.Fatal(err)
	}

	db, err = sql.Open("postgres", pgUrl)

	if err != nil {
		log.Fatal(err)
	}

	// routes - Folder

	router := mux.NewRouter()

	router.HandleFunc("/v1/feedback/createFeedback", insertHandler).Methods("POST")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal(err)
	}

}

func insertHandler(w http.ResponseWriter, r *http.Request) {

	// Controller - Folder

	fmt.Println("I am in Insert Handler")

	var indiSuggestReg SuggestionRequest
	var indiSuggestRes SuggestionResponse
	var error Error

	fmt.Println("indiSuggestReg:", indiSuggestReg)
	fmt.Println("r.Body:", r.Body)

	err := json.NewDecoder(r.Body).Decode(&indiSuggestReg)

	if err != nil {
		error.Message = "Bad data"
		responseWithError(w, http.StatusBadRequest, error)
		return
	}

	if indiSuggestReg.Email == "" {
		error.Message = "Email ID should not be empty"
		responseWithError(w, http.StatusBadRequest, error)
		return
	}

	// services & Domain - folder
	fmt.Println("indiSuggestReg:", indiSuggestReg)

	queryDet := "insert into userfeedback (email, detail, date) values($1, $2, $3) RETURNING id;"

	err1 := db.QueryRow(queryDet, indiSuggestReg.Email, indiSuggestReg.Detail, indiSuggestReg.Date).Scan(&indiSuggestRes.Id)

	if err1 != nil {
		log.Fatal(err1)
	}

	indiSuggestRes.Email = indiSuggestReg.Email
	indiSuggestRes.Detail = indiSuggestReg.Detail
	indiSuggestRes.Date = indiSuggestRes.Date
	//	indiSuggestRes.Id = "1"

	w.Header().Set("content-type", "application/json")

	json.NewEncoder(w).Encode(indiSuggestRes)
}

func responseWithError(w http.ResponseWriter, status int, error Error) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(error)
}
