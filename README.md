# HTTP Server and Proxy - Lab 1 (TDA596/DIT240)
## Overview 
A HTTP server and proxy implementation in Go where methods such as GET (file serving) and POST (file uploading) are supported with proper error handling.

### HTTP Server
- Concurrent request handling: Uses goroutines to handle at most 10 connections at the same time
- GET method: Serves files with proper content-type headers
- POST method: Accepts files uploads and stores them
- Supported file types: html, txt, gif, jpg, jpeg, css
- Error handling: Status code 400, 404, 500, 501

### Proxy Server
- Forward proxy: intercepts and forwarsd HTTP requests
- Concurrency: handles at most 10 connections at the same time
- Protocol compliance: HTTP/1.0 response handling
- Error handling: Appropriate status code for invalid requests

## Project Structure
```
lab1-http-server/
├── server/
│   └── main.go             # HTTP server implementation
├── proxy/
│   └── main.go             # Proxy server implementation  
├── uploads/                # Directory for served/uploaded files
├── Dockerfile              # Docker definition for server and proxy
├── docker-compose.yml      # Multi-container orchestration
├── .dockerignore           # Files to exclude from Docker builds
├── go.mod                  # Go module dependencies
├── http_server             # Compiled HTTP server (generated)
├── webproxy                # Compiled proxy (generated)
└── README.md               # This file, Project documentation
```

## Installation & Usage
### Prerequisites
Go 1.25 

### Building 
go build -o http_server ./server

go build -o webproxy ./proxy

### Running
./http_server 8080

./proxy 9090

## Cloud Deployment
The application has been tested on AWS EC2 instances. Ensure security groups allow traffic on the configured ports (default: 8080, 9090).