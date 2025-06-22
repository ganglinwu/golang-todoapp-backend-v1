package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ganglinwu/todoapp-backend-v1/mongostore"
	"github.com/ganglinwu/todoapp-backend-v1/server"
)

func main() {
	addr := flag.String("addr", ":8080", "http address")

	flag.Parse()

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

	s := http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 1 * time.Second,
		WriteTimeout:      2 * time.Second,
		// ErrorLog errLogger,
	}

	go func() {
		err := s.ListenAndServe()
		if err != nil {
			log.Fatal("failed to listen and serve. Reason:", err.Error())
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	signal.Notify(sigChan, syscall.SIGTERM)

	sig := <-sigChan
	log.Println("received terminate, shutting down gracefully. Signal received:", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	s.Shutdown(ctx)
}
