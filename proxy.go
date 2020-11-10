package main

import (
	"flag"
	"log"
	"net"
)

const version = "0.1"

var serverAddr = flag.String("a", "localhost:6912", "local address")

func handleConn(in <-chan *net.TCPConn, out chan<- *net.TCPConn) {
	for conn := range in {
		log.Printf("Handling %s", conn.RemoteAddr().String())
		proxify(conn)
		out <- conn
	}
}

func closeConn(in <-chan *net.TCPConn) {
	for conn := range in {
		log.Printf("Closing %s", conn.RemoteAddr().String())
		conn.Close()
	}
}

func main() {
	flag.Parse()

	log.Printf("Proxy Next. v %s started", version)
	log.Printf("Openning: %s", *serverAddr)

	addr, err := net.ResolveTCPAddr("tcp", *serverAddr)

	if err != nil {
		panic(err)
	}

	listener, err := net.ListenTCP("tcp", addr)

	if err != nil {
		panic(err)
	}

	log.Println("Port opened successfuly")

	pending, complete := make(chan *net.TCPConn), make(chan *net.TCPConn)

	for i := 0; i < 10; i++ {
		go handleConn(pending, complete)
	}

	go closeConn(complete)

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			panic(err)
		}

		pending <- conn
	}

}
