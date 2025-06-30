package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/ganglinwu/todoapp-backend-v1/mongostore"
	"github.com/ganglinwu/todoapp-backend-v1/postgres_store"
	"github.com/ganglinwu/todoapp-backend-v1/server"
)

func main() {
	addr := flag.String("addr", ":8080", "http address")
	datastore := flag.String("store", "mongo", "data store: mongo or postgres")
	mongoDSN := flag.String("mongoDSN", "", "mongoDB DSN")
	mongodbname := flag.String("mongoDBname", "", "mongoDB database name")
	mongocollectionname := flag.String("mongoCollection", "", "mongoDB collecton name")
	postgresDSN := flag.String("postgresDSN", "", "postgreSQL DSN")

	flag.Parse()

	handler := &server.TodoServer{}

	switch strings.ToLower(*datastore) {
	case "mongo":
		conn, err := mongostore.NewConnection(mongoDSN)
		if err != nil {
			log.Fatal("error initializing New mongo connection", err)
		}

		dbName, collName, err := mongostore.GetDBNameCollectionName(mongodbname, mongocollectionname)
		if err != nil {
			log.Fatal("error fetching mongo dbname and collection name: ", err)
		}

		store := &mongostore.MongoStore{}

		store.Conn = conn
		store.Collection = conn.Database(*dbName).Collection(*collName)

		handler = server.NewTodoServer(store)
	case "postgres":
		db, err := postgres_store.NewConnection(*postgresDSN)
		if err != nil {
			log.Fatal("error initializing New postgres connection: ", err)
		}

		// test connection
		err = db.Ping()
		if err != nil {
			log.Fatal("error sending PING to postgres DB: ", err)
		}
		newPostgresStore := &postgres_store.PostGresStore{DB: db}
		handler = server.NewTodoServer(newPostgresStore)

	default:
		log.Fatalf("the datastore %s, is not supported \n", *datastore)
	}
	s := http.Server{
		Addr:              *addr,
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
