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

const NumClients = 10 //at most 10 according to exercise

func main() {
	if len(os.Args) < 2 { //slice of strings
		fmt.Println("start by entering this format: ./http_server <port>")
		return
	}

	var port string = os.Args[1]

	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalln(err)
	}
	defer ln.Close() //close listener properly when program exits
	fmt.Printf("server listening on port %s\n", port)

	slots := make(chan struct{}, NumClients) //limiting channel to at most 10 connections

	for {
		//accpet incoming connection
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		slots <- struct{}{} //write in channel (blocks if 10 clients are active)
		//handle connection
		go func(c net.Conn) { //create goroutine for each new client request
			defer func() {
				c.Close()
				<-slots //reads from channel and releases slot (no overwhelming)
			}()
			handleConnection(c) //using c as copy of conn as the value can change when used by several routines
		}(conn)
	}
}

func handleConnection(conn net.Conn) {
	reader := bufio.NewReader(conn) //buffer that reads data from client, offers better read operations than just conn.Reader

	request, err := http.ReadRequest(reader) //parses bytes from reader and converts it to structured object
	if err != nil {
		sendResponse(conn, 400, "Bad Request", "text/plain", []byte("Bad Request"))
		return
	}

	defer request.Body.Close()

	fmt.Printf("Request: %s %s from %s\n", request.Method, request.URL.Path, conn.RemoteAddr())

	switch request.Method {
	case "GET":
		getRequests(conn, request)
	case "POST":
		postRequests(conn, request)
	default:
		sendResponse(conn, 501, "Not Implemented", "text/plain", []byte("Not Implemented"))
	}

}

func getRequests(conn net.Conn, request *http.Request) { //* -> pointer so no copy which makes it faster
	extractedPath := request.URL.Path
	var contentType string
	if strings.HasSuffix(extractedPath, ".html") {
		contentType = "text/html"
	} else if strings.HasSuffix(extractedPath, ".txt") {
		contentType = "text/plain"
	} else if strings.HasSuffix(extractedPath, ".gif") {
		contentType = "image/gif"
	} else if strings.HasSuffix(extractedPath, ".jpeg") {
		contentType = "image/jpeg"
	} else if strings.HasSuffix(extractedPath, ".jpg") {
		contentType = "image/jpeg"
	} else if strings.HasSuffix(extractedPath, ".css") {
		contentType = "text/css"
	} else {
		sendResponse(conn, 400, "Bad Request", "text/plain", []byte("400 Bad Request"))
		return
	}

	data, err := os.ReadFile("." + extractedPath)
	if err != nil {
		sendResponse(conn, 404, "Not Found", "text/plain", []byte("404 Not Found")) //file does not exist
		return
	}
	sendResponse(conn, 200, "OK", contentType, data)
}

func postRequests(conn net.Conn, request *http.Request) {
	body, err := io.ReadAll(request.Body)
	if err != nil {
		sendResponse(conn, 404, "Not Found", "text/plain", []byte("404 Not Found"))
		return
	}
	err = os.WriteFile("uploads"+request.URL.Path, body, 0664)
	if err != nil {
		sendResponse(conn, 500, "Internal Server Error", "text/plain", []byte("500 Internal Server Error"))
		return
	}
	sendResponse(conn, 200, "OK", "text/plain", []byte("File uploaded successfully"))
}

func sendResponse(conn net.Conn, code int, status string, contentType string, body []byte) {
	response := fmt.Sprintf("HTTP/1.0 %d %s\r\n", code, status) //http/1.0 one request per connection
	response += fmt.Sprintf("Content-Type: %s\r\n", contentType)
	response += fmt.Sprintf("Content-Length: %d\r\n", len(body))
	response += "\r\n"

	conn.Write([]byte(response))
	conn.Write(body)
}
