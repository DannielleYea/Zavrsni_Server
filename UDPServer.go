package main

import (
	"net"
	"fmt"
	"time"
)

type player struct {
	userId  string
	address net.Addr
	active  bool
}

var conn *net.UDPConn
var queue []player

func start(port string) {
	udpAddr, err := net.ResolveUDPAddr("udp4", port)

	fmt.Println(udpAddr)

	if err != nil {
		panic(err)
		return
	}

	conn, err = net.ListenUDP("udp", udpAddr)

	if err != nil {
		panic(err)
		return
	}

	//go matchMakingAlgorithm()

	for {
		handleConnection(conn)
	}
}

func matchMakingAlgorithm() {
	for {
		if len(queue) >= 2 {
			conn.WriteTo([]byte("{\"game_id\":\"asfjhasf\"}"), queue[0].address)
			conn.WriteTo([]byte("{\"game_id\":\"asfjhasf\"}"), queue[1].address)

			go func() {
				x := findElementIndex(queue[0].userId)
				queue = append(queue[:x], queue[x+1:]...)
			}()

			go func() {
				y := findElementIndex(queue[1].userId)
				queue = append(queue[:y], queue[y+1:]...)
			}()

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
	fmt.Println(string(buffer))

	if contains(string(buffer)) {
		updateQueuePlayer(string(buffer))
	} else {
		var ply player
		ply.userId = string(buffer)
		ply.address = addr
		ply.active = true
		queue = append(queue, ply)

		go checkAlive(queue[len(queue)-1].userId)
	}
}

func updateQueuePlayer(userId string) {
	if len(queue) == 0 {
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
			if index >= len(queue) {
				return
			}
		}
	}

	index := findElementIndex(userId)

	if index >= 0 {
		queue = append(queue[:index], queue[index+1:]...)
	}

	fmt.Println(queue)
}

func findElementIndex(userId string) int {
	if len(queue) == 0 {
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

func sendGameData(data []byte, addr net.Addr) {

	conn.WriteTo(data, addr)
}