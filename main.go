package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	"github.com/daviddengcn/go-colortext"
	"github.com/hawklithm/wechatcmd/ui"
	chat "github.com/hawklithm/wechatcmd/wechat"
)

const (
	maxChanSize = 50
)

type Config struct {
	SaveToFile   bool     `json:"save_to_file"`
	AutoReply    bool     `json:"auto_reply"`
	AutoReplySrc bool     `json:"auto_reply_src"`
	ReplyMsg     []string `json:"reply_msg"`
}

func main() {

	ct.Foreground(ct.Green, true)
	flag.Parse()
	logger := log.New(os.Stdout, "[*🤔 *]->:", log.LstdFlags)

	logger.Println("启动...")
	fileName := "log.txt"
	var logFile *os.File
	logFile, err := os.OpenFile(fileName, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)

	defer logFile.Close()
	if err != nil {
		logger.Printf("打开文件失败!\n")
	}

	wxLogger := log.New(logFile, "[*]", log.LstdFlags)

	wechat := chat.NewWechat(wxLogger)

	if err := wechat.WaitForLogin(); err != nil {
		logger.Fatalf("等待失败：%s\n", err.Error())
		return
	}
	srcPath, err := os.Getwd()
	if err != nil {
		logger.Printf("获得路径失败:%#v\n", err)
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

	logger.Printf("登陆...\n")

	wechat.AutoReplyMode = config.AutoReply
	wechat.ReplyMsgs = config.ReplyMsg
	wechat.AutoReplySrc = config.AutoReplySrc

	if err := wechat.Login(); err != nil {
		logger.Printf("登陆失败：%v\n", err)
		return
	}
	logger.Printf("配置文件:%+v\n", config)

	logger.Println("成功!")

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

	nickNameList := []string{}
	userIDList := []string{}

	for _, member := range wechat.InitContactList {
		nickNameList = append(nickNameList, member.NickName)
		userIDList = append(userIDList, member.UserName)

	}

	for _, member := range wechat.ContactList {
		nickNameList = append(nickNameList, member.NickName)
		userIDList = append(userIDList, member.UserName)
	}

	for _, member := range wechat.PublicUserList {
		nickNameList = append(nickNameList, member.NickName)
		userIDList = append(userIDList, member.UserName)

	}
	groupIdList := []string{}
	for _, user := range userIDList {
		if strings.HasPrefix(user, "@@") {
			groupIdList = append(groupIdList, user)
		}
	}

	//群成员列表
	groupMemberList, err := wechat.GetContactsInBatch(groupIdList)
	if err != nil {
		logger.Fatal("get batch contact error=", err)
		return
	}

	msgIn := make(chan chat.Message, maxChanSize)
	msgOut := make(chan chat.MessageOut, maxChanSize)
	imageIn := make(chan chat.MessageImage, maxChanSize)
	closeChan := make(chan int, 1)
	autoChan := make(chan int, 1)

	go wechat.SyncDaemon(msgIn, imageIn)

	go wechat.MsgDaemon(msgOut, autoChan)
	ui.NewLayout(nickNameList, userIDList, groupMemberList,
		wechat.User.NickName,
		wechat.User.UserName, msgIn, msgOut, imageIn, closeChan, autoChan, wxLogger)

}
