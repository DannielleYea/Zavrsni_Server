package main

import (
	"net"
	"fmt"
	"encoding/json"
)

type GameMove struct {
	UserId string `json:"user_id"`
	Turn   int    `json:"turn"`
	Feield int    `json:"field"`
	Player int    `json:"player"`
	Winner int    `json:"winner"`
}

func startGameServer(address string) {
	udpAddr, err := net.ResolveUDPAddr("udp4", address)

	fmt.Println(udpAddr)

	if err != nil {
		panic(err)
		return
	}

	gameConn, err := net.ListenUDP("udp", udpAddr)
	fmt.Println("Started listening: Game server")

	if err != nil {
		panic(err)
		return
	}

	for {
		startListening(gameConn)
	}
}

func startListening(gameConn *net.UDPConn) {
	buffer := make([]byte, 1024)
	n, addr, err := gameConn.ReadFromUDP(buffer)

	if err != nil {
		fmt.Println(err)
		return
	}

	go handleMove(gameConn, addr, buffer, n)
}

func handleMove(gameConn *net.UDPConn, addr *net.UDPAddr, buffer []byte, n int) {

	var move GameMove

	err := json.Unmarshal(buffer[:n], &move)

	if err != nil {
		panic(err)
		return
	}

	if move.Turn == 0 {
		for index, game range activeGames {
			if game.GameId = move.GameId {
				if move.Player = 1 {
					activeGames[index].PlayerOne.address = addr
					fmt.Fprintln("ACK", addr)
				} else {
					activeGames[index].PlayerTwo.address = addr
					fmt.Fprintln("ACK", addr)
				}
			}
		}
	} else {
		if move.Player == 1 {
			fmt.Println("Player 1")
			for _, game := range activeGames {
				if game.playerOne.userId == move.UserId {
					sendTo := game.playerTwo.address
					gameConn.WriteTo(buffer[:n], sendTo)
				}
			}
		} else {
			fmt.Println("Player 2")
			for _, game := range activeGames {
				if game.playerTwo.userId == move.UserId {
					sendTo := game.playerOne.address
					gameConn.WriteTo(buffer, sendTo)
				}
			}
		}
	}

	//*buffer = nil
	//gameConn.WriteTo([]byte("{\"field\":4,\"player\":2,\"turn\":0,\"user_id\":\"1\",\"winner\":0}"), addr)
}
