package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"time"
)

func readRequest(client *net.TCPConn) (string, int, string, error) {
	s := bufio.NewScanner(client)

	s.Scan()
	host := s.Text()

	s.Scan()
	prt := s.Text()
	port, err := strconv.Atoi(prt)

	if err != nil {
		log.Printf("Port parsing error: %s", err.Error())
		return "", 0, "", err
	}

	s.Scan()
	req := s.Text()

	log.Printf("Request %s:%d%s", host, port, req)
	return host, port, req, nil
}

func sendChunk(client *net.TCPConn, reader *bufio.Reader) (int, error) {
	buff := make([]byte, 256)

	writer := bufio.NewWriter(client)

	deadline := time.Now().Add(time.Duration(100) * time.Second)

	client.SetDeadline(deadline)
	n, err := io.ReadFull(reader, buff)

	if err != nil {
		if err != io.EOF {
			log.Printf("Receiving chunk error: %s", err.Error())
		}
	}

	if n == 0 {
		return 0, err
	}

	for i := 0; i < n; i++ {
		writer.WriteByte(buff[i])
	}

	return n, writer.Flush()
}

func processChunks(client *net.TCPConn, remote net.Conn) {
	n := 1
	var err error = nil

	clientReader := bufio.NewReader(client)
	remoteReader := bufio.NewReader(remote)
	for n > 0 {
		n, err = sendChunk(client, remoteReader)

		if err != nil {
			if err != io.EOF {
				log.Printf("Proxing request error: %s", err.Error())
			}

			remote.Close()

			return
		}

		_, err = clientReader.ReadByte()
		if err != nil {
			log.Printf("Lost client with error: %s", err.Error())

			return
		}
	}

	remote.Close()
}

func proxify(connection *net.TCPConn) {
	host, port, req, err := readRequest(connection)

	if err != nil {
		return
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	log.Printf("Connecting to: %s", addr)

	remote, err := net.Dial("tcp", addr)
	if err != nil {
		log.Printf("Connection to remote host error: %s", err.Error())
	}

	fmt.Fprintf(remote, "%s\r\n", req) // Sending request to server

	processChunks(connection, remote)
}
