package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
)

var store = map[string]string{
	"k1": "value1",
	"k2": "value2",
	"k3": "value3",
}

func main() {
	addr := "127.0.0.1:9000"
	fmt.Println("[Server] Listening on", addr)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("[Server] accept error:", err)
			continue
		}
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()
	remote := conn.RemoteAddr().String()
	fmt.Println("[Server] Accepted connection from", remote)

	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	reader := bufio.NewReader(conn)

	// Read header
	line, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("[Server] error reading proxy header:", err)
		return
	}
	fmt.Printf("[Server] RAW Header Received (%d bytes): %q\n", len(line), line)

	if !strings.HasSuffix(line, "\r\n") {
		fmt.Println("[Server] invalid proxy header ending (must be CRLF). closing.")
		return
	}
	header := strings.TrimRight(line, "\r\n")

	if !strings.HasPrefix(header, "PROXY TCP4 ") {
		fmt.Println("[Server] invalid or unsupported PROXY header:", header)
		return
	}
	fmt.Println("[Server] Parsed PROXY header:", header)

	conn.SetReadDeadline(time.Time{})

	// Serve commands
	for {
		cmdLine, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("[Server] connection closed or read error:", err)
			return
		}
		fmt.Printf("[Server] RAW Command Received (%d bytes): %q\n", len(cmdLine), cmdLine)

		cmd := strings.TrimRight(cmdLine, "\r\n")
		cmd = strings.TrimSpace(cmd)
		if cmd == "" {
			continue
		}

		if strings.HasPrefix(strings.ToUpper(cmd), "GET ") {
			parts := strings.SplitN(cmd, " ", 2)
			if len(parts) != 2 || parts[1] == "" {
				resp := "ERR Invalid GET\r\n"
				conn.Write([]byte(resp))
				fmt.Printf("[Server] Sent (%d bytes): %q\n", len(resp), resp)
				continue
			}
			key := strings.TrimSpace(parts[1])
			var resp string
			if val, ok := store[key]; ok {
				resp = fmt.Sprintf("Hey the value is %s\r\n", val)
			} else {
				resp = "Unknown key\r\n"
			}
			conn.Write([]byte(resp))
			fmt.Printf("[Server] Sent (%d bytes): %q\n", len(resp), resp)
		} else {
			resp := "ERR Unknown command\r\n"
			conn.Write([]byte(resp))
			fmt.Printf("[Server] Sent (%d bytes): %q\n", len(resp), resp)
		}
	}
}
