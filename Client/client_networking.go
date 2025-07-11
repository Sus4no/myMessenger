package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"time"
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
	caCert, err := os.ReadFile("crypto/server.crt")
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate: %v", err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, errors.New("failed to parse CA certificate")
	}

	config := &tls.Config{
		RootCAs:            caCertPool,
		InsecureSkipVerify: false,
		ServerName:         "localhost",
		MinVersion:         tls.VersionTLS12,
	}

	config.VerifyConnection = func(cs tls.ConnectionState) error {
		if len(cs.PeerCertificates) == 0 {
			return errors.New("no peer certificates")
		}

		cert := cs.PeerCertificates[0]
		opts := x509.VerifyOptions{
			Roots:       caCertPool,
			CurrentTime: time.Now(),
			KeyUsages:   []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		}

		sanVerified := false
		for _, name := range cert.DNSNames {
			if name == config.ServerName {
				sanVerified = true
				break
			}
		}

		for _, ip := range cert.IPAddresses {
			if ip.String() == config.ServerName {
				sanVerified = true
				break
			}
		}

		if !sanVerified {
			return fmt.Errorf("server name '%s' not found in SANs", config.ServerName)
		}

		if _, err := cert.Verify(opts); err != nil {
			return fmt.Errorf("certificate verification failed: %v", err)
		}

		return nil
	}

	conn, err := tls.Dial("tcp", addr, config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %v", err)
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
