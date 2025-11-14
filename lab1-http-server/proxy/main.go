package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
)

const NumClients = 10

func main() {
	if len(os.Args) < 2 {
		fmt.Println("start by entering this format: ./webproxy <port>")
		return
	}

	var port string = os.Args[1]

	ln, err := net.Listen("tcp", ":"+port)
	log.Printf("%s", ln.Addr().String())

	if err != nil {
		log.Fatalln(err)
	}
	defer ln.Close()
	fmt.Printf("proxy listening on port %s\n", port)

	slots := make(chan struct{}, NumClients)

	for {
		//accpet incoming connection
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		fmt.Println("Remote address:", conn.RemoteAddr())
		fmt.Println("Local address:", conn.LocalAddr())

		slots <- struct{}{} //write in channel (blocks if 10 clients are active)
		//handle connection
		go func(c net.Conn) { //create goroutine for each new client request
			defer func() {
				c.Close()
				<-slots //reads from channel and releases slot (no overwhelming)
			}()
			handleProxyConnection(c) //using c as copy of conn as the value can change when used by several routines
		}(conn)
	}

}

func handleProxyConnection(conn net.Conn) {
	reader := bufio.NewReader(conn)

	request, err := http.ReadRequest(reader)
	if err != nil {
		fmt.Fprint(conn, "HTTP/1.0 400 Bad Request\r\n\r\n")
		return
	}
	defer request.Body.Close()

	fmt.Println("Proxy received:", request.Method, request.URL.String())

	if request.Method != "GET" {
		fmt.Fprint(conn, "HTTP/1.0 501 Not Implemented\r\n\r\n")
		return
	}

	serverAddress := request.URL.Host //host:port
	if !strings.Contains(serverAddress, ":") {
		serverAddress += ":80"
	}

	serverConn, err := net.Dial("tcp", serverAddress)
	if err != nil {
		fmt.Fprint(serverConn, "HTTP/1.0 502 Bad Gateway\r\n\r\n")
		return
	}
	defer serverConn.Close()

	request.Write(serverConn)

	io.Copy(conn, serverConn)
}
