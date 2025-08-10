package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	lbAddr := "127.0.0.1:8000"
	stdin := bufio.NewReader(os.Stdin)

	var conn net.Conn
	var netReader *bufio.Reader

	for {
		if conn == nil {
			fmt.Print("Enter key to GET (k1,k2,k3 or other). Empty to skip: ")
			key, _ := stdin.ReadString('\n')
			key = strings.TrimSpace(key)
			if key == "" {
				continue
			}

			c, err := net.Dial("tcp", lbAddr)
			if err != nil {
				fmt.Println("[Client] dial error:", err)
				continue
			}
			conn = c
			netReader = bufio.NewReader(conn)

			sendGet(conn, key)
			resp, _ := netReader.ReadString('\n')
			fmt.Printf("[Client] Response: %s\n", stripCRLF(resp))
		} else {
			fmt.Print("\nDo you want to kill current connection? (y/n): ")
			ans, _ := stdin.ReadString('\n')
			ans = strings.ToLower(strings.TrimSpace(ans))

			if ans == "y" {
				conn.Close()
				conn = nil
				netReader = nil
				fmt.Println("[Client] Connection closed.")
				continue
			}

			fmt.Print("Enter next key to GET: ")
			key, _ := stdin.ReadString('\n')
			key = strings.TrimSpace(key)
			if key == "" {
				continue
			}
			sendGet(conn, key)
			resp, err := netReader.ReadString('\n')
			if err != nil {
				fmt.Println("[Client] read error (server/connection closed):", err)
				conn.Close()
				conn = nil
				netReader = nil
				continue
			}
			fmt.Printf("[Client] Response: %s\n", stripCRLF(resp))
		}
	}
}

func sendGet(conn net.Conn, key string) {
	fmt.Fprintf(conn, "GET %s\r\n", key)
}

func stripCRLF(s string) string {
	return strings.TrimRight(s, "\r\n")
}
