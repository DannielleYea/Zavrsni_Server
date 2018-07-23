package main

import (
	"net"
	"time"
	"fmt"
	"encoding/json"
	"math/rand"
)

type player struct {
	userId  string
	address net.Addr
	active  bool
}

var queueNumberOfPlayers int

type game struct {
	gameId    string
	playerOne player
	playerTwo player
	board     [3][3]int
}

type sendGameData struct {
	GameId    string `json:"game_id"`
	PlayerOne string `json:"player_one"`
	PlayerTwo string `json:"player_two"`
}

var conn *net.UDPConn
var queue []player
var activeGames []game

func start(port string) {

	udpAddr, err := net.ResolveUDPAddr("udp4", port)
	fmt.Println(udpAddr)

	if err != nil {
		panic(err)
		return
	}

	conn, err = net.ListenUDP("udp", udpAddr)
	fmt.Println("Started listening: Queue server")

	if err != nil {
		panic(err)
		return
	}

	go matchMakingAlgorithm(conn)

	for {
		handleConnection(conn)
	}
}

func matchMakingAlgorithm(con *net.UDPConn) {
	for {
		if queueNumberOfPlayers >= 2 {

			fmt.Println("Connected: " + queue[0].address.String() + " AND " + queue[1].address.String())
			var gameData sendGameData
			gameData.GameId = "asf" + string(rand.Intn(100))
			gameData.PlayerTwo = string(queue[0].userId)
			gameData.PlayerOne = string(queue[1].userId)
			decoded, err := json.Marshal(gameData)

			if err != nil {
				panic(err)
				return
			}

			con.WriteTo([]byte(decoded), queue[0].address)
			con.WriteTo([]byte(decoded), queue[1].address)
			//con.WriteTo([]byte("{\"game_id\":\"asfjhasf\"}"), queue[0].address)
			//y := findElementIndex(queue[1].userId)
			//queue = append(queue[:0], queue[1:]...)
			playerOne, queue := queue[0], queue[1:]
			playerTwo, queue := queue[0], queue[1:]

			activeGames = append(activeGames, game{gameId: gameData.GameId, playerOne: playerOne, playerTwo: playerTwo})
			playerOne = player{}
			playerTwo = player{}
			queueNumberOfPlayers = queueNumberOfPlayers - 2
			//forDelete = nil
			//_, queue = queue[0], queue[1:]
			//x := findElementIndex(queue[0].userId)
			fmt.Println(queue)
		} else {
			time.Sleep(time.Second)
		}
	}
}

func handleConnection(con *net.UDPConn) {
	buffer := make([]byte, 1024)
	n, addr, err := con.ReadFromUDP(buffer)

	if err != nil {
		panic(err)
		return
	}

	userId := string(buffer[:n])
	//conn.WriteTo([]byte("ACK"), addr)
	//fmt.Println(con.LocalAddr().String())
	fmt.Println("Player with id: " + userId)

	if contains(userId) {
		updateQueuePlayer(userId)
	} else {
		fmt.Println("Player " + userId + " has come to queue")
		var ply player
		ply.userId = userId
		ply.address = addr
		ply.active = true

		if contains(ply.userId) {
			index := findElementIndex(ply.userId)
			if index < 0 {
				queue[index] = ply
			}
		} else {
			queue = append(queue, ply)
		}
		queueNumberOfPlayers = queueNumberOfPlayers + 1

		go checkAlive(queue[len(queue)-1].userId)
	}
}

func updateQueuePlayer(userId string) {
	if cap(queue) == 0 {
		return
	}

	for index, a := range queue {
		if a.userId == userId {
			queue[index].active = true
		}
	}
}

func checkAlive(userId string) {
	active := true
	for active {
		time.Sleep(5 * time.Second)
		for index, a := range queue {
			if a.userId == userId {
				if a.active {
					fmt.Println("Alive: " + queue[index].address.String())
					queue[index].active = false
					break
				} else {
					//queueNumberOfPlayers = queueNumberOfPlayers - 1
					fmt.Println("Dead: " + queue[index].address.String())
					queue[index] = player{}
					queue = append(queue[:index], queue[index+1:]...)
					active = false
					break
				}
			}
		}
	}
	fmt.Println(queue)
}

func findElementIndex(userId string) int {
	if cap(queue) == 0 {
		return 0
	}

	for index, a := range queue {
		if a.userId == userId {
			return index
		}
	}
	return -1
}

func contains(user string) bool {
	for _, a := range queue {
		if a.userId == user {
			return true
		}
	}
	return false
}
