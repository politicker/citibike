package main

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/apoliticker/citibike/db"
	_ "github.com/lib/pq"
)

var databaseURL string
var queries *db.Queries

func init() {
	databaseURL = os.Getenv("DATABASE_URL")
	fmt.Println("connecting to: ", databaseURL)

	if os.Getenv("GO_ENV") == "production" {
		databaseURL = fmt.Sprintf("%s?sslmode=disable", databaseURL)
	}

	database, err := sql.Open("postgres", databaseURL)
	if err != nil {
		panic("failed to connect to database")
	}

	queries = db.New(database)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	// TODO: Pass cancellable context to poller and server

	poller := NewPoller(queries, 1*time.Minute)
	go poller.Start()

	srv := NewServer(port, queries)
	srv.Start()
}
