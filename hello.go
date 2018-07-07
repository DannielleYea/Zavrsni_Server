package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"net/http"
)

var db *sql.DB
var err error

type regData struct {
	username string
	password string
	email    string
}

func login(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Login data")
}

func register(w http.ResponseWriter, r *http.Request) {
	//fmt.Fprintf(w, "Register data")
	r.ParseForm()

	var data regData
	err := r.ParseForm()
	if err != nil {
		fmt.Println(err)
		return
	}

	data.username = r.FormValue("username")
	data.password = r.FormValue("password")
	data.email = r.FormValue("email")

	db.Begin()

	insert, err := db.Query("INSERT INTO user (user_id, username, password, email) VALUES(DEFAULT, '" + r.Form.Get("username") + "', '" + (data.password) + "', '" + (data.email) + "');")

	if err != nil {
		fmt.Println(err)
	}

	insert.Close()
}

func main() {

	mux := http.NewServeMux()

	db, err = sql.Open("mysql", "root:xKji27rC@tcp(localhost:3306)/mydb")

	if err != nil {
		fmt.Println(err)
		db.Close()
	}

	if err != nil {
		fmt.Println(err)
	}
	// be careful deferring Queries if you are using transactions
	//defer insert.Close()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, you've requested: %s\n", r.URL.Path)
	})

	mux.HandleFunc("/login", login)
	mux.HandleFunc("/register", register)

	http.ListenAndServe(":80", mux)

	fmt.Println("Hello world!")
}
