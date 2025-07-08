package main

import (
	"encoding/json"
	"errors"
	"log"
	"net"
)

const (
	ERROR    = "error"
	OK       = "ok"
	REGISTER = "register"
	MESSAGE  = "message"
)

type Message struct {
	Type     string `json:"type"`
	Data     string `json:"data"`
	Sender   string `json:"sender"`
	Receiver string `json:"receiver"`
}

type ClientNetworking struct {
	Username string
	Conn     net.Conn
	Outgoing chan Message
	Incoming chan Message
}

func GetClient(addr string, username string) (*ClientNetworking, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	registerMsg := Message{Type: REGISTER, Sender: username}
	if err := json.NewEncoder(conn).Encode(registerMsg); err != nil {
		return nil, err
	}

	var registerResponse Message
	if err := json.NewDecoder(conn).Decode(&registerResponse); err != nil {
		return nil, err
	}

	if registerResponse.Type == ERROR {
		return nil, errors.New(registerResponse.Data)
	}

	return &ClientNetworking{
		Username: username,
		Conn:     conn,
		Outgoing: make(chan Message, 2),
		Incoming: make(chan Message, 2),
	}, nil
}

func (this *ClientNetworking) RunClient() {
	go this.writeLoop()
	go this.readLoop()
}

func (this *ClientNetworking) writeLoop() {
	encoder := json.NewEncoder(this.Conn)
	for msg := range this.Outgoing {
		if err := encoder.Encode(msg); err != nil {
			log.Printf("Write error %v", err)
			return
		}
	}
}

func (this *ClientNetworking) readLoop() {
	decoder := json.NewDecoder(this.Conn)
	for {
		var msg Message
		if err := decoder.Decode(&msg); err != nil {
			log.Printf("Read error %v", err)
			return
		}
		this.Incoming <- msg
	}
}

func (this *ClientNetworking) sendMessage(text string, receiver string) {
	msg := Message{
		Type:     MESSAGE,
		Data:     text,
		Sender:   this.Username,
		Receiver: receiver,
	}
	this.Outgoing <- msg
}
