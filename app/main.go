package main

import (
	"bytes"
	"compress/gzip"
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
		sendOK(conn)
	} else if strings.HasPrefix(target, "/echo") {
		response := getOkResponseWithTP()
		body := strings.Split(target, "/")[2]
		var finalStringToConvert string
		schemes, err := getEncodingsList(strReq)
		if err == nil {
			response += "Content-Encoding:" + schemes[0] + "\r\n"
			compressedBody := getCompressedBody(body)
			finalStringToConvert = fmt.Sprintf(response+"Content-Length: %d\r\n\r\n%s", len(compressedBody), compressedBody)
		} else {
			finalStringToConvert = fmt.Sprintf(response+"Content-Length: %d\r\n\r\n%s", len(body), body)
		}
		conn.Write([]byte(finalStringToConvert))
	} else if strings.HasPrefix(target, "/user-agent") {
		body := getUserAgent(strReq)
		response := getOkResponseWithTP()
		finalStringToConvert := fmt.Sprintf(response+"Content-Length: %d\r\n\r\n%s", len(body), body)
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
			response := getOkResponseWithOstream()
			finalStringToConvert := fmt.Sprintf(response+"Content-Length: %d\r\n\r\n%s", len(body), body)
			conn.Write([]byte(finalStringToConvert))
		} else if getMethodType(strReq) == "POST" {
			body := getBody(strReq)
			err := os.WriteFile(dir+filename, []byte(body), 0666)
			if err != nil {
				sendNotFound(conn)
				return
			}
			sendCreated(conn)
		}
	} else {
		sendNotFound(conn)
	}
}

func sendNotFound(conn net.Conn) {
	conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
}

func getOkResponseWithTP() string {
	return "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\n"
}

func getOkResponseWithOstream() string {
	return "HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\n"
}

func sendOK(conn net.Conn) {
	conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
}

func sendCreated(conn net.Conn) {
	conn.Write([]byte("HTTP/1.1 201 Created\r\n\r\n"))
}

func getCompressedBody(body string) string {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	w.Write([]byte(body))
	w.Close()
	compressedBody := b.String()
	return compressedBody
}

func getEncodingsList(request string) ([]string, error) {
	var schemesToReturn []string
	schemesLine := strings.Split(request, "\r\n")[2]
	if len(schemesLine) > 0 {
		schemes := strings.Fields(schemesLine[16:])
		for _, v := range schemes {
			v = strings.TrimRight(v, ",")
			if slices.Contains(supportedEncodingSchemes, v) {
				schemesToReturn = append(schemesToReturn, v)
			}
		}
		if len(schemesToReturn) > 0 {
			return schemesToReturn, nil
		}
	}
	return schemesToReturn, errors.New("empty scheme slice, no encoding provided")
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
