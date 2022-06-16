package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	listner, err := net.Listen("tcp", "localhost:8000")

	if err != nil {
		panic(err)
	}

	for {
		conn, err := listner.Accept()

		if err != nil {
			panic(err)
		}

		go handlerConn(conn)
	}
}

func handlerConn(conn net.Conn) {
	defer conn.Close()
	EnableKeepAlive(conn)
	p := request(conn)
	response(conn, p)
}

func request(conn net.Conn) string {
	scanner := bufio.NewScanner(conn)
	i := 0
	var path string
	for scanner.Scan() {
		line := scanner.Text()
		if i == 0 {
			firstLine := strings.Fields(line)
			path = firstLine[1]
		}
		if line == "" {
			break
		}
		i++
	}
	return path
}

func response(conn net.Conn, path string) {
	// 絶対パス変換
	path, err := getAbsPath(path)
	if err != nil {
		panic(err)
	}
	// 指定したパス以下のディレクトリ(複数)を取得
	paths, err := underDir(path)
	if err != nil {
		show404Page(conn)
	}
	// htmlファイルを返す
	htmlExists := false
	for _, path := range paths {
		if strings.Contains(path, "index.html") {
			c, err := getFileContent(path)
			if err != nil {
				panic(err)
			}
			writeContent(conn, c, 200)
			htmlExists = true
		}
	}
	if !htmlExists {
		show404Page(conn)
	}
}

func getAbsPath(path string) (string, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	err = preventTraversalAttack(path)
	if err != nil {
		return "", err
	}
	return pwd + path, nil
}

func preventTraversalAttack(path string) error {
	if strings.Contains(path, "..") {
		return errors.New("Directly Traversal Attack has been detected")
	}
	return nil
}

func getFileContent(path string) ([]byte, error) {
	content, err := ioutil.ReadFile(path)
	return content, err
}

func writeContent(conn net.Conn, content []byte, status_code int) {
	fmt.Fprintf(conn, "HTTP/1.1 %v OK\r\n", status_code)
	fmt.Fprintf(conn, "Content-Length: %d\r\n", len(content))
	fmt.Fprint(conn, "Content-Type: text/html\r\n")
	fmt.Fprint(conn, "Connection: keep-alive\r\n")
	fmt.Fprint(conn, "\r\n")
	fmt.Fprint(conn, string(content))
}

func show404Page(conn net.Conn) {
	c, err := getFileContent("error.html")
	if err != nil {
		panic(err)
	}
	writeContent(conn, c, 404)
}

func underDir(dir string) ([]string, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var paths []string
	for _, file := range files {
		paths = append(paths, filepath.Join(dir, file.Name()))
	}
	return paths, nil
}

type Conn struct {
	*net.TCPConn
	fd int
}

func EnableKeepAlive(conn net.Conn) (*Conn, error) {
	tcp, ok := conn.(*net.TCPConn)
	if !ok {
		return nil, fmt.Errorf("Bad conn type: %T", conn)
	}
	if err := tcp.SetKeepAlive(true); err != nil {
		return nil, err
	}
	file, err := tcp.File()
	if err != nil {
		return nil, err
	}
	fd := int(file.Fd())
	return &Conn{TCPConn: tcp, fd: fd}, nil
}
