package main

import (
	"net"
	"fmt"
	"strings"
)

func startGameServer(address string) {
	udpAddr, err := net.ResolveUDPAddr("udp4", address)

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

	for {
		startListening(conn)
	}
}

func startListening(udpConn *net.UDPConn) {
	var buffer = make([]byte, 1024)
	_, addr, err := udpConn.ReadFromUDP(buffer)

	if err != nil {
		fmt.Println(err)
		return
	}

	go handleMove(udpConn, addr, &buffer)
}

func handleMove(udpConn *net.UDPConn, addr *net.UDPAddr, buffer *[]byte) {
	fmt.Println(string(*buffer))
	strings.Contains(string(*buffer), "game_id")
	*buffer = nil
}
