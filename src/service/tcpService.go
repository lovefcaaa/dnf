package service

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strconv"
)

func tcpZonesHandler(conn net.Conn) {
	/*
	   protocol:
	   struct pkg {
	       int32_t magic; // 0xCAFE
	       int32_t size;  // LITTLE ENDIAN
	       const char data[];
	   };
	*/
	defer conn.Close()

	var magic, size int32
	var data []byte

	if err := binary.Read(conn, binary.LittleEndian, &magic); err != nil {
		fmt.Println("Read magic error:", err)
		return
	}

	if magic != 0xCAFE {
		fmt.Println("magic number error:", magic)
		return
	}

	if err := binary.Read(conn, binary.LittleEndian, &size); err != nil {
		fmt.Println("Read size error:", err)
		return
	}

	if size <= 0 {
		fmt.Println("Size error:", size)
	}

	data = make([]byte, 0, int(size))

	if _, err := io.ReadFull(conn, data); err != nil {
		fmt.Println("Read data error:", err)
		return
	}

	handleZoneRequestData(conn, data)
}

func handleZoneRequestData(conn net.Conn, data []byte) {
	// TODO: parse json and invoke commitor.GetZonesInfo()
	conn.Write([]byte("hello world"))
}

func doTcpServe(port int, tcpHandler func(net.Conn)) {
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		panic("Listen tcp :" + strconv.Itoa(port) + "error" + err.Error())
	}
	defer ln.Close()
	for {
		if conn, err := ln.Accept(); err != nil {
			go tcpHandler(conn)
		} else {
			fmt.Println("TcpServe Accep conn error:", err)
		}
	}
}

func TcpServe() {
	// TODO: read port from conf file
	go doTcpServe(7778, tcpZonesHandler)
	//go doTcpServe(7779, tcpSearchHandler)
}
