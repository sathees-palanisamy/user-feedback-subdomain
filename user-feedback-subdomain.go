package main

import (
	"crypto/aes"
	"crypto/cipher"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/dgrijalva/jwt-go"
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

var (
	db            *sql.DB
	initialVector = "1010101010101010"
	passphrase    = []byte{0x31, 0x31, 0x31, 0x31, 0x31, 0x31, 0x31, 0x31, 0x31, 0x31, 0x31, 0x31, 0x31, 0x31, 0x31, 0x31, 0x31, 0x31, 0x31, 0x31, 0x31, 0x31, 0x31, 0x31, 0x31, 0x31, 0x31, 0x31, 0x31, 0x31, 0x31, 0x31}
)

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

	router.HandleFunc("/v1/feedback/createFeedback", AuthMiddleWare(insertHandler)).Methods("POST")
	router.HandleFunc("/v1/feedback/deleteFeedback", AuthMiddleWare(deleteHandler)).Methods("POST")
	router.HandleFunc("/v1/feedback/listFeedback", AuthMiddleWare(selectHandler)).Methods("GET")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal(err)
	}

}

func insertHandler(w http.ResponseWriter, r *http.Request) {

	// Controller - Folder

	fmt.Println("***********************")
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

	encryptedData, _ := base64.StdEncoding.DecodeString(indiSuggestReg.Detail)
	fmt.Println("encryptedData:", encryptedData)
	decryptedText := AESDecrypt(encryptedData, []byte(passphrase))
	fmt.Println("decryptedText:", decryptedText)

	queryDet := "insert into userfeedback (email, detail, date) values($1, $2, $3) RETURNING id;"

	err1 := db.QueryRow(queryDet, indiSuggestReg.Email, decryptedText, indiSuggestReg.Date).Scan(&indiSuggestRes.Id)

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

	fmt.Println("***********************")
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

	fmt.Println("***********************")
	fmt.Println("I am in Select Handler")

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

func AuthMiddleWare(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		fmt.Println("I am in Auth MiddleWare")

		/* Basic Auth
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Basic ") {
			log.Print("Invalid authorization:", auth)
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		up, _ := base64.StdEncoding.DecodeString(auth[6:])
		fmt.Println("up:", string(up))

		if string(up) != "123456:usrpass2" {
			log.Print("invalid username:password:", string(up))
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		*/

		/* JWT Validation */
		var errorObject Error
		authHeader := r.Header.Get("Authorization")
		bearerToken := strings.Split(authHeader, " ")

		fmt.Println("bearerToken:", bearerToken)

		if len(bearerToken) == 2 {
			authToken := bearerToken[1]

			token, error := jwt.Parse(authToken, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("There was an error")
				}

				return []byte("mysecrets"), nil
			})

			fmt.Println("error:", error)
			if error != nil {
				errorObject.Message = error.Error()
				RespondWithError(w, http.StatusUnauthorized, errorObject)
				return
			}

			fmt.Println("token:", token)
			if token.Valid {
				next.ServeHTTP(w, r)
			} else {
				errorObject.Message = error.Error()
				RespondWithError(w, http.StatusUnauthorized, errorObject)
				return
			}
		} else {
			errorObject.Message = "Invalid token."
			RespondWithError(w, http.StatusUnauthorized, errorObject)
			return
		}

	})
}

func RespondWithError(w http.ResponseWriter, status int, error Error) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(error)
}

func AESDecrypt(crypt []byte, key []byte) []byte {
	block, err := aes.NewCipher(key)
	if err != nil {
		fmt.Println("key error1", err)
	}
	if len(crypt) == 0 {
		fmt.Println("plain content empty")
	}
	ecb := cipher.NewCBCDecrypter(block, []byte(initialVector))
	decrypted := make([]byte, len(crypt))
	ecb.CryptBlocks(decrypted, crypt)
	return PKCS5Trimming(decrypted)
}

func PKCS5Trimming(encrypt []byte) []byte {
	padding := encrypt[len(encrypt)-1]
	return encrypt[:len(encrypt)-int(padding)]
}
