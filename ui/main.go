package main

import (
	"flag"

	"github.com/liushuchun/wechatcmd/box/chat"

	"github.com/liushuchun/wechatcmd/box/box"
)

var (
	room      = flag.String("room", "", "Room to relax and chat in")
	username  = flag.String("username", "Bernard Minet", "User name")
	brokerURL = flag.String("broker", "", "Kafka broker URL")
	clientID  = flag.String("client-id", "", "Kafka client ID")
)

func main() {
	flag.Parse()
	maxChanSize := 10000

	//log.SetLevel(log.DebugLevel)
	msgIn := make(chan chat.Message, maxChanSize)
	textOut := make(chan string, maxChanSize)

	layout := box.NewLayout(msgIn, textOut)
	layout.Init()
}
