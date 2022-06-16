package main

import (
	"bufio"
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
			// method := firstLine[0]
			path = firstLine[1]
			// proto = firstLine[2]
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
	paths := dirWalk(path)
	// htmlファイルを返す
	for _, path := range paths {
		if strings.Contains(path, "index.html") {
			c, err := getFileContent(path)
			if err != nil {
				panic(err)
			}
			writeContent(conn, c)
		}
	}
}

func getAbsPath(path string) (string, error) {
	pwd, err := os.Getwd()
	return pwd + path, err
}

func getFileContent(path string) ([]byte, error) {
	content, err := ioutil.ReadFile(path)
	return content, err
}

func writeContent(conn net.Conn, content []byte) {
	fmt.Fprint(conn, "HTTP/1.1 200 OK\r\n")
	fmt.Fprintf(conn, "Content-Length: %d\r\n", len(content))
	fmt.Fprint(conn, "Content-Type: text/html\r\n")
	fmt.Fprint(conn, "Connection: keep-alive\r\n")
	fmt.Fprint(conn, "\r\n")
	fmt.Fprint(conn, string(content))
}

func dirWalk(dir string) []string {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	var paths []string
	for _, file := range files {
		if file.IsDir() {
			paths = append(paths, dirWalk(filepath.Join(dir, file.Name()))...)
			continue
		}
		paths = append(paths, filepath.Join(dir, file.Name()))
	}
	return paths
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
