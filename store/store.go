package store

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/ohxxx/tgpt/utils"
	gpt3 "github.com/sashabaranov/go-gpt3"
)

const (
	configFilename = "config.json"
	contentDirname = "content"
)

type Config struct {
	OpenAiApiKey string `json:"OpenAiApiKey"`
}

func InitStore(path string) error {
	if err := createDirectory(path); err != nil {
		return err
	}
	if err := createConfigFile(path); err != nil {
		return err
	}
	if err := createContentDirectory(path); err != nil {
		return err
	}
	return nil
}

func GetConfig(path string) (*Config, error) {
	configPath := filepath.Join(path, configFilename)
	file, err := os.Open(configPath)
	if err != nil {
		utils.LogFileError("read", path, configFilename)
		return nil, err
	}
	defer file.Close()

	var config Config
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		utils.LogFileError("read", path, configFilename)
		return nil, err
	}
	return &config, nil
}

func createDirectory(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		utils.LogFileError("create", path)
		return err
	}
	return nil
}

func createConfigFile(path string) error {
	configPath := filepath.Join(path, configFilename)
	if _, err := os.Stat(configPath); err == nil {
		return nil
	}

	configJSON, err := json.MarshalIndent(&Config{OpenAiApiKey: ""}, "", "  ")
	if err != nil {
		utils.LogFileError("write")
		return err
	}
	if err := writeToFile(configPath, configJSON); err != nil {
		utils.LogFileError("write", path, configFilename)
		return err
	}
	return nil
}

func createContentDirectory(path string) error {
	contentPath := filepath.Join(path, contentDirname)
	if err := os.MkdirAll(contentPath, 0755); err != nil {
		utils.LogFileError("create", path, contentDirname)
		return err
	}
	return nil
}

func SetOpenAiApiKey(path string, apiKey string) error {
	config, err := GetConfig(path)
	if err != nil {
		return err
	}
	config.OpenAiApiKey = apiKey

	configJSON, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		utils.LogFileError("write")
		return err
	}
	if err := writeToFile(filepath.Join(path, configFilename), configJSON); err != nil {
		utils.LogFileError("write", path, configFilename)
		return err
	}
	return nil
}

func UpdateContent(path string, content []gpt3.ChatCompletionMessage) error {
	timestamp := time.Now().Unix()
	contentPath := filepath.Join(path, contentDirname, fmt.Sprintf("%d", timestamp)+".txt")

	contentJSON, err := json.MarshalIndent(content, "", "  ")
	if err != nil {
		utils.LogFileError("write")
		return err
	}
	if err := writeToFile(contentPath, contentJSON); err != nil {
		utils.LogFileError("write", path, configFilename)
		return err
	}
	return nil
}

func GetLatestContent(path string) (string, error) {
	contentPath := filepath.Join(path, contentDirname)
	files, err := ioutil.ReadDir(contentPath)
	if err != nil {
		return "", err
	}
	if len(files) == 0 {
		return "", nil
	}

	var latestFile os.FileInfo
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if latestFile == nil || file.ModTime().After(latestFile.ModTime()) {
			latestFile = file
		}
	}
	return latestFile.Name(), nil
}

func writeToFile(filePath string, data []byte) error {
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		return err
	}
	return nil
}
