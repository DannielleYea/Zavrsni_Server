package main

import (
	"net"
	"time"
	"fmt"
)

var buffer = make([]byte, 1024)
var conn *net.UDPConn
var queryAddress []net.Addr

func start(port string) {
	udpAddr, err := net.ResolveUDPAddr("udp4", port)
	fmt.Println(udpAddr.Port)
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

	for true {

		_, addr, err := conn.ReadFromUDP(buffer)

		if err != nil {
			fmt.Println(err)
		} else if !contains(&queryAddress, addr) {
			queryAddress = append(queryAddress, addr)
			fmt.Println(queryAddress)
			if len(queryAddress) == 2 {
				sendGameData([]byte("Start game"), queryAddress[0])
				sendGameData([]byte("Start game"), queryAddress[1])
				_, queryAddress = queryAddress[0], queryAddress[1:]
				_, queryAddress = queryAddress[0], queryAddress[1:]

			} else {
				go testConnection(addr)
			}
		}
	}
}

func contains(adresses *[]net.Addr, address net.Addr) bool {
	for _, a := range *adresses {
		if a.String() == address.String() {
			return true
		}
	}
	return false
}

func testConnection(addr net.Addr) {
	buff := make([]byte, 1024)

	var active = true

	for active {
		fmt.Println("testing")
		buff = make([]byte, 1024)
		time.Sleep(2 * time.Second)

		conn.WriteTo([]byte(addr.String()), addr)
		conn.SetDeadline(time.Now().Add(2 * time.Second))
		conn.ReadFromUDP(buff)

		if buff == nil {
			active = false
		}
	}
}

func sendGameData(data []byte, addr net.Addr) {

	conn.WriteTo(data, addr)
}
