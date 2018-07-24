package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"net/http"
	"container/list"
)

var addr = "192.168.5.12"

var db *sql.DB
var err error

type QueryUser struct {
	user   LoginData
	writer http.ResponseWriter
}

//var queryData []QueryUser

type SendData struct {
	Code    int       `json:"code"`
	Message string    `json:"message"`
	Data    LoginData `json:"data"`
}
type SnedFriendList struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Data    []Friend `json:"data"`
}

type RegisterData struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type LoginData struct {
	UserId   string `json:"user_id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type ResponseMessage struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
type Friend struct {
	UserId   string `json:"user_id"`
	Username string `json:"username"`
}

var query list.List

func login(w http.ResponseWriter, r *http.Request) {

	if !checkAuth(r) {
		sendResonseMessage(w, 8, "Authentication needed!")
		return
	}

	err := r.ParseForm()

	if r.Form.Encode() == "" {
		sendResonseMessage(w, 7, "Empty post form")
		return
	}

	if err != nil {
		sendResonseMessage(w, 5, "Error with form data")
	}

	var data RegisterData

	data.Username = r.FormValue("username")
	data.Password = r.FormValue("password")

	selectResult, err := db.Query("SELECT * FROM user WHERE username='" + data.Username + "' AND password = '" + data.Password + "';")

	if err != nil {
		sendResonseMessage(w, 4, "Error with Database")
		return
	}

	for selectResult.Next() {
		var data LoginData

		err = selectResult.Scan(&data.UserId, &data.Username, &data.Password, &data.Email)
		if err != nil {
			panic(err.Error())
			sendResonseMessage(w, 6, "Error with Database")
		}

		var send SendData
		send.Code = 1
		send.Data = data

		decoded, err := json.Marshal(send)
		if err != nil {
			panic(err)
			sendResonseMessage(w, 10, "Internal error - parsing JSON")
		}
		fmt.Fprintln(w, string(decoded))
		return
	}

	sendResonseMessage(w, 12, "Wrong username/password")

}

func register(w http.ResponseWriter, r *http.Request) {

	if !checkAuth(r) {
		sendResonseMessage(w, 8, "Authentication needed!")
		return
	}

	r.ParseForm()

	if r.Form.Encode() == "" {
		sendResonseMessage(w, 7, "Empty post form")
		return
	}

	var data RegisterData
	err := r.ParseForm()
	if err != nil {
		sendResonseMessage(w, 3, "Error with Form data")
		fmt.Println(err)
		return
	}

	data.Username = r.FormValue("username")
	data.Password = r.FormValue("password")
	data.Email = r.FormValue("email")

	db.Begin()

	querySelect, err := db.Query("SELECT COUNT(*) FROM (SELECT * FROM user where username='" + data.Username + "') AS Count;")

	querySelect.Next()

	if err != nil {
		sendResonseMessage(w, 2, "Error with Database")
		fmt.Println(err)
	}
	var Count int
	err = querySelect.Scan(&Count)
	if err != nil {
		fmt.Println(err)
		return
	}

	if Count > 0 {
		sendResonseMessage(w, 8, "Username already exists")
		return
	}
	insert, err := db.Query("INSERT INTO user (user_id, username, password, email) VALUES(DEFAULT, '" + r.Form.Get("username") + "', '" + (data.Password) + "', '" + (data.Email) + "');")

	if err != nil {
		sendResonseMessage(w, 2, "Error with Database")
		fmt.Println(err)
	}

	registeredUser, err := db.Query("SELECT * FROM user WHERE username='" + r.Form.Get("username") + "';")

	if err != nil {
		fmt.Println(err)
		return
	}

	for registeredUser.Next() {
		var data LoginData

		err = registeredUser.Scan(&data.UserId, &data.Username, &data.Password, &data.Email)
		if err != nil {
			panic(err.Error())
			sendResonseMessage(w, 6, "Error with Database")
		}

		var send SendData
		send.Code = 1
		send.Data = data

		decoded, err := json.Marshal(send)
		if err != nil {
			panic(err)
			sendResonseMessage(w, 10, "Internal error - parsing JSON")
		}
		fmt.Fprintln(w, string(decoded))
		return
	}

	//sendResonseMessage(w, 1, "User registered successfully")
	insert.Close()
}

func forgottenPassword(w http.ResponseWriter, r *http.Request) {
	if !checkAuth(r) {
		sendResonseMessage(w, 8, "Authentication needed!")
		return
	}

	err = r.ParseForm()

	if r.Form.Encode() == "" {
		sendResonseMessage(w, 7, "Empty post form")
		return
	}

	if err != nil {
		sendResonseMessage(w, 3, "Error with Form data")
		return
	}

	var data LoginData

	data.Username = r.FormValue("username")
	data.Email = r.FormValue("email")

	selectQuery, err := db.Query("SELECT * FROM user WHERE username='" + data.Username + "' AND email='" + data.Email + "'")

	if err != nil {
		sendResonseMessage(w, 2, "Error with Database")
		fmt.Println(err)
		return
	}

	for selectQuery.Next() {
		var data LoginData

		data.Password = ""

		err = selectQuery.Scan(&data.UserId, &data.Username, &data.Password, &data.Email)
		if err != nil {
			panic(err.Error())
			sendResonseMessage(w, 6, "Error with Database")
		}

		if data.Password == "" {
			sendResonseMessage(w, 9, "Username and Email doesn't match any user")
			return
		}

		decoded, err := json.Marshal(data)
		if err != nil {
			panic(err)
			sendResonseMessage(w, 10, "Internal error")
		}
		fmt.Fprintln(w, string(decoded))
	}
}

func getInQuery(w http.ResponseWriter, r *http.Request) {
	if !checkAuth(r) {
		sendResonseMessage(w, 8, "Authentication needed!")
		return
	}

	err = r.ParseForm()

	if r.Form.Encode() == "" {
		sendResonseMessage(w, 7, "Empty post form")
		return
	}

	if err != nil {
		sendResonseMessage(w, 3, "Error with Form data")
		return
	}

	var user LoginData

	userId := r.FormValue("user_id")

	queryResult, err := db.Query("SELECT * FROM user WHERE user_id='" + userId + "'")

	if err != nil {
		panic(err)
		sendResonseMessage(w, 9, "User Id doesn't exist")
	}

	for queryResult.Next() {
		queryResult.Scan(&user.UserId, &user.Username, user.Password, &user.Email)
	}

	var query QueryUser
	query.user = user
	query.writer = w

	//queryData = append(queryData, query)

	decoded, err := json.Marshal(addr)

	if err != nil {
		panic(err)
		sendResonseMessage(w, 10, "Internal error")
	}
	fmt.Fprint(w, string(decoded))
	fmt.Println(addr)

}

func getServerAddress(w http.ResponseWriter, r *http.Request) {
	if !checkAuth(r) {
		sendResonseMessage(w, 8, "Authentication needed!")
		return
	}

	decoded, err := json.Marshal(addr)

	if err != nil {
		panic(err)
		sendResonseMessage(w, 10, "Internal error")
	}
	fmt.Fprint(w, string(decoded))

}
func resetPassword(w http.ResponseWriter, r *http.Request) {

	if !checkAuth(r) {
		sendResonseMessage(w, 8, "Authentication needed!")
		return
	}

	err = r.ParseForm()

	if r.Form.Encode() == "" {
		sendResonseMessage(w, 7, "Empty post form")
		return
	}

	if err != nil {
		sendResonseMessage(w, 3, "Error with Form data")
		return
	}

	var data LoginData

	password := r.FormValue("password")
	repeatedPassword := r.FormValue("repeated_password")
	data.Email = r.FormValue("email")
	data.Password = r.FormValue("new_password")

	if data.Password == repeatedPassword {
		queryResult, err := db.Query("SELECT * FROM user WHERE email='" + data.Email + "' AND password='" + password + "'")

		if err != nil {
			panic(err)
			sendResonseMessage(w, 9, "Username and Email doesn't match any user")
		}

		var pass string
		for queryResult.Next() {
			queryResult.Scan(&data.UserId, &data.Username, pass, &data.Email)

			decoded, err := json.Marshal(data)
			if err != nil {
				panic(err)
				sendResonseMessage(w, 10, "Internal error")
			}
			fmt.Fprintln(w, string(decoded))

			insert, err := db.Query("UPDATE user SET password='" + data.Password + "' WHERE user_id=" + data.UserId + ";")

			if err != nil {
				sendResonseMessage(w, 2, "Error with Database")
				fmt.Println(err)
			}

			sendResonseMessage(w, 11, "Password updated successfully")
			insert.Close()
		}
	} else {
		sendResonseMessage(w, 12, "Password are not equal")
	}
}

func getFriendList(w http.ResponseWriter, r *http.Request) {

	if !checkAuth(r) {
		sendResonseMessage(w, 8, "Authentication needed!")
		return
	}

	err = r.ParseForm()

	if r.Form.Encode() == "" {
		sendResonseMessage(w, 7, "Empty post form")
		return
	}

	if err != nil {
		sendResonseMessage(w, 3, "Error with Form data")
		return
	}

	userId := r.FormValue("user_id")

	selectQuery, err := db.Query("SELECT us.user_id, us.username FROM friendList f JOIN user u ON f.user_id = u.user_id JOIN user us ON us.user_id = f.friend_id WHERE u.user_id=" + userId)

	if err != nil {
		panic(err)
		sendResonseMessage(w, 2, "Error with Database")
		return
	}

	var user Friend
	lista := []Friend{}

	for selectQuery.Next() {
		selectQuery.Scan(&user.UserId, &user.Username)

		lista = append(lista, user)
	}

	var data SnedFriendList

	data.Code = 1
	data.Data = lista

	decoded, err := json.Marshal(data)
	if err != nil {
		panic(err)
		sendResonseMessage(w, 10, "Internal error")
	}
	fmt.Fprint(w, string(decoded))
}

func getUserById(w http.ResponseWriter, r *http.Request) {

	if !checkAuth(r) {
		sendResonseMessage(w, 8, "Authentication needed!")
		return
	}

	err = r.ParseForm()

	if r.Form.Encode() == "" {
		sendResonseMessage(w, 7, "Empty post form")
		return
	}

	if err != nil {
		sendResonseMessage(w, 3, "Error with Form data")
		return
	}

	userId := r.FormValue("user_id")

	seletQuery, err := db.Query("SELECT * FROM user WHERE user_id=" + userId)

	if err != nil {
		panic(err)
		sendResonseMessage(w, 9, "Username and Email doesn't match any user")
	}

	for seletQuery.Next() {
		var user LoginData

		seletQuery.Scan(&user.UserId, &user.Username, &user.Password, &user.Email)

		decoded, err := json.Marshal(user)
		if err != nil {
			panic(err)
			sendResonseMessage(w, 10, "Internal error")
		}
		fmt.Fprint(w, string(decoded))
	}

}

func removePlayerFromFriendQueue(w http.ResponseWriter, r *http.Request) {
	if !checkAuth(r) {
		sendResonseMessage(w, 8, "Authentication needed!")
		return
	}

	err = r.ParseForm()

	if r.Form.Encode() == "" {
		sendResonseMessage(w, 7, "Empty post form")
		return
	}

	if err != nil {
		sendResonseMessage(w, 3, "Error with Form data")
		return
	}

	userId := r.FormValue("user_id")

	for index, player := range FriendActivePlayers {
		if player.UserId == userId {
			FriendActivePlayers[index] = FriendActivePlayer{}
			FriendActivePlayers = append(FriendActivePlayers[index:], FriendActivePlayers[index+1:]...)
		}
	}
}

func sendResonseMessage(w http.ResponseWriter, code int, message string) {
	var response ResponseMessage
	response.Code = code
	response.Message = message

	decoded, err := json.Marshal(response)
	if err != nil {
		panic(err)
	}
	fmt.Fprintln(w, string(decoded))
}

func checkAuth(r *http.Request) bool {

	if r.Header.Get("auth") != "K7DT8M18PLOM" || r.Header.Get("auth") == "" {
		return false
	}
	return true
}

func main() {

	go startGameServer(addr + ":4500")
	go start(addr + ":1000")
	go startFriendServer(addr + ":1500")
	mux := http.NewServeMux()

	db, err = sql.Open("mysql", "root:xKji27rC@tcp(localhost:3306)/mydb")

	if err != nil {
		fmt.Println(err)
		db.Close()
		return
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("auth")
		if auth == "K7DT8M18PLOM" {
			fmt.Fprintln(w, "List of active API end points ")
			fmt.Fprintln(w, "Register: /register \nLogin: /login\nForgotten Password /forgottenPassword\nReset Password /resetPassword\nGet in query /getInQuery"+
				"\n/GetFriendList: /getFriendList\nGet user by Id: /getUserById\n/Get server address: /getServerAddress"+
				"\n/Remove Player From Friend Queue: /removePlayerFromFriendQueue")
		} else {
			fmt.Fprintln(w, "Autentification needed!!")
		}
	})

	mux.HandleFunc("/login", login)
	mux.HandleFunc("/register", register)
	mux.HandleFunc("/forgottenPassword", forgottenPassword)
	mux.HandleFunc("/resetPassword", resetPassword)
	mux.HandleFunc("/getInQuery", getInQuery)
	mux.HandleFunc("/getFriendList", getFriendList)
	mux.HandleFunc("/getUserById", getUserById)
	mux.HandleFunc("/getServerAddress", getServerAddress)
	mux.HandleFunc("/removePlayerFromFriendQueue", removePlayerFromFriendQueue)

	http.ListenAndServe(":80", mux)

}
