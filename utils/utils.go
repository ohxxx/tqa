package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	gpt3 "github.com/sashabaranov/go-gpt3"
)

func LogFileError(action string, info ...string) {
	path := info[0]
	filename := ""
	if len(info) > 1 {
		filename = info[1]
	}
	fmt.Printf("Failed to %s config file: %s/%s\n", action, path, filename)
}

func DownloadFile(filePath string) {
	fileContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	var messages []gpt3.ChatCompletionMessage
	err = json.Unmarshal(fileContent, &messages)
	if err != nil {
		fmt.Printf("Error parsing json: %v\n", err)
		return
	}

	filename := strings.Split(filepath.Base(filePath), ".")[0]

	var md strings.Builder
	md.WriteString("# TGPT \n\n")
	for _, msg := range messages {
		if msg.Role == "user" {
			md.WriteString("## Question \n> " + msg.Content + "\n\n")
		}
		if msg.Role == "assistant" {
			md.WriteString("## Answer \n> " + msg.Content + "\n\n")
		}
	}

	downloadDir := filepath.Join(os.Getenv("HOME"), "Downloads")
	err = os.MkdirAll(downloadDir, os.ModePerm)
	if err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		return
	}

	downloadFile, err := os.Create(filepath.Join(downloadDir, "TGPT_"+filename+".md"))
	if err != nil {
		fmt.Printf("Error creating file: %v\n", err)
		return
	}
	defer downloadFile.Close()

	_, err = downloadFile.WriteString(md.String())
	if err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		return
	}
}
