package main

import (
	"net"
	"fmt"
	"encoding/json"
)

type GameMove struct {
	UserId string `json:"user_id"`
	GameId string `json:"game_id"`
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

	fmt.Println(string(buffer[:n]))

	err := json.Unmarshal(buffer[:n], &move)

	if err != nil {
		panic(err)
		return
	}

	if move.Turn == 0 {
		fmt.Println("Player " + move.UserId + " confirmed game")
		for index, game := range activeGames {
			if game.gameId == move.GameId {
				if move.Player == 1 {
					activeGames[index].playerOne.address = addr
					//gameConn.WriteTo([]byte("OK"), activeGames[index].playerOne.address)
					gameConn.WriteTo([]byte("OK"), addr)
				} else {
					activeGames[index].playerTwo.address = addr
					//gameConn.WriteTo([]byte("OK"), activeGames[index].playerTwo.address)
					gameConn.WriteTo([]byte("OK"), addr)
				}
			}
		}
	} else {
		if move.Player == 1 {
			fmt.Println("Player 1")
			for index, game := range activeGames {
				if game.gameId == move.GameId {
					sendTo := game.playerTwo.address
					writeAMove(&activeGames[index], move)
					if move.Turn > 4 {
						checkGame(&activeGames[index], &move)
					}

					decoded, _ := json.Marshal(move)
					gameConn.WriteTo(decoded, sendTo)
					if move.Winner != 0 {
						gameConn.WriteTo(decoded, addr)
					}
				}
			}
		} else {
			fmt.Println("Player 2")
			for index, game := range activeGames {
				if game.gameId == move.GameId {
					sendTo := game.playerOne.address

					writeAMove(&activeGames[index], move)
					if move.Turn > 4 {
						checkGame(&activeGames[index], &move)
					}

					decoded, _ := json.Marshal(move)
					gameConn.WriteTo(decoded, sendTo)
					if move.Winner != 0 {
						gameConn.WriteTo(decoded, addr)
					}
				}
			}
		}
	}

	if move.Winner != 0 {
		for index, currentGame := range activeGames {
			if currentGame.gameId == move.GameId {
				activeGames[index] = game{}
				activeGames = append(activeGames[:index], activeGames[index + 1:]...)
			}
		}
	}

	//*buffer = nil
	//gameConn.WriteTo([]byte("{\"field\":4,\"player\":2,\"turn\":0,\"user_id\":\"1\",\"winner\":0}"), addr)
}
