package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"tinysql-cs/protocol"
	"tinysql-cs/util"
)

const (
	nmHost                   = "H"
	nmPort                   = "P"
	nmUser					 = "U"
)

var (
	host             = flag.String(nmHost, "127.0.0.1", "tidb server host")
	port             = flag.String(nmPort, "4000", "tidb server port")
	user             = flag.String(nmUser, "root", "tidb server port")
)


func main(){
	flag.Parse()

	client, err := NewClient()

	if err != nil {
		fmt.Println(os.Stderr, err.Error())
		os.Exit(1)
	}

	if err := client.Handshake(); err !=nil {
		fmt.Println(os.Stderr, err.Error())
		os.Exit(1)
	}

	fmt.Println(os.Stderr, "successfully connected to client")

	OnCmd(client)
	return
}

func OnCmd(client *Client) {
	defer client.Close()

	for {
		str, err := client.reader.ReadString(';')

		if err != nil {
			fmt.Println(os.Stderr, err.Error())
			continue
		}

		content := strings.ToLower(strings.Trim(str, "\r\n"))

		if strings.EqualFold(content, "exit;"){
			client.SendMessage("", protocol.ExitMessageType)
			os.Exit(0)
		}

		err = client.SendMessage(content, protocol.QueryMessageType)
		if err != nil {
			fmt.Println(os.Stderr, err.Error())
			continue
		}

		reply, err := client.ReceiveMessage()
		if err != nil {
			fmt.Println(os.Stderr, err.Error())
			continue
		}

		fmt.Println(os.Stderr, util.String(reply.Content))
	}
}

