package main

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/briandowns/spinner"
	"github.com/charmbracelet/glamour"
	"github.com/common-nighthawk/go-figure"
	"github.com/mitchellh/go-homedir"
	"github.com/ohxxx/tgpt/store"
	"github.com/ohxxx/tgpt/utils"
	gpt3 "github.com/sashabaranov/go-gpt3"
)

var home, err = homedir.Dir()
var tgptPath = home + "/.tgpt"
var gptStore = store.InitStore(tgptPath)

func init() {
	flag.String("k", "", "Set OpenAI API key")
	flag.String("e", "", "Exporting content ")
}

func main() {
	initFlags()
}

func initFlags() {
	flag.Parse()

	if flag.NFlag() == 0 && flag.NArg() == 0 {
		processQA()
	}

	if flag.NFlag() == 0 && flag.NArg() > 0 {
		actionFlags(flag.CommandLine)
	}
}

func actionFlags(f *flag.FlagSet) {
	if name := f.Arg(0); name != "" {
		switch name {
		case "k":
			apiKey := f.Arg(1)
			if apiKey == "" {
				fmt.Println("No API key.")
				return
			}

			if err := store.SetOpenAiApiKey(tgptPath, apiKey); err != nil {
				fmt.Println(err)
				return
			}
		case "e":
			filename, err := store.GetLatestContent(tgptPath)
			if err != nil {
				fmt.Println("File export failed")
				return
			}
			utils.DownloadFile(tgptPath + "/content/" + filename)
			fmt.Println("The file has been imported into the Downloads file")
		default:
			fmt.Println("No command.")
		}
	}
}

func processQA() {
	config, err := store.GetConfig(tgptPath)
	if err != nil {
		fmt.Println(err)
		return
	}

	if config.OpenAiApiKey == "" {
		fmt.Println("OpenAiApiKey is empty")
		return
	}

	apiKey := config.OpenAiApiKey

	mdRenderer := initGlamour()

	halo()

	GPTClient := createGPTClient(apiKey)

	msg := []gpt3.ChatCompletionMessage{
		{
			Role:    "system",
			Content: "You are my all-purpose assistant. Please answer my questions as accurately and concisely as possible.",
		},
	}

	quitChan := make(chan os.Signal, 1)
	signal.Notify(
		quitChan,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGKILL,
	)

	iptChan := make(chan string)

	go func() {
		for {
			input := getInput()
			if input == "" {
				iptChan <- ""
				return
			}
			iptChan <- input
		}
	}()

	for {
		select {
		case <-quitChan:
			store.UpdateContent(tgptPath, msg)
			return
		case input := <-iptChan:

			done := make(chan bool)
			go func() {
				loading(done)
				done <- true
			}()
			msg = append(msg, gpt3.ChatCompletionMessage{
				Role:    "user",
				Content: input,
			})

			res, err := GPTClient.CreateChatCompletion(
				context.Background(),
				gpt3.ChatCompletionRequest{
					Model:       gpt3.GPT3Dot5Turbo0301,
					Messages:    msg,
					MaxTokens:   2048,
					Temperature: 0,
					N:           1,
				})
			done <- false
			if err != nil {
				fmt.Println(err)
				return
			}

			renderRespone(mdRenderer, res.Choices[0].Message.Content)
			msg = append(
				msg, gpt3.ChatCompletionMessage{
					Role:    "assistant",
					Content: res.Choices[0].Message.Content,
				},
			)
			fmt.Println("================================")
			fmt.Print("Please enter your question: ")
		}

	}
}

func initGlamour() *glamour.TermRenderer {
	renderStyle := glamour.WithEnvironmentConfig()
	mdRenderer, err := glamour.NewTermRenderer(
		renderStyle,
	)
	if err != nil {
		fmt.Println("MD initialization failed.", err)
		return nil
	}
	return mdRenderer
}

func halo() {
	halo := figure.NewFigure("halo", "", true)
	color := "\033[38;2;68;189;135m"
	var logo bytes.Buffer
	for _, char := range halo.String() {
		if char == ' ' {
			logo.WriteByte(' ')
		} else {
			logo.WriteString(fmt.Sprintf("%s%c\033[0m", color, char))
		}
	}
	fmt.Println(logo.String())
}

func loading(done chan bool) {
	for {
		select {
		case <-done:
			return
		default:
			s := spinner.New(spinner.CharSets[24], 100*time.Millisecond)
			s.Prefix = " Making up lies: "
			s.Start()
			time.Sleep(2 * time.Second)
			s.Stop()
		}
	}
}

func createGPTClient(apiKey string) *gpt3.Client {
	GPTClient := gpt3.NewClient(apiKey)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return GPTClient
}

func renderRespone(mdRenderer *glamour.TermRenderer, markdown string) {
	fmt.Println("【GPT】==========================")
	rendered, err := mdRenderer.Render(markdown)
	if err != nil {
		fmt.Println("MD rendering failed.", err)
		return
	}
	fmt.Println(rendered)
}

func getInput() string {
	fmt.Print("Please enter your question: ")
	scanner := bufio.NewScanner(os.Stdin)
	input := ""
	for scanner.Scan() {
		input += scanner.Text()
	}
	return input
}
