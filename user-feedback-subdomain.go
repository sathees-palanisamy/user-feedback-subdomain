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
	pgUrl, err := pq.ParseURL("postgres://xsswjxse:lnGml1jOsTDYha2yjV0o3UZz1GJnK0Ie@rogue.db.elephantsql.com:5432/xsswjxse")

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
	router.HandleFunc("/v1/feedback/deleteFeedback", deleteHandler).Methods("POST")
	router.HandleFunc("/v1/feedback/listFeedback", selectHandler).Methods("GET")

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
	indiSuggestRes.Date = indiSuggestReg.Date
	//	indiSuggestRes.Id = "1"

	w.Header().Set("content-type", "application/json")

	json.NewEncoder(w).Encode(indiSuggestRes)
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {

	// Controller - Folder

	fmt.Println("I am in Delete Handler")

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

	if indiSuggestReg.Id == "" {
		error.Message = "ID should not be empty"
		responseWithError(w, http.StatusBadRequest, error)
		return
	}

	// services & Domain - folder
	fmt.Println("indiSuggestReg:", indiSuggestReg)

	row := db.QueryRow("select * from userfeedback where id=$1", indiSuggestReg.Id)

	err = row.Scan(&indiSuggestRes.Id, &indiSuggestRes.Email, &indiSuggestRes.Detail, &indiSuggestRes.Date)

	if err != nil {
		if err == sql.ErrNoRows {
			error.Message = "User does not exist"
			responseWithError(w, http.StatusBadRequest, error)
			return
		} else {
			log.Fatal(err)
			error.Message = "Server error"
			responseWithError(w, http.StatusInternalServerError, error)
			return
		}

	}

	stmt := "delete from userfeedback where id=$1;"

	res, err := db.Exec(stmt, indiSuggestReg.Id)

	if err != nil {
		error.Message = "Server error"
		responseWithError(w, http.StatusInternalServerError, error)
		return
	}

	fmt.Println("res:", res)
	fmt.Println("err:", err)

	w.Header().Set("content-type", "application/json")

	json.NewEncoder(w).Encode(indiSuggestRes)
}

func selectHandler(w http.ResponseWriter, r *http.Request) {

	var feedbackList []SuggestionResponse
	var error Error

	rows, err := db.Query("select * from userfeedback")

	if err != nil {
		panic(err)
	}

	defer rows.Close()

	for rows.Next() {
		var feedbackResponse SuggestionResponse

		if err := rows.Scan(&feedbackResponse.Id, &feedbackResponse.Email, &feedbackResponse.Detail, &feedbackResponse.Date); err != nil {
			log.Fatal(err)
		}

		fmt.Println("feedbackResponse:", feedbackResponse)

		feedbackList = append(feedbackList, feedbackResponse)
	}

	err = rows.Err()

	if err != nil {
		error.Message = "Server error"
		responseWithError(w, http.StatusInternalServerError, error)
		return
	}

	fmt.Println("feedbackList:", feedbackList)

	w.Header().Set("content-type", "application/json")

	json.NewEncoder(w).Encode(feedbackList)

}

func responseWithError(w http.ResponseWriter, status int, error Error) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(error)
}
