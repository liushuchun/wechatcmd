package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path"

	ct "github.com/daviddengcn/go-colortext"
	"github.com/liushuchun/wechatcmd/ui"
	chat "github.com/liushuchun/wechatcmd/wechat"
)

const (
	maxChanSize = 50
)

type Config struct {
	SaveToFile bool     `json:"save_to_file"`
	AutoReply  bool     `json:"auto_reply"`
	ReplyMsg   []string `json:"reply_msg"`
}

func main() {

	ct.Foreground(ct.Green, true)
	flag.Parse()
	logger := log.New(os.Stdout, "[*AI*]:", log.LstdFlags)

	logger.Println("启动...")
	fileName := "main.log"
	var logFile *os.File
	if _, err := os.Stat(fileName); err != nil {
		logFile, err = os.Create(fileName)
		if err != nil {
			logger.Println("创建日志文件失败")
			return
		}
	} else {
		logFile, err = os.Open(fileName)
		if err != nil {
			logger.Println("打开日志文件失败")
			return
		}
	}
	defer logFile.Close()

	wxLogger := log.New(logFile, "[*]", log.LstdFlags)

	wechat := chat.NewWechat(wxLogger)

	if err := wechat.WaitForLogin(); err != nil {
		logger.Fatalf("等待失败：%s\n", err.Error())
		return
	}
	srcPath, err := os.Getwd()
	if err != nil {
		logger.Printf("获得路径失败:%#v", err)
	}
	configFile := path.Join(path.Clean(srcPath), "config.json")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		logger.Fatalln("请提供配置文件：config.json")
		return
	}

	b, err := ioutil.ReadFile(configFile)
	if err != nil {
		logger.Fatalln("读取文件失败：%#v", err)
		return
	}
	var config *Config
	err = json.Unmarshal(b, &config)

	logger.Printf("登陆...")
	if err := wechat.Login(); err != nil {
		logger.Printf("登陆失败：%v\n", err)
		return
	}
	logger.Printf("配置文件:%+v\n", config)

	logger.Println("成功")

	logger.Println("微信初始化成功...")

	logger.Println("开启状态栏通知...")
	if err := wechat.StatusNotify(); err != nil {
		return
	}
	if err := wechat.GetContacts(); err != nil {
		logger.Fatalf("拉取联系人失败:%v\n", err)
		return
	}
	if err := wechat.TestCheck(); err != nil {
		logger.Fatalf("检查状态失败:%v\n", err)
		return
	}

	itemList := []string{}
	logger.Printf("the initcontact:%+v", wechat.InitContactList)
	for _, member := range wechat.InitContactList {
		itemList = append(itemList, member.NickName)
	}
	userList := []string{}
	for _, member := range wechat.PublicUserList {
		userList = append(userList, member.NickName)
	}

	msgIn := make(chan chat.Message, maxChanSize)
	msgOut := make(chan chat.MessageOut, maxChanSize)
	chatIn := make(chan chat.Message, maxChanSize)
	closeChan := make(chan int, 1)

	layout := ui.NewLayout(itemList, userList, chatIn, msgIn, msgOut, closeChan)

	go wechat.SyncDaemon(msgIn)

	go wechat.MsgDaemon(msgOut)

	layout.Init()

}
