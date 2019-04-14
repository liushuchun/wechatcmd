package ui

import (
	"flag"
	"log"
	"os"
	"testing"

	"github.com/hawklithm/wechatcmd/wechat"
)

func Test_UI(t *testing.T) {
	flag.Parse()
	maxChanSize := 10000

	//log.SetLevel(log.DebugLevel)
	msgIn := make(chan wechat.Message, maxChanSize)
	textOut := make(chan wechat.MessageOut, maxChanSize)
	imageIn := make(chan wechat.MessageImage, maxChanSize)
	initList := []string{"普罗米修斯", "啊琉球私", "盗火者", "拉风小丸子", "自强不吸"}
	userList := initList

	closeChan := make(chan int, 1)
	autoChan := make(chan int, 1)

	fileName := "log.txt"
	var logFile *os.File
	logFile, err := os.OpenFile(fileName, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)

	defer logFile.Close()
	if err != nil {
		print("打开文件失败!\n")
	}

	wxLogger := log.New(logFile, "[*]", log.LstdFlags)

	NewLayout(initList, userList, []wechat.Member{}, "myName", "12235235",
		msgIn, textOut, imageIn,
		closeChan, autoChan, wxLogger)
}
