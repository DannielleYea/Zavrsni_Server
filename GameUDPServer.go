package main

import (
	"net"
	"fmt"
)

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
	_, addr, err := gameConn.ReadFromUDP(buffer)

	if err != nil {
		fmt.Println(err)
		return
	}

	go handleMove(gameConn, addr, buffer)
}

func handleMove(gameConn *net.UDPConn, addr *net.UDPAddr, buffer []byte) {
	fmt.Println(string(buffer))
	//*buffer = nil
	gameConn.WriteTo([]byte("ACK"), addr)
}
