package main

import (
	"fmt"
	"log"
	"net"
	"sort"
	"tcp_to_http/internal/request"
)

func main() {
	ln, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatal("failed to listen to the localhost: ", err)
	}
	defer ln.Close()

	fmt.Println("Listening on :42069...")

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal("failed to accept connection: ", err)
		}
		fmt.Println("connection accepted")

		go func(c net.Conn) {
			defer c.Close()
			req, err := request.RequestFromReader(c)
			if err != nil {
				log.Println("failed to read request: ", err)
				return
			}
			fmt.Println("Request line:")
			fmt.Printf("- Method: %s\n", req.RequestLine.Method)
			fmt.Printf("- Target: %s\n", req.RequestLine.RequestTarget)
			fmt.Printf("- Version: %s\n", req.RequestLine.HttpVersion)

			fmt.Println("Headers:")
			// Sort keys to ensure stable output ordering
			keys := make([]string, 0, len(req.Headers))
			for k := range req.Headers {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			for _, k := range keys {
				fmt.Printf("- %s: %s\n", k, req.Headers[k])
			}
			fmt.Printf("Body: %s\n", string(req.Body))
		}(conn)
	}
}
