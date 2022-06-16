package main

import (
	"fmt"
	"io"
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
	p := request(conn)
	response(conn, p)
}

func request(conn net.Conn) string {
	// とりあえずはcommand line引数でpathを指定する方式
	buf := make([]byte, 4096)
	n, _ := conn.Read(buf)
	return string(buf[:n])
}

func response(conn net.Conn, path string) {
	// 絶対パス変換
	path, err := getAbsPath(path)
	if err != nil {
		panic(err)
	}
	fmt.Println(path)
	// 指定したパス以下のディレクトリ(複数)を取得
	paths := dirWalk(path)
	// htmlファイルを返す
	for _, path := range paths {
		if strings.Contains(path, "index.html") {
			c, err := readFileContent(path)
			if err != nil {
				panic(err)
			}
			io.WriteString(conn, string(c))
		}
	}
}

func getAbsPath(path string) (string, error) {
	pwd, err := os.Getwd()
	return pwd + path, err
}

func readFileContent(path string) ([]byte, error) {
	content, err := ioutil.ReadFile(path)
	return content, err
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
