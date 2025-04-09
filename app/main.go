package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"slices"
	"strings"
)

var supportedEncodingSchemes = []string{"gzip"}

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

	target := getRequestTarget(strReq)

	if target == "/" {
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	} else if strings.HasPrefix(target, "/echo") {
		response := "HTTP/1.1 200 OK\r\n"
		schemes, err := getEncodingsList(strReq)
		fmt.Println("SCHEMEA: ", schemes)
		if err == nil {
			response += "Content-Encoding:" + schemes[0] + "\r\n"
		}
		body := strings.Split(target, "/")[2]
		finalStringToConvert := fmt.Sprintf(response+"Content-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(body), body)
		conn.Write([]byte(finalStringToConvert))
	} else if strings.HasPrefix(target, "/user-agent") {
		body := getUserAgent(strReq)
		finalStringToConvert := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(body), body)
		conn.Write([]byte(finalStringToConvert))
	} else if strings.HasPrefix(target, "/files/") {
		// /tmp/data/codecrafters.io/http-server-tester/ from /tmp/codecrafters-build-http-server-go --directory /tmp/data/codecrafters.io/http-server-tester/
		dir := os.Args[2]
		filename := getFilename(strReq)
		if getMethodType(strReq) == "GET" {
			body, err := os.ReadFile(dir + filename)
			if err != nil {
				sendNotFound(conn)
				return
			}
			finalStringToConvert := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", len(body), body)
			conn.Write([]byte(finalStringToConvert))
		} else if getMethodType(strReq) == "POST" {
			body := getBody(strReq)
			err := os.WriteFile(dir+filename, []byte(body), 0666)
			if err != nil {
				sendNotFound(conn)
				return
			}
			conn.Write([]byte("HTTP/1.1 201 Created\r\n\r\n"))
		}
	} else {
		sendNotFound(conn)
	}
}

func sendNotFound(conn net.Conn) {
	conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
}

func filterValidSchemes(schemes []string) ([]string, error) {
	var validSchemes []string
	for _, v := range schemes {
		if slices.Contains(supportedEncodingSchemes, v) {
			validSchemes = append(validSchemes, v)
		}
	}
	if len(validSchemes) > 0 {
		return validSchemes, nil
	}
	return nil, errors.New("empty scheme slice, no encoding provided")
}

func getEncodingsList(request string) ([]string, error) {
	var schemesToReturn []string
	var err error
	schemesLine := strings.Split(request, "\r\n")[2]
	fmt.Println("FSHEM LINE", schemesLine)
	if len(schemesLine) > 0 {
		schemes := strings.Fields(schemesLine[16:])
		schemesToReturn, err = filterValidSchemes(schemes)
	}
	return schemesToReturn, err
}

func getMethodType(request string) string {
	methodLine := strings.Split(request, "\r\n")[0]
	methodType := strings.Split(methodLine, " ")[0]
	return methodType
}

func getFilename(request string) string {
	fileString := strings.Split(request, " ")[1]
	fileName := strings.Split(fileString, "/")[2]
	return fileName
}

func getRequestTarget(request string) string {
	requestLine := strings.Split(request, "\r\n")[0]
	target := strings.Split(requestLine, " ")[1]
	return target
}

func getUserAgent(request string) string {
	userAgentLine := strings.Split(request, "\r\n")[2]
	userAgentValue := strings.Split(userAgentLine, " ")[1]
	return userAgentValue
}

func getBody(request string) string {
	bodyLine := strings.Split(request, "\r\n")[5]
	bodyMessage := strings.Split(bodyLine, "\x00")[0]
	return bodyMessage
}
