package main

import (
	"fmt"
	"log"
	"net"
)

type Message struct {
	from    string
	payload []byte
}

type Server struct {
	ln         net.Listener
	quitch     chan struct{}
	msgch      chan Message
	listenAddr string
}

func NewServer(listenAddr string) *Server {
	return &Server{
		listenAddr: listenAddr,
		quitch:     make(chan struct{}),
		msgch:      make(chan Message, 10),
	}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		return err
	}
	defer ln.Close()
	s.ln = ln

	go s.acceptLoop()

	<-s.quitch
	close(s.msgch)

	return nil
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			fmt.Printf("accept error: %s", err)
			break
		}

		fmt.Printf("new connection to the server: %s \n", conn.RemoteAddr())

		go s.readLoop(conn)
	}
}

func (s *Server) readLoop(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 2048)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Printf("read error: %s", err)
			break
		}

		msg := buf[:n]

		s.msgch <- Message{
			from:    conn.RemoteAddr().String(),
			payload: msg,
		}

		conn.Write([]byte("message recieved!\n"))

	}
}

func main() {
	server := NewServer(":3000")

	go func() {
		for msg := range server.msgch {
			fmt.Printf("recieved message from connection (%s): %s", msg.from, string(msg.payload))
		}
	}()

	log.Fatal(server.Start())
}
