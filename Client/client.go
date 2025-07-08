package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

const (
	START_MODE     = 1
	COMMAND_MODE   = 2
	IN_CHAT_MODE   = 3
	EXIT_MODE      = 4
	SERVER_ADDRESS = "127.0.0.1:8080"
)

var (
	networking  *ClientNetworking
	states      = make(map[int](func()))
	curState    = START_MODE
	curReceiver string
)

func main() {
	states[START_MODE] = handleStartMode
	states[COMMAND_MODE] = handleCommandMode
	states[IN_CHAT_MODE] = handleInChatMode
	states[EXIT_MODE] = handleExitMode
	for {
		states[curState]()
	}
}

func handleStartMode() {
	fmt.Println("Welcome to my messenger. Enter your username")

	var username string
	err := errors.New("")
	for err != nil {
		username = inputString()
		networking, err = GetClient(SERVER_ADDRESS, username)
		if err != nil {
			fmt.Println(err)
		}
	}
	fmt.Println("Succesfully registered on server")
	networking.RunClient()
	go printIncomingMessages()

	curState = COMMAND_MODE
}

func handleCommandMode() {
	for {
		msg := inputString()
		splitted := strings.Split(msg, " ")
		if len(splitted) == 1 {
			switch splitted[0] {
			case "exit":
				curState = EXIT_MODE
				return
			case "help":
				fmt.Println("choose <username>\t selects the user, who will receive your messages\nexit\t\t\t finish the programm")
			default:
				fmt.Println("No such command. To see all available commands, please enter \"help\"")
			}
		} else if len(splitted) == 2 && splitted[0] == "choose" {
			receiver := splitted[1]
			curReceiver = receiver
			curState = IN_CHAT_MODE
			return
		} else {
			fmt.Println("No such command. To see all available commands, please enter \"help\"")
		}
	}
}

func handleInChatMode() {
	fmt.Println("Enter your message to", curReceiver)
	for {
		text := inputString()
		if text == "/back" {
			curState = COMMAND_MODE
			return
		} else {
			networking.sendMessage(text, curReceiver)
		}
	}
}

func handleExitMode() {
	fmt.Println("Exiting..")
	os.Exit(0)
}

func printIncomingMessages() {
	for msg := range networking.Incoming {
		fmt.Println("-----" + msg.Sender + "-----\n" + msg.Data + "\n-----" + strings.Repeat("-", len(msg.Sender)) + "-----")
	}
}

func inputString() string {
	msg, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	return strings.Trim(msg, "\n ")
}
