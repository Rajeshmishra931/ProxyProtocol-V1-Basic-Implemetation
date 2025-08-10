package main

import (
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	lbListen := "127.0.0.1:8000"
	backend := "127.0.0.1:9000"

	ln, err := net.Listen("tcp", lbListen)
	if err != nil {
		panic(err)
	}
	fmt.Println("[LB] Listening on", lbListen)

	for {
		clientConn, err := ln.Accept()
		if err != nil {
			fmt.Println("[LB] accept error:", err)
			continue
		}
		go handleConnection(clientConn, backend, lbListen)
	}
}

func handleConnection(clientConn net.Conn, backendAddr, lbAddr string) {
	start := time.Now()
	clientAddr := clientConn.RemoteAddr().String()
	fmt.Println("\n[LB] Accepted from client:", clientAddr)

	serverConn, err := net.Dial("tcp", backendAddr)
	if err != nil {
		fmt.Println("[LB] error connecting to backend:", err)
		clientConn.Close()
		return
	}
	fmt.Println("[LB] Connected to backend:", backendAddr)

	srcHost, srcPort, _ := net.SplitHostPort(clientAddr)
	dstHost, dstPort, _ := net.SplitHostPort(backendAddr) // backend address
	proxyLine := fmt.Sprintf("PROXY TCP4 %s %s %s %s\r\n", srcHost, dstHost, srcPort, dstPort)


	// time.Sleep(4 * time.Second)  // simulate delay
	_, err = serverConn.Write([]byte(proxyLine))
	if err != nil {
		fmt.Println("[LB] failed to write PROXY header:", err)
		clientConn.Close()
		serverConn.Close()
		return
	}
	fmt.Println("[LB] Sent PROXY header to backend:", strings.TrimRight(proxyLine, "\r\n"))

	var bytesC2S int64
	var bytesS2C int64
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		n, _ := io.Copy(serverConn, clientConn)
		atomic.AddInt64(&bytesC2S, n)
		serverConn.Close()
	}()

	go func() {
		defer wg.Done()
		n, _ := io.Copy(clientConn, serverConn)
		atomic.AddInt64(&bytesS2C, n)
		clientConn.Close()
	}()

	wg.Wait()

	elapsed := time.Since(start)
	fmt.Println("[LB Connection Stats]")
	fmt.Println("  Client:", clientAddr)
	fmt.Println("  LB:", lbAddr)
	fmt.Println("  Server:", backendAddr)
	fmt.Println("  Bytes Client to Server:", atomic.LoadInt64(&bytesC2S))
	fmt.Println("  Bytes Server to Client:", atomic.LoadInt64(&bytesS2C))
	fmt.Println("  Total Time:", elapsed)
}
