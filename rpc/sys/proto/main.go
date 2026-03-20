package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
)

const protoHeader = "syntax = \"proto3\";\n\npackage main;\n\noption go_package = \"./proto\";\n"

/*
由于go-zero的goctl生成gprc服务只能是一个proto，服务多的时候 比较乱，所以用这个工具来合成一个
Author: LiuFeiHua
Date: 2024/6/11 10:52
*/
func main() {
	folderPath := "rpc/sys/proto"         // 替换为你的文件夹路径
	outputFilePath := "rpc/sys/sys.proto" // 新文件的路径

	files, err := ioutil.ReadDir(folderPath)
	if err != nil {
		fmt.Println("Error reading directory:", err)
		return
	}

	var fileContents []byte
	var wroteHeader bool
	for _, file := range files {
		if file.IsDir() || file.Name() == "main.go" || filepath.Ext(file.Name()) != ".proto" {
			continue
		}

		fileName := filepath.Join(folderPath, file.Name())
		fileData, err := ioutil.ReadFile(fileName)
		if err != nil {
			fmt.Printf("Error reading file %s: %s\n", fileName, err)
			continue
		}

		fileText := string(fileData)
		if !strings.HasPrefix(fileText, protoHeader) {
			fmt.Printf("Skip file %s: unexpected proto header\n", fileName)
			continue
		}

		if !wroteHeader {
			fileContents = append(fileContents, []byte(protoHeader)...)
			wroteHeader = true
		}

		body := strings.TrimPrefix(fileText, protoHeader)
		body = strings.TrimLeft(body, "\n")
		if strings.TrimSpace(body) == "" {
			continue
		}
		if !strings.HasSuffix(body, "\n") {
			body += "\n"
		}

		fileData = []byte("\n" + body)

		fileContents = append(fileContents, fileData...)
	}

	start := strings.Replace(string(fileContents), "package main", "package sysclient", 1)
	start = strings.Replace(start, "option go_package = \"./proto\"", "option go_package = \"./sysclient\"", 1)
	err = ioutil.WriteFile(outputFilePath, []byte(start), 0644)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}

	fmt.Println("Merged files to", outputFilePath)
}
