package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
)

type response struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	Message   string    `json:"message"`
}

func main() {
	port := os.Getenv("port")

	helloHandler := func(w http.ResponseWriter, req *http.Request) {
		log.Println("hello-handler")

		resp := response{ID: uuid.New(), CreatedAt: time.Now(), Message: "Hello, world"}
		w.Header().Add("content-type", "application/json")
		json.NewEncoder(w).Encode(resp) //nolint: errcheck
	}

	http.HandleFunc("/hello", helloHandler)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
