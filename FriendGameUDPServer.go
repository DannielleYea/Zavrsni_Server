package main

import (
	"net"
	"fmt"
	"strings"
	"encoding/json"
	"math/rand"
)

type FriendRequest struct {
	UserId       string `json:"user_id"`
	FriendUserId string `json:"friend_user_id"`
	Status       string `json:"status"`
	Game         game   `json:"game_data"`
}

type FriendActivePlayer struct {
	UserId  string
	Address net.Addr
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

	go handleGame(friendConn, addr, buffer, n)
}

func handleGame(friendConn *net.UDPConn, udpAddr *net.UDPAddr, bytes []byte, i int) {

	var data FriendRequest

	json.Unmarshal(bytes, &data)

	if strings.Compare(data.Status, "Ask") == 0 {
		fmt.Println("Ask")
		for index, currentPlayer := range FriendActivePlayers {
			if currentPlayer.UserId == data.UserId {
				playerOne := player{userId: data.UserId, address: udpAddr}
				game := game{gameId: "gsdag" + string(rand.Intn(100)), playerOne: playerOne}
				GameOnHold = append(GameOnHold, game)
				request := FriendRequest{UserId: data.FriendUserId, FriendUserId: data.UserId, Status: "Request", Game: game}
				decoded, _ := json.Marshal(request)
				friendConn.WriteTo(decoded, FriendActivePlayers[index].Address)
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

				encode, _ := json.Marshal(startGame)
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
