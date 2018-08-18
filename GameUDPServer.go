package main

import (
	"net"
	"fmt"
	"encoding/json"
	"time"
	"sync"
	"strconv"
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
		confirmGameMutex := sync.Mutex{}
		for index, game := range activeGames {
			confirmGameMutex.Lock()
			if game.gameId == move.GameId {
				fmt.Println("Player " + move.UserId + " confirmed game ")
				if move.Player == 1 {
					fmt.Println("Plyer 1")
					activeGames[index].playerOne.address = addr
					gameConn.WriteTo([]byte("OK"), addr)
					//gameConn.WriteTo([]byte("OK"), activeGames[index].playerOne.address)
					activeGames[index].confirmed++
				} else {
					fmt.Println("Plyer 2")
					activeGames[index].playerTwo.address = addr
					gameConn.WriteTo([]byte("OK"), addr)
					//gameConn.WriteTo([]byte("OK"), activeGames[index].playerTwo.address)
					activeGames[index].confirmed++
				}

				if activeGames[index].confirmed == 2 {
					currentTime := int(time.Now().Unix())
					i64 := strconv.Itoa(currentTime)
					fmt.Println(i64)
					querr := "INSERT INTO game(game_id, player_one, player_two, time)" +
						"VALUES('" + game.gameId + "', '" + game.playerOne.userId + "', '" + game.playerTwo.userId + "', " + i64 + ")"
					fmt.Println(querr)
					_, err = db.Query(querr)

					if err != nil {
						fmt.Println(err)
					}
				}
			}
			confirmGameMutex.Unlock()
		}
	} else {
		gameTurnMutex := sync.Mutex{}
		if move.Player == 1 {
			gameTurnMutex.Lock()
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
					_, err = db.Query("INSERT INTO turns(game_id, turn, player, cell)" +
						"VALUES('" + game.gameId + "', " + strconv.Itoa(move.Turn) + ", " + move.UserId + "," + strconv.Itoa(move.Feield) + ");")

					if err != nil {
						fmt.Println(err)
					}

					if move.Winner != 0 {
						gameConn.WriteTo(decoded, addr)
						_, err = db.Query("UPDATE game SET winner=" + move.UserId + ", turns=" + strconv.Itoa(move.Turn) + " WHERE game_id='" + move.GameId + "';")

						if err != nil {
							fmt.Println(err)
						}
					}
				}
			}
			gameTurnMutex.Unlock()
		} else {
			gameTurnMutex.Lock()
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
					_, err = db.Query("INSERT INTO turns(game_id, turn, player, cell)" +
						"VALUES('" + game.gameId + "', " + strconv.Itoa(move.Turn) + ", " + move.UserId + "," + strconv.Itoa(move.Feield) + ");")

					if err != nil {
						fmt.Println(err)
					}

					if move.Winner != 0 {
						gameConn.WriteTo(decoded, addr)
						_, err = db.Query("UPDATE game SET winner=" + move.UserId + ", turns=" + strconv.Itoa(move.Turn) + " WHERE game_id='" + move.GameId + "';")

						if err != nil {
							fmt.Println(err)
						}

					}
				}
			}
			gameTurnMutex.Unlock()
		}
	}

	if move.Winner != 0 {
		for index, currentGame := range activeGames {
			if currentGame.gameId == move.GameId {
				activeGames[index] = game{}
				activeGames = append(activeGames[:index], activeGames[index+1:]...)
			}
		}
	}

	//*buffer = nil
	//gameConn.WriteTo([]byte("{\"field\":4,\"player\":2,\"turn\":0,\"user_id\":\"1\",\"winner\":0}"), addr)
}
