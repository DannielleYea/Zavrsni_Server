package main

import (
	"net"
	"time"
	"fmt"
	"encoding/json"
	"strings"
)

type player struct {
	userId  string
	address net.Addr
	active  bool
	inGame  bool
}

var queueNumberOfPlayers int

type GameRequest struct {
	UserId string `json:"user_id"`
	Status string `json:"status"`
}

type game struct {
	gameId    string
	playerOne player
	playerTwo player
	board     [3][3]int
	confirmed int
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
			gameData.GameId = gameIdGenerator(20)
			gameData.PlayerTwo = string(queue[0].userId)
			gameData.PlayerOne = string(queue[1].userId)
			decoded, err := json.Marshal(gameData)

			if err != nil {
				panic(err)
				return
			}

			con.WriteTo([]byte(decoded), queue[0].address)
			con.WriteTo([]byte(decoded), queue[1].address)

			queue[0].inGame = true
			queue[1].inGame = true
			playerOne, queue := queue[0], queue[1:]
			playerTwo, queue := queue[0], queue[1:]

			activeGames = append(activeGames, game{gameId: gameData.GameId, playerOne: playerOne, playerTwo: playerTwo})
			playerOne = player{}
			playerTwo = player{}
			queueNumberOfPlayers = queueNumberOfPlayers - 2

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

	var gameRequest GameRequest
	json.Unmarshal(buffer[:n], &gameRequest)
	fmt.Println("Player with id: " + gameRequest.UserId)

	if strings.Compare(gameRequest.Status, "Update") == 0 {
		updateQueuePlayer(gameRequest.UserId)
	} else if strings.Compare(gameRequest.Status, "Register") == 0 {
		fmt.Println("Player " + gameRequest.UserId + " has come to queue")
		var ply player
		ply.userId = gameRequest.UserId
		ply.address = addr
		ply.active = true
		ply.inGame = false

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
		time.Sleep(2*time.Second + 750*time.Millisecond)
		for index, a := range queue {
			if a.userId == userId {
				if a.active {
					fmt.Println("Alive: " + queue[index].address.String())
					queue[index].active = false
					break
				} else {
					if !a.inGame {
						queueNumberOfPlayers = queueNumberOfPlayers - 1
					}
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
