package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"stakeholders-service/model"

	"github.com/gorilla/mux"
)

func test(resp http.ResponseWriter, req *http.Request) {
	user := model.User{
		Username: "kita",
		Name:     "Ivan",
		Surname:  "Novakovic",
	}

	json.NewEncoder(resp).Encode(user)
}

func handleReq(resp http.ResponseWriter, req *http.Request) {
	fmt.Println(req.Header)
	resp.Write([]byte("test"))
}

func handlePathReq(resp http.ResponseWriter, req *http.Request) {
	path := mux.Vars(req)["path"]
	resp.Write([]byte(path))
}

func main() {

	router := mux.NewRouter()

	router.HandleFunc("/", handleReq)
	router.HandleFunc("/user", test)
	router.HandleFunc("/{path}", handlePathReq).Methods("GET")

	log.Fatal(http.ListenAndServe(":8080", router))
}
