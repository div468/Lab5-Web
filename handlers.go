package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"io"
	"net"
	"net/url"
	"strings"

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
	query := u.Query()

	switch {
	case method == "GET" && path == "/":
		handleIndex(conn, db, query)

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

func handleIndex(conn net.Conn, db *sql.DB, query url.Values) {
	pageStr := query.Get("page")
	page := 1
	if pageStr != "" {
		fmt.Sscan(pageStr, &page)
	}

	limit := 5
	offset := (page - 1) * limit
	var totalSeries int
	db.QueryRow("SELECT COUNT(*) from series").Scan(&totalSeries)
	totalPages := (totalSeries + limit - 1) / limit

	hasNext := page < totalPages
	hasPrev := page > 1
	rows, err := db.Query("SELECT id, name, current_episode, total_episodes FROM series LIMIT ? OFFSET ?", limit, offset)
	if err != nil {
		fmt.Println("Error en Queary:", err)
		handleNotFound(conn)
		return
	}
	prevPage := page - 1
	nextPage := page + 1
	defer rows.Close()
	var id int
	var name string
	var current_episode int
	var total_episodes int
	table_data := ""
	for rows.Next() {
		rows.Scan(&id, &name, &current_episode, &total_episodes)
		line := fmt.Sprintf(`<tr>
		<td>%d</td>
		<td>%s</td>
		<td id='episode-%d'>%d</td>
		<td>%d</td>
		<td><progress id="progress-%d" value=%d max=%d></progress></td>
		<td><button onclick="addEpisode(%d)">+1</button></td>
		</tr>`,
			id, name, id, current_episode, total_episodes, id, current_episode, total_episodes, id)

		fmt.Println(id, name, current_episode, total_episodes)
		table_data += line
	}

	html := `<html>
	<head>
	<meta charset="UTF-8">
	<title>Tracker de series</title>
	</head>
	<body>
	<script>

async function addEpisode(id){
	const url = "/update?id=" + id;
	const response = await fetch(url, {method: "POST"})
	const newValue = await response.text()

	const new_episode = document.getElementById("episode-" + id);
	new_episode.textContent = newValue

	const new_progress = document.getElementById("progress-" + id);
	new_progress.value = newValue
	}

	let sortDirection = {}

	function sortTable(column){

		const table = document.getElementById("seriesTable")
		const rows = Array.from(table.rows)

		sortDirection[column] = !sortDirection[column]

		const direction = sortDirection[column] ? 1 : -1

		rows.sort((a,b) => {

			let A = a.cells[column].innerText
			let B = b.cells[column].innerText

			if(!isNaN(A) && !isNaN(B)){
				return (A - B) * direction
			}

			return A.localeCompare(B) * direction
		})

		table.innerHTML = ""

		rows.forEach(row => table.appendChild(row))
	}

	</script>
	
	<table border="3" cellpadding="10" align="center" cellspacing="5">
	<caption>Mi lista de series</caption>
	<thead>
	<tr bgcolor="lightgray">
	<th onClick="sortTable(0)">ID de la serie</th>
	<th onClick="sortTable(1)">Nombre de la serie</th>
	<th onClick="sortTable(2)">Episodio en el que voy</th>
	<th onClick="sortTable(3)">Episodios totales</th>
	<th>Progreso de la serie</th>
	<th>Agregar episodio visto</th>
	</tr>
	</thead>
	<tbody id="seriesTable">
	`
	html += table_data
	html += `</tbody>`

	nav := ""

	if hasPrev {
		nav += fmt.Sprintf(`<a href="/?page=%d">Anterior</a>`, prevPage)
	}

	nav += "&nbsp&nbsp"

	if hasNext {
		nav += fmt.Sprintf(`<a href="/?page=%d">Siguiente</a>`, nextPage)
	}

	html += fmt.Sprintf(`</table>	
	<a href='./create'>Añadir nueva serie</a>
	<br><br>
	%s
	</body></html>`, nav)

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

func handleUpdate(conn net.Conn, db *sql.DB, query url.Values) {
	id := query.Get("id")
	_, err := db.Exec(`UPDATE series
	SET current_episode = current_episode + 1
	WHERE id = ? AND current_episode < total_episodes`, id)

	if err != nil {
		fmt.Print("Error en update", err)
	}

	var newEpisode int
	db.QueryRow("SELECT current_episode FROM series WHERE id = ?", id).Scan(&newEpisode)
	body := fmt.Sprintf("%d", newEpisode)
	response := fmt.Sprintf(
		"HTTP/1.1 200 OK \r\n"+
			"Content-Type: text/plain\r\n"+
			"Content-Length: %d\r\n\r\n%s", len(body), body,
	)
	conn.Write([]byte(response))
}
