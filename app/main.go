package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}

	req := make([]byte, 1024)
	conn.Read(req)

	// if !strings.HasPrefix(string(req), "GET / HTTP/1.1") {
	// 	conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	// 	return
	// }
	body := string(req)[10:]
	body = body[:len(body)-9]
	number := len(body)
	fmt.Println(body)

	var finalStringToConvert string = "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: " + strconv.Itoa(number) + "\r\n" + body
	conn.Write([]byte(finalStringToConvert))
}
