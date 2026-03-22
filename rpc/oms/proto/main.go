package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
)

/*
由于go-zero的goctl生成gprc服务只能是一个proto，服务多的时候 比较乱，所以用这个工具来合成一个
Author: LiuFeiHua
Date: 2024/6/11 10:52
*/
func main() {
	folderPath := "rpc/oms/proto"         // 替换为你的文件夹路径
	outputFilePath := "rpc/oms/oms.proto" // 新文件的路径

	files, err := ioutil.ReadDir(folderPath)
	if err != nil {
		fmt.Println("Error reading directory:", err)
		return
	}

	var fileContents []byte
	var startContents []byte
	for _, file := range files {
		if file.IsDir() || file.Name() == "main.go" || !strings.HasSuffix(file.Name(), ".proto") {
			continue
		}

		fileName := filepath.Join(folderPath, file.Name())
		fileData, err := ioutil.ReadFile(fileName)
		if err != nil {
			fmt.Printf("Error reading file %s: %s\n", fileName, err)
			continue
		}
		if len(startContents) == 0 {
			startContents = []byte(protoHeader(fileData))
		}
		fileData = []byte(trimProtoHeader(fileData))

		fileContents = append(fileContents, fileData...)
	}
	startContents = append(startContents, fileContents...)

	start := strings.Replace(string(startContents), "package main", "package omsclient", 1)
	start = strings.Replace(start, "option go_package = \"./proto\"", "option go_package = \"./omsclient\"", 1)
	err = ioutil.WriteFile(outputFilePath, []byte(start), 0644)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}

	fmt.Println("Merged files to", outputFilePath)
}

func protoHeader(fileData []byte) string {
	lines := strings.Split(string(fileData), "\n")
	header := make([]string, 0, 4)
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "syntax = ") || strings.HasPrefix(trimmed, "package ") || strings.HasPrefix(trimmed, "option go_package = ") || trimmed == "" {
			header = append(header, line)
			if strings.HasPrefix(trimmed, "option go_package = ") {
				header = append(header, "")
				break
			}
		}
	}
	return strings.Join(header, "\n")
}

func trimProtoHeader(fileData []byte) string {
	lines := strings.Split(string(fileData), "\n")
	start := 0
	for start < len(lines) {
		trimmed := strings.TrimSpace(lines[start])
		if strings.HasPrefix(trimmed, "syntax = ") || strings.HasPrefix(trimmed, "package ") || strings.HasPrefix(trimmed, "option go_package = ") || trimmed == "" {
			start++
			continue
		}
		break
	}
	return strings.Join(lines[start:], "\n")
}
