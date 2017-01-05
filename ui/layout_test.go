package ui

import (
	"flag"

	"github.com/liushuchun/wechatcmd/wechat"
)

func Test_UI() {
	flag.Parse()
	maxChanSize := 10000

	//log.SetLevel(log.DebugLevel)
	msgIn := make(chan wechat.Message, maxChanSize)
	textOut := make(chan string, maxChanSize)
	initList := []string{"普罗米修斯", "啊琉球私", "盗火者", "拉风小丸子", "自强不吸"}
	userList := initList

	layout := NewLayout(initList, userList, msgIn, textOut, nil)
	layout.Init()
}
