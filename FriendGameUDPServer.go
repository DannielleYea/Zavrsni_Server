package main

import (
	"net"
	"fmt"
	"strings"
	"encoding/json"
	"math/rand"
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
	fmt.Println(string(bytes[:i]) + "Address: " + udpAddr.String())
	json.Unmarshal(bytes[:i], &data)

	if strings.Compare(data.Status, "Ask") == 0 {
		fmt.Println("Ask")
		for index, currentPlayer := range FriendActivePlayers {
			if currentPlayer.UserId == data.FriendUserId {
				playerOne := player{userId: data.UserId, address: udpAddr}
				sendForSend := sendGameData{GameId: "gsdag" + string(rand.Intn(100)), PlayerOne: data.UserId, PlayerTwo: data.FriendUserId}
				game := game{gameId: sendForSend.GameId, playerOne: playerOne}

				GameOnHold = append(GameOnHold, game)
				request := FriendRequest{UserId: data.FriendUserId, FriendUserId: data.UserId, Status: "Request", Game: sendForSend}
				decoded, _ := json.Marshal(request)
				fmt.Println("Address: " + udpAddr.String())
				friendConn.WriteTo(decoded, FriendActivePlayers[index].Address)
				fmt.Println(string(decoded))
				return
			}
		}
		decoded, _ := json.Marshal(FriendRequest{Status: "Not Found"})
		fmt.Println("Not Found")
		friendConn.WriteTo(decoded, udpAddr)
	} else if strings.Compare(data.Status, "Register") == 0 {

		fmt.Println("Register")
		FriendActivePlayers = append(FriendActivePlayers, FriendActivePlayer{UserId: data.UserId, Address: udpAddr})
	} else if strings.Compare(data.Status, "Accept") == 0 {

		fmt.Println("Accept")
		for index, currentGame := range GameOnHold {
			if currentGame.playerOne.userId == data.FriendUserId {
				startGame := GameOnHold[index]
				GameOnHold = append(GameOnHold[:index], GameOnHold[index+1:]...)
				startGame.playerTwo = player{userId: data.UserId, address: udpAddr}
				activeGames = append(activeGames, startGame)

				friendGameData := FriendGameData{GameData: data.Game, Status: "Accept"}

				encode, _ := json.Marshal(friendGameData)
				friendConn.WriteTo(encode, startGame.playerOne.address)
				friendConn.WriteTo(encode, startGame.playerTwo.address)

				startGame = game{}
			}
		}
	} else if strings.Compare(data.Status, "Deny") == 0 {
		fmt.Println("Deny")
		decoded, _ := json.Marshal(FriendRequest{Status: "Deny"})
		for index, currentGame := range GameOnHold {
			if currentGame.playerOne.userId == data.FriendUserId {
				GameOnHold[index].playerOne = player{}
				GameOnHold[index].playerTwo = player{}
				GameOnHold[index] = game{}
				GameOnHold = append(GameOnHold[:index], GameOnHold[index+1:]...)
				friendConn.WriteTo(decoded, currentGame.playerOne.address)
			}
		}
	}
}
