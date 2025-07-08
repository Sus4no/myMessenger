package main

import (
	"encoding/json"
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

type ServerNetworking struct {
	Listener   net.Listener
	Outgoing   chan OutgoingMessage
	Incoming   chan IncomingMessage
	UserToConn map[string]net.Conn
	ConnToUser map[net.Conn]string
}

type IncomingMessage struct {
	Conn net.Conn
	Msg  Message
}

type OutgoingMessage struct {
	Conn net.Conn
	Msg  Message
}

func GetServer(addr string) (*ServerNetworking, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	return &ServerNetworking{
		Listener:   listener,
		Outgoing:   make(chan OutgoingMessage, 2),
		Incoming:   make(chan IncomingMessage, 2),
		UserToConn: make(map[string]net.Conn),
		ConnToUser: make(map[net.Conn]string),
	}, nil
}

func (this *ServerNetworking) RunServer() {
	go this.writeLoop()
	go this.messageHandler()

	for {
		conn, err := this.Listener.Accept()
		if err != nil {
			log.Printf("Accept error: %v", err)
			continue
		}
		go this.readLoop(conn)
	}
}

func (this *ServerNetworking) readLoop(conn net.Conn) {
	defer func() {
		conn.Close()
		if user, ok := this.ConnToUser[conn]; ok {
			delete(this.ConnToUser, conn)
			delete(this.UserToConn, user)
			log.Printf("Client %s disconnected", user)
		}
	}()

	decoder := json.NewDecoder(conn)
	for {
		var msg Message
		if err := decoder.Decode(&msg); err != nil {
			log.Printf("Read error %v", err)
			return
		}
		this.Incoming <- IncomingMessage{
			Conn: conn,
			Msg:  msg,
		}
	}
}

func (this *ServerNetworking) writeLoop() {
	for outMsg := range this.Outgoing {
		encoded, err := json.Marshal(outMsg.Msg)
		if err != nil {
			log.Printf("Encode error %v", err)
			continue
		}
		if _, err := outMsg.Conn.Write(encoded); err != nil {
			log.Printf("Write error %v", err)
		}
	}
}

func (this *ServerNetworking) messageHandler() {
	for incMsg := range this.Incoming {
		conn := incMsg.Conn
		msg := incMsg.Msg

		switch msg.Type {
		case REGISTER:
			userId := msg.Sender

			if userId == "" {
				this.Outgoing <- OutgoingMessage{
					Conn: conn,
					Msg: Message{
						Type: ERROR,
						Data: "Empty userID",
					},
				}
				continue
			}

			if _, ok := this.UserToConn[userId]; ok {
				this.Outgoing <- OutgoingMessage{
					Conn: conn,
					Msg: Message{
						Type: ERROR,
						Data: "UserId already exists",
					},
				}
				continue
			}

			this.UserToConn[msg.Sender] = conn
			this.ConnToUser[conn] = userId

			this.Outgoing <- OutgoingMessage{
				Conn: conn,
				Msg: Message{
					Type: OK,
					Data: "",
				},
			}
			log.Printf("Client %s registered", userId)

		case MESSAGE:
			senderId, ok := this.ConnToUser[conn]
			if !ok {
				this.Outgoing <- OutgoingMessage{
					Conn: conn,
					Msg: Message{
						Type: ERROR,
						Data: "You didn't registered",
					},
				}
				continue
			}

			receiverConn, ok := this.UserToConn[msg.Receiver]
			if !ok {
				this.Outgoing <- OutgoingMessage{
					Conn: conn,
					Msg: Message{
						Type: ERROR,
						Data: "Receiver does not exist",
					},
				}
				continue
			}

			msg.Sender = senderId
			this.Outgoing <- OutgoingMessage{
				Conn: receiverConn,
				Msg:  msg,
			}

		default:
			this.Outgoing <- OutgoingMessage{
				Conn: conn,
				Msg: Message{
					Type: ERROR,
					Data: "Unknown message type",
				},
			}
		}
	}
}
