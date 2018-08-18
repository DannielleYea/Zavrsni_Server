package main

import (
	"net"
	"fmt"
	"strings"
	"encoding/json"
	"sync"
)

type FriendRequest struct {
	UserId       string       `json:"user_id"`
	FriendUserId string       `json:"friend_user_id"`
	Status       string       `json:"status"`
	Game         sendGameData `json:"game_data"`
}

type FriendActivePlayer struct {
	UserId  string
	Address net.Addr
}

type FriendGameData struct {
	GameData sendGameData `json:"game_data"`
	Status   string       `json:"status"`
}

var RematchQueue []game
var rematchMutex sync.Mutex
var friendGameMutex sync.Mutex
var gameOnHoldMutex sync.Mutex
var GameOnHold []game
var FriendActivePlayers []FriendActivePlayer

func startFriendServer(address string) {
	udpAddr, err := net.ResolveUDPAddr("udp4", address)

	fmt.Println(udpAddr)

	if err != nil {
		panic(err)
		return
	}

	friendConn, err := net.ListenUDP("udp", udpAddr)
	fmt.Println("Started listening: Friend server")

	if err != nil {
		panic(err)
		return
	}

	for {
		handleFriendRequest(friendConn)
	}
}

func handleFriendRequest(friendConn *net.UDPConn) {
	buffer := make([]byte, 1024)
	n, addr, err := friendConn.ReadFromUDP(buffer)

	if err != nil {
		fmt.Println(err)
		return
	}

	go handleFriendGame(friendConn, addr, buffer, n)
}

