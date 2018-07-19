package main

import (
	"net"
	"time"
	"fmt"
)

type player struct {
	userId  string
	address net.Addr
	active  bool
}

type game struct {
	gameId    string
	playerOne player
	playerTwo player
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
		if cap(queue) >= 2 {
			//fmt.Println(cap(queue))
			//fmt.Println("Game found")
			fmt.Println("Connected: " + queue[0].address.String() + " AND " + queue[1].address.String())
			con.WriteTo([]byte("{\"game_id\":\"asfjhasf\"}"), queue[1].address)
			con.WriteTo([]byte("{\"game_id\":\"asfjhasf\"}"), queue[0].address)
			//y := findElementIndex(queue[1].userId)
			//queue = append(queue[:0], queue[1:]...)
			queue[0] = player{}
			playerOne, queue := queue[0], queue[1:]
			queue[0] = player{}
			playerTwo, queue := queue[0], queue[1:]

			activeGames = append(activeGames, game{playerOne: playerOne, playerTwo: playerTwo})
			//forDelete = nil
			//_, queue = queue[0], queue[1:]
			//x := findElementIndex(queue[0].userId)
		} else {
			time.Sleep(time.Second)
		}
	}
}

func handleConnection(con *net.UDPConn) {
	buffer := make([]byte, 1024)
	_, addr, err := con.ReadFromUDP(buffer)

	if err != nil {
		panic(err)
		return
	}
	//conn.WriteTo([]byte("ACK"), addr)
	//fmt.Println(con.LocalAddr().String())
	//fmt.Println("Player with id: " + string(buffer))

	if contains(string(buffer)) {
		updateQueuePlayer(string(buffer))
	} else {
		var ply player
		ply.userId = string(buffer)
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
					fmt.Println("Dead: " + queue[index].address.String())
					active = false
					break
				}
			}
			index++
			if index >= cap(queue) {
				return
			}
		}
	}

	index := findElementIndex(userId)

	if index >= 0 {
		if index == 0 {
			_, queue = queue[0], queue[:1]
		} else {
			queue[index] = player{}
			queue = append(queue[:index], queue[index+1:]...)
		}
	}

	//fmt.Println(queue)
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
