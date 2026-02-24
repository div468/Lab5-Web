package main

import (
	"fmt"
	"log"
	"net"
	"database/sql"

	_ "modernc.org/sqlite"
)

func handleClient(conn net.Conn, db *sql.DB) {
	defer conn.Close()
	rows, _ := db.Query("SELECT id, name, current_episode, total_episodes FROM series")
	defer rows.Close()

	var id int
	var name string
	var current_episode int
	var total_episodes int
	table_data := ""
		for rows.Next(){
			rows.Scan(&id, &name, &current_episode, &total_episodes)
			line := fmt.Sprintf("<tr><td>%d</td><td>%s</td><td>%d</td><td>%d</td></tr>", id, name, current_episode, total_episodes)
			fmt.Println(id, name, current_episode, total_episodes)
			table_data += line
			
		}

	html := `<html>
	<head></head>
	<body>
	<table border="3" cellpadding="10" align="center" cellspacing="5">
	<caption>Mi lista de series</caption>
	<tr bgcolor="lightgray">
	<th>ID de la serie</th>
	<th>Nombre de la serie</th>
	<th>Episodio en el que voy</th>
	<th>Episodios totales</th>
	</tr>`
	html += table_data
	html += "</table></body></html>"
	

	body := html

	response := fmt.Sprintf(
		"HTTP/1.1 200 OK\r\n"+
			"Content-Type: text/html\r\n"+
			"Content-Length: %d\r\n"+
			"\r\n"+
			"%s",
		len(body),
		body,
	)

	fmt.Println(response)
	conn.Write([]byte(response))

	}
func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	db, _ := sql.Open("sqlite", "file:series.db")
	defer db.Close()

	log.Println("Listening on port 8080...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}

		go handleClient(conn, db)
	}
}

