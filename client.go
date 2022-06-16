package main

import (
	"fmt"
	"net"
	"os"
)

func client(b []byte) (string, error) {
	conn, err := net.Dial("tcp", ":8000")
	if err != nil {
		fmt.Printf("connection error: %v", err)
		return "", err
	}

	defer conn.Close()

	ln, err := conn.Write(b)
	if err != nil {
		fmt.Printf("writing error: %v", err)
		fmt.Println(ln)
		return "", err
	}

	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Printf("Error: %v", err)
		return "", err
	}

	return string(buf[:n]), nil

}

func main() {
	if len(os.Args) != 2 {
		fmt.Println(os.Args)
		fmt.Println("Args length error")
		return
	}
	msg := os.Args[1]

	res, err := client([]byte(msg))
	if err != nil {
		fmt.Printf("Err: %v", err)
		return
	}

	fmt.Println(res)
}
