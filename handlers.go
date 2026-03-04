package main

import (
	"fmt"
	"net"
	"net/url"
	"database/sql"
	"bufio"
	"strings"
	"io"

	_ "modernc.org/sqlite"	
)


func HandleConnection(conn net.Conn, db *sql.DB) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	requestLine, err := reader.ReadString('\n')
	if err != nil {
		return
	}

	parts := strings.Fields(requestLine)
	if len(parts) < 2 {
		return
	}

	method := parts[0]
	URL := parts[1]
	u, _ := url.Parse(URL)
	path := u.Path
	query:=u.Query()

	switch {
	case method == "GET" && path == "/":
		handleIndex(conn, db)

	case method == "GET" && path == "/create":
		handleAddForm(conn)

	case method == "POST" && path == "/create_series":
		handleAddSeries(conn, reader, db)

	case method == "POST" && path == "/update":
		handleUpdate(conn, db, query)
	default:
		handleNotFound(conn)
	}
}

func handleIndex(conn net.Conn, db *sql.DB) {
	rows, err := db.Query("SELECT id, name, current_episode, total_episodes FROM series")
	if err != nil{
		fmt.Println("Error en Queary:", err)
		handleNotFound(conn)
		return
	}
	defer rows.Close()
	var id int
	var name string
	var current_episode int
	var total_episodes int
	table_data := ""
		for rows.Next(){
			rows.Scan(&id, &name, &current_episode, &total_episodes)
			line := fmt.Sprintf("<tr><td>%d</td><td>%s</td><td>%d</td><td>%d</td><td><button onclick=\"nextEpisode(%d)\">+1</button></td></tr>", id, name, current_episode, total_episodes, id)
			fmt.Println(id, name, current_episode, total_episodes)
			table_data += line
			
		}

	html := `<html>
	<head>
	</head>
	<body>
	<script>
	async function nextEpisode(id){
	const url = "/update?id=" + id;
	const response = await fetch(url, {method: "POST"})
	location.reload();
	}
	</script>
	
	<table border="3" cellpadding="10" align="center" cellspacing="5">
	<caption>Mi lista de series</caption>
	<tr bgcolor="lightgray">
	<th>ID de la serie</th>
	<th>Nombre de la serie</th>
	<th>Episodio en el que voy</th>
	<th>Episodios totales</th>
	<th>Añadir serie</th>
	</tr>`
	html += table_data
	html += "</table><script>alert('Estas viendo muy buenas series :)')</script>"
	html += "<a href='./create'>Añadir nueva serie</a></body></html>"

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

func handleAddForm(conn net.Conn) {

	html := `
	<html>
	<body>
	<h2>Agregar nueva serie</h2>
	<form method="POST" action="/create_series">
	Nombre: <input type="text" name="name"><br>
	Episodio actual: <input type="number" name="current"><br>
	Episodios totales: <input type="number" name="total"><br>
	<input type="submit" value="Agregar">
	<a href="/">Volver al track de series></a>
	</form>
	</body>
	</html>`

	response := fmt.Sprintf(
		"HTTP/1.1 200 OK\r\n"+
			"Content-Type: text/html\r\n"+
			"Content-Length: %d\r\n\r\n%s",
		len(html),
		html,
	)

	conn.Write([]byte(response))
}

func handleAddSeries(conn net.Conn, reader *bufio.Reader, db *sql.DB) {
	for {
		line, _ := reader.ReadString('\n')
		if line == "\r\n" {
			break
		}
	}

	bodyBytes, _ := io.ReadAll(reader)
	body := string(bodyBytes)

	values, _ := url.ParseQuery(body)

	name := values.Get("name")
	current := values.Get("current")
	total := values.Get("total")

	db.Exec("INSERT INTO series (name, current_episode, total_episodes) VALUES (?, ?, ?)",
		name, current, total)

	response := "HTTP/1.1 303 See Other\r\nLocation: /\r\n\r\n"
	conn.Write([]byte(response))
}

func handleNotFound(conn net.Conn) {
	body := "<h1>404 - Página no encontrada</h1>"

	response := fmt.Sprintf(
		"HTTP/1.1 404 Not Found\r\n"+
			"Content-Type: text/html\r\n"+
			"Content-Length: %d\r\n\r\n%s",
		len(body),
		body,
	)

	conn.Write([]byte(response))
}

func handleUpdate(conn net.Conn, db *sql.DB, query url.Values){
	id := query.Get("id")
	_, err := db.Exec(`UPDATE series
	SET current_episode = current_episode + 1
	WHERE id = ? AND current_episode < total_episodes`, id)

	if err !=nil {
		fmt.Print("Error en update", err)
	}

	response:= "HTTP/1.1 200 OK \r\n" +
	"Content-Type: text/plain\r\n\r\n" + "ok"	

	conn.Write([]byte(response))
}
	