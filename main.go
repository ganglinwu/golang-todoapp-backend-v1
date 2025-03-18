package main

import (
	"log"
	"net/http"

	"github.com/ganglinwu/todoapp-backend-v1/inmemorystore"
	"github.com/ganglinwu/todoapp-backend-v1/server"
)

func main() {
	store := &inmemorystore.InMemoryStore{Store: map[string]string{}}

	handler := &server.TodoServer{TodoStore: store}

	log.Fatal(http.ListenAndServe(":8080", handler))
}