func handleFriendGame(friendConn *net.UDPConn, udpAddr *net.UDPAddr, bytes []byte, i int) {

	var data FriendRequest

	json.Unmarshal(bytes[:i], &data)

	if strings.Compare(data.Status, "Ask") == 0 {
		friendGameMutex.Lock()
		gameOnHoldMutex.Lock()
		fmt.Println("Ask")
		for index, currentPlayer := range FriendActivePlayers {
			if currentPlayer.UserId == data.FriendUserId {
				playerOne := player{userId: data.UserId, address: udpAddr}
				sendForSend := sendGameData{GameId: gameIdGenerator(20), PlayerOne: data.UserId, PlayerTwo: data.FriendUserId}
				game := game{gameId: sendForSend.GameId, playerOne: playerOne}

				GameOnHold = append(GameOnHold, game)
				request := FriendRequest{UserId: data.FriendUserId, FriendUserId: data.UserId, Status: "Request", Game: sendForSend}
				decoded, _ := json.Marshal(request)
				fmt.Println("Address: " + udpAddr.String())
				go removePlayerFromFriendGameQueue(data.UserId)
				friendConn.WriteTo(decoded, FriendActivePlayers[index].Address)
				fmt.Println(string(decoded))
				friendGameMutex.Unlock()
				gameOnHoldMutex.Unlock()
				return
			}
		}
		decoded, _ := json.Marshal(FriendRequest{Status: "Not Found"})
		fmt.Println("Not Found")
		friendConn.WriteTo(decoded, udpAddr)
		gameOnHoldMutex.Unlock()
		friendGameMutex.Unlock()
	} else if strings.Compare(data.Status, "Register") == 0 {

		fmt.Println("Register")
		FriendActivePlayers = append(FriendActivePlayers, FriendActivePlayer{UserId: data.UserId, Address: udpAddr})
		fmt.Println(FriendActivePlayers)
	} else if strings.Compare(data.Status, "Unregister") == 0 {

		friendGameMutex.Lock()
		fmt.Println("Unregister")
		go removePlayerFromFriendGameQueue(data.UserId)

		//FriendActivePlayers = append(FriendActivePlayers, FriendActivePlayer{UserId: data.UserId, Address: udpAddr})

		friendGameMutex.Unlock()
	} else if strings.Compare(data.Status, "Accept") == 0 {
		gameOnHoldMutex.Lock()
		fmt.Println("Accept")
		for index, currentGame := range GameOnHold {
			if currentGame.playerOne.userId == data.FriendUserId {
				go removePlayerFromFriendGameQueue(data.UserId)
				startGame := GameOnHold[index]
				GameOnHold = append(GameOnHold[:index], GameOnHold[index+1:]...)
				startGame.playerTwo = player{userId: data.UserId, address: udpAddr}
				activeGames = append(activeGames, startGame)

				friendGameData := FriendGameData{GameData: data.Game, Status: "Accept"}

				encode, _ := json.Marshal(friendGameData)
				friendConn.WriteTo(encode, startGame.playerOne.address)
				friendConn.WriteTo(encode, startGame.playerTwo.address)

				fmt.Println(FriendActivePlayers)

				startGame = game{}
				gameOnHoldMutex.Unlock()
				return
			}
		}
		gameOnHoldMutex.Unlock()
	} else if strings.Compare(data.Status, "Deny") == 0 {
		gameOnHoldMutex.Lock()
		fmt.Println("Deny")
		decoded, _ := json.Marshal(FriendRequest{Status: "Deny"})
		for index, currentGame := range GameOnHold {
			if currentGame.playerOne.userId == data.FriendUserId {
				//FriendActivePlayers = append(FriendActivePlayers, FriendActivePlayer{UserId: data.FriendUserId, Address: GameOnHold[index].playerOne.address})
				GameOnHold[index].playerOne = player{}
				GameOnHold[index].playerTwo = player{}
				GameOnHold[index] = game{}
				GameOnHold = append(GameOnHold[:index], GameOnHold[index+1:]...)
				friendConn.WriteTo(decoded, currentGame.playerOne.address)
				gameOnHoldMutex.Unlock()
				return
			}
		}
		gameOnHoldMutex.Unlock()
	} else if strings.Compare(data.Status, "Again") == 0 {
		var rematchData FriendRequest

		fmt.Println("Receiced rematch:" + string(bytes[:i]))
		json.Unmarshal(bytes[:i], &rematchData)

		rematchMutex.Lock()
		for index, currentGame := range RematchQueue {
			if currentGame.gameId == rematchData.Game.GameId {
				playerTwo := player{userId: rematchData.UserId, address: udpAddr}
				RematchQueue[index].playerTwo = playerTwo

				if currentGame.confirmed == 1 {
					newGame := sendGameData{GameId: gameIdGenerator(20), PlayerOne: currentGame.playerOne.userId, PlayerTwo: playerTwo.userId}
					friendGame := FriendGameData{GameData: newGame, Status: "Accepted"}
					encoded, _ := json.Marshal(friendGame)
					friendConn.WriteTo(encoded, playerTwo.address)
					fmt.Println("Sended: " + string(encoded))
					friendConn.WriteTo(encoded, currentGame.playerOne.address)
					RematchQueue[index].gameId = newGame.GameId
					activeGames = append(activeGames, RematchQueue[index])
				} else {
					friendGame := FriendGameData{Status: "Denied"}
					encoded, _ := json.Marshal(friendGame)
					friendConn.WriteTo(encoded, playerTwo.address)
					friendConn.WriteTo(encoded, currentGame.playerOne.address)
				}
				RematchQueue[index] = game{gameId: "", playerOne: player{}, playerTwo: player{}}
				RematchQueue = append(RematchQueue[index:], RematchQueue[:index+1]...)
				rematchMutex.Unlock()
				return
			}
		}

		playerOne := player{userId: rematchData.UserId, address: udpAddr}
		rematchRequestData := game{gameId: rematchData.Game.GameId, playerOne: playerOne, confirmed: 1}
		RematchQueue = append(RematchQueue, rematchRequestData)

		rematchMutex.Unlock()
	} else if strings.Compare(data.Status, "Not again") == 0 {
		var rematchData FriendRequest

		fmt.Println("Receiced rematch:" + string(bytes[:i]))
		json.Unmarshal(bytes[:i], &rematchData)

		rematchMutex.Lock()
		for index, currentGame := range RematchQueue {
			if currentGame.gameId == rematchData.Game.GameId {
				playerTwo := player{userId: rematchData.UserId, address: udpAddr}
				RematchQueue[index].playerTwo = playerTwo

				if currentGame.confirmed == 1 {
					newGame := sendGameData{GameId: gameIdGenerator(20), PlayerOne: currentGame.playerOne.userId, PlayerTwo: playerTwo.userId}
					friendGame := FriendGameData{GameData: newGame, Status: "Accepted"}
					encoded, _ := json.Marshal(friendGame)
					friendConn.WriteTo(encoded, playerTwo.address)
					fmt.Println("Sended: " + string(encoded))
					friendConn.WriteTo(encoded, currentGame.playerOne.address)
					RematchQueue[index].gameId = newGame.GameId
					activeGames = append(activeGames, RematchQueue[index])
				} else {
					friendGame := FriendGameData{Status: "Denied"}
					encoded, _ := json.Marshal(friendGame)
					friendConn.WriteTo(encoded, playerTwo.address)
					friendConn.WriteTo(encoded, currentGame.playerOne.address)
				}
				RematchQueue[index] = game{gameId: "", playerOne: player{}, playerTwo: player{}}
				RematchQueue = append(RematchQueue[index:], RematchQueue[:index+1]...)
				rematchMutex.Unlock()
				return
			}
		}

		playerOne := player{userId: rematchData.UserId, address: udpAddr}
		rematchRequestData := game{gameId: rematchData.Game.GameId, playerOne: playerOne, confirmed: 1}
		RematchQueue = append(RematchQueue, rematchRequestData)

		rematchMutex.Unlock()
	}
}
func removePlayerFromFriendGameQueue(userId string) {
	for index, player := range FriendActivePlayers {
		if player.UserId == userId {
			FriendActivePlayers = append(FriendActivePlayers[:index], FriendActivePlayers[index+1:]...)
			return
		}
	}
}
