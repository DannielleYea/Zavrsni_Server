package main

import ("fmt"
	"net/http"
)

func login(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Login data")
}

func register(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Register data")
}

func main() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, you've requested: %s\n", r.URL.Path)
	})

	http.HandleFunc("/login", login);
	
	http.ListenAndServe(":80", nil)
	
	fmt.Println("Hello world!")
}