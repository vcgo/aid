package main

import (
	"fmt"
	"net"

	"os"
	"time"

	hook "github.com/robotn/gohook"
)

var clients []net.Conn
var writeStr, readStr = make([]byte, 1024), make([]byte, 1024)

func main() {
	var (
		host   = "localhost"
		port   = "7077"
		remote = host + ":" + port
		data   = make([]byte, 1024)
	)
	fmt.Println("Initiating server...")

	lis, err := net.Listen("tcp", remote)
	defer lis.Close()

	if err != nil {
		fmt.Printf("Error when listen: %s, Err: %s\n", remote, err)
		os.Exit(-1)
	}

	go func() {
		for {
			var res string
			conn, err := lis.Accept()
			if err != nil {
				fmt.Println("Error accepting client: ", err.Error())
				os.Exit(0)
			}
			clients = append(clients, conn)

			go func(con net.Conn) {
				fmt.Println("New connection: ", con.RemoteAddr())

				// Get client's name
				name := "master"

				// Begin recieve message from client
				for {
					length, err := con.Read(data)
					if err != nil {
						fmt.Printf("Client %s quit.\n", name)
						con.Close()
						disconnect(con, name)
						return
					}
					res = string(data[:length])
					sprdMsg := name + ": " + res
					fmt.Println(sprdMsg)
					res = "You said:" + res
					con.Write([]byte(res))
					notify(con, sprdMsg)
				}
			}(conn)
		}
	}()

	time.Sleep(time.Duration(999) * time.Millisecond)
	con, err := net.Dial("tcp", remote)
	defer con.Close()

	hooks := make(chan string, 1)
	go func() {
		s := hook.Start()
		defer hook.End()
		for ev := range s {
			hooks <- ev.String()
		}
	}()

	for {
		in, err := con.Write([]byte(<-hooks))
		if err != nil {
			fmt.Printf("Error when send to server: %d\n", in)
			os.Exit(0)
		}
	}
}

func notify(conn net.Conn, msg string) {
	for _, con := range clients {
		if con.RemoteAddr() != conn.RemoteAddr() {
			con.Write([]byte(msg))
		}
	}
}

func disconnect(conn net.Conn, name string) {
	for index, con := range clients {
		if con.RemoteAddr() == conn.RemoteAddr() {
			disMsg := name + " has left the room."
			fmt.Println(disMsg)
			clients = append(clients[:index], clients[index+1:]...)
			notify(conn, disMsg)
		}
	}
}
