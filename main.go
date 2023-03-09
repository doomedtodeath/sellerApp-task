package main

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"math/big"
	"math/rand"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type data struct {
	id         string
	link       string
	created_on time.Time
}

var db *sql.DB

func generateUUID() string {
	rand.Seed(time.Now().UnixNano())
	e := rand.Uint64()
	eb := big.NewInt(int64(e))
	return base64.RawURLEncoding.EncodeToString(eb.Bytes())
}

func handleLongLink(w http.ResponseWriter, r *http.Request) {
	var err error
	var call data
	mappedData := r.URL.Query()
	id := generateUUID()
	for {
		err = db.QueryRow("Select id, link from Links where id = ?", id).Scan(&call.id, &call.link)
		if err != nil {
			break
		}
		id = generateUUID()
	}
	if len(mappedData["link"]) > 0 {
		originalLink := mappedData["link"]
		for _, val := range originalLink {
			_, err := db.Exec("INSERT INTO Links(id, link, created_on) VALUES (?,?, NOW())", id, val)
			if err != nil {
				panic(err.Error())
			}
			w.WriteHeader(http.StatusCreated)
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprint(w, "<a href=http://localhost:8080/"+id+">Your shortend URL</a>")

		}
	}
}

func handleShortLink(w http.ResponseWriter, r *http.Request) {
	urlCode := r.URL.Path[1:]
	var url data
	err := db.QueryRow("Select id, link, created_on from links where id = ?", urlCode).Scan(&url.id, &url.link, &url.created_on)
	if err != nil {
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, "<b>Link not found</b>")
		return
	}
	if time.Since(url.created_on).Hours() < 24.0 {
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, "<b>Link Expired</b>")
		return
	}
	http.Redirect(w, r, url.link, http.StatusPermanentRedirect)

}

func main() {
	var err error
	db, err = sql.Open("mysql", "root:appointy@tcp(127.0.0.1:3306)/Task?parseTime=true")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()
	create, err := db.Query("create table if not exists Links(id varchar(50) NOT NULL PRIMARY KEY, link varchar(50), created_on datetime)")
	if err != nil {
		panic(err.Error())
	}
	defer create.Close()
	http.HandleFunc("/", handleShortLink)
	http.HandleFunc("/original-link", handleLongLink)
	fmt.Println("Running.....")
	http.ListenAndServe(":8080", nil)

}
