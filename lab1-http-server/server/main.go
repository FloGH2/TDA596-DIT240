package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const numClients = 10 //at most 10 according to exercise

func main() {
	if len(os.Args) < 2 { //slice of strings
		fmt.Println("start by entering this format: ./http_server <port>")
		return
	}

	var port string = os.Args[1]

	//server setup
	ln, err := net.Listen("tcp", ":"+port) //listen to tcp connectionis accordint to exercise
	if err != nil {
		log.Fatalln(err)
	}
	defer ln.Close() //close listener properly at the end of main function
	fmt.Printf("server listening on port %s\n", port)

	slots := make(chan struct{}, numClients) //limiting channel to at most 10 connections

	for {
		//infinite loop waiting to accpet incoming connections
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		slots <- struct{}{} //writes in channel (blocks if 10 clients are active)
		//handle connection
		go func(c net.Conn) { //create goroutine for each new client request
			defer func() { //anonymous function to close connection and release slot when done
				c.Close()
				<-slots //reads from channel and releases slot (no overwhelming)
			}()
			handleConnection(c) //using c as copy of conn as the value can change when used by several routines
		}(conn)
	}
}

func handleConnection(conn net.Conn) {
	reader := bufio.NewReader(conn) //buffer that reads data from client, offers better read operations than just conn.Reader

	request, err := http.ReadRequest(reader) //parses bytes from reader and converts it to structured object, request is a pointer to http.Request struct
	if err != nil {
		sendResponse(conn, 400, "Bad Request", "text/plain", []byte("Bad Request"))
		return
	}

	defer request.Body.Close() //body is a stream that needs to be closed after reading

	fmt.Printf("Request: %s %s from %s\n", request.Method, request.URL.Path, conn.RemoteAddr())

	switch request.Method {
	case "GET":
		getRequests(conn, request)
	case "POST":
		postRequests(conn, request)
	default: //not supported methods return 501
		sendResponse(conn, 501, "Not Implemented", "text/plain", []byte("501 Not Implemented"))
	}

}

func getRequests(conn net.Conn, request *http.Request) { //* means pointer which we use to avoid copying the whole struct
	extractedPath := request.URL.Path //"/fileName.html"

	if !isValidFileExtension(extractedPath) {
		sendResponse(conn, 400, "Bad Request", "text/plain", []byte("400 Bad Request: Invalid file extension"))
		return
	}

	baseDir := "uploads" //serve files from uploads directory
	safePath := filepath.Join(baseDir, filepath.Clean(extractedPath))

	if !strings.HasPrefix(safePath, baseDir) { //prevents directory traversal attacks
		sendResponse(conn, 400, "Bad Request", "text/plain", []byte("invalid path"))
		return
	}

	contentType := getContentType(extractedPath) //content type based on file extension

	data, err := os.ReadFile(safePath) //read file content
	if err != nil {
		if os.IsNotExist(err) {
			sendResponse(conn, 404, "Not Found", "text/plain", []byte("404 Not Found"))
		} else {
			sendResponse(conn, 500, "Internal Server Error", "text/plain", []byte("500 Internal Server Error"))
		}
		return
	}
	sendResponse(conn, 200, "OK", contentType, data)
}

func postRequests(conn net.Conn, request *http.Request) {
	extractedPath := request.URL.Path

	if !isValidFileExtension(extractedPath) {
		sendResponse(conn, 400, "Bad Request", "text/plain", []byte("400 Bad Request: Invalid file extension"))
		return
	}

	baseDir := "uploads"
	safePath := filepath.Join(baseDir, filepath.Clean(extractedPath))

	if !strings.HasPrefix(safePath, baseDir) {
		sendResponse(conn, 400, "Bad Request", "text/plain", []byte("invalid path"))
		return
	}

	err := os.MkdirAll(baseDir, 0755)
	if err != nil {
		sendResponse(conn, 500, "Internal Server Error", "text/plain", []byte("Error creating upload directory"))
		return
	}

	body, err := io.ReadAll(request.Body)
	if err != nil {
		sendResponse(conn, 400, "Bad Request", "text/plain", []byte("Error reading request body"))
		return
	}

	err = os.WriteFile(safePath, body, 0664)
	if err != nil {
		sendResponse(conn, 500, "Internal Server Error", "text/plain", []byte("Error saving file"))
		return
	}
	sendResponse(conn, 200, "OK", "text/plain", []byte("File uploaded successfully"))
}

// helper function to get content type based on file extension
func getContentType(path string) string {
	if strings.HasSuffix(path, ".html") {
		return "text/html"
	} else if strings.HasSuffix(path, ".txt") {
		return "text/plain"
	} else if strings.HasSuffix(path, ".gif") {
		return "image/gif"
	} else if strings.HasSuffix(path, ".jpeg") || strings.HasSuffix(path, ".jpg") {
		return "image/jpeg"
	} else if strings.HasSuffix(path, ".css") {
		return "text/css"
	}
	return "text/plain" //fallback according to exercise
}

// helper function to validate file extensions
func isValidFileExtension(path string) bool {
	validExtensions := []string{".html", ".txt", ".gif", ".jpeg", ".jpg", ".css"}
	for _, ext := range validExtensions {
		if strings.HasSuffix(path, ext) {
			return true
		}
	}
	return false
}

func sendResponse(conn net.Conn, code int, status string, contentType string, body []byte) {
	response := fmt.Sprintf("HTTP/1.0 %d %s\r\n", code, status) //http/1.0 one request per connection
	response += fmt.Sprintf("Content-Type: %s\r\n", contentType)
	response += fmt.Sprintf("Content-Length: %d\r\n", len(body))
	response += "\r\n"

	if _, err := conn.Write([]byte(response)); err != nil {
		log.Printf("Failed to send response header: %v", err)
		return
	}

	if len(body) > 0 {
		if _, err := conn.Write(body); err != nil {
			log.Printf("Failed to send response body: %v", err)
			return
		}
	}
}
