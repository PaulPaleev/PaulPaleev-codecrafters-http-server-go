package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleRequest(conn)
	}
}

func handleRequest(conn net.Conn) {
	defer conn.Close()

	req := make([]byte, 1024)
	conn.Read(req)
	strReq := string(req)
	fmt.Println("1. " + strReq)

	target := getRequestTarget(strReq)

	if target == "/" {
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	} else if strings.HasPrefix(target, "/echo") {
		body := strings.Split(target, "/")[2]
		finalStringToConvert := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(body), body)
		conn.Write([]byte(finalStringToConvert))
	} else if strings.HasPrefix(target, "/user-agent") {
		body := getUserAgent(string(req))
		finalStringToConvert := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(body), body)
		conn.Write([]byte(finalStringToConvert))
	} else if strings.HasPrefix(target, "/files/") {
		filename := getFilename(strReq)
		dir := os.Args[2] // ????
		body, err := os.ReadFile(dir + filename)
		if err != nil {
			sendNotFound(conn)
			return
		}
		finalStringToConvert := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", len(body), body)
		conn.Write([]byte(finalStringToConvert))
	} else {
		sendNotFound(conn)
	}
}

func sendNotFound(conn net.Conn) {
	conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
}

func getFilename(request string) string {
	fileString := strings.Split(request, "\r\n")[0]
	fileName := strings.Split(fileString, "/")[1]
	return fileName
}

func getRequestTarget(request string) string {
	fmt.Println("req: " + request)
	requestLine := strings.Split(request, "\r\n")[0]
	fmt.Println("RL: " + requestLine)
	target := strings.Split(requestLine, " ")[1]
	return target
}

func getUserAgent(request string) string {
	userAgentLine := strings.Split(request, "\r\n")[2]
	userAgentValue := strings.Split(userAgentLine, " ")[1]
	return userAgentValue
}
