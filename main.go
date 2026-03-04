package main

import (
	"log"
	"net"
	"database/sql"
)

var db *sql.DB
func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	db, err = sql.Open("sqlite", "file:series.db?_pragma=busy_timeout(5000)")
	if err != nil {
	log.Fatal(err)
}

db.SetMaxOpenConns(1)	
	defer db.Close()

	log.Println("Listening on port 8080...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}

		go HandleConnection(conn, db)
	}
}