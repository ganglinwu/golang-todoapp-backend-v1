package main

import (
	"log"
	"net/http"

	"github.com/ganglinwu/todoapp-backend-v1/mongostore"
	"github.com/ganglinwu/todoapp-backend-v1/server"
)

func main() {
	conn, err := mongostore.NewConnection()
	if err != nil {
		log.Fatal(err)
	}

	dbName, collName, err := mongostore.GetDBNameCollectionName()
	if err != nil {
		log.Fatal(err)
	}

	store := &mongostore.MongoStore{}

	store.Conn = conn
	store.Collection = conn.Database(dbName).Collection(collName)

	handler := server.NewTodoServer(store)

	log.Fatal(http.ListenAndServe(":8080", handler))
}
