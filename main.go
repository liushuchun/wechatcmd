package main

import (
	"flag"
	"log"
	"os"

	ct "github.com/daviddengcn/go-colortext"
	"github.com/liushuchun/wechatcmd/ui"
	chat "github.com/liushuchun/wechatcmd/wechat"
)

func main() {
	const (
		maxChanSize = 50
	)
	ct.Foreground(ct.Yellow, true)
	flag.Parse()
	logger := log.New(os.Stdout, "[**AI**]:", log.LstdFlags)

	logger.Println("启动...")
	wechat := chat.NewWechat(logger)

	if err := wechat.WaitForLogin(); err != nil {
		wechat.Log.Fatalf("等待失败：%s\n", err.Error())
		return
	}

	wechat.Log.Printf("登陆...")
	if err := wechat.Login(); err != nil {
		wechat.Log.Printf("登陆失败：%v\n", err)
		return
	}

	wechat.Log.Println("成功")

	wechat.Log.Println("微信初始化成功...")

	wechat.Log.Println("开启状态栏通知...")
	if err := wechat.StatusNotify(); err != nil {

		return
	}
	if err := wechat.GetContacts(); err != nil {
		wechat.Log.Fatalf("拉取联系人失败%v\n", err)
		return
	}

	itemList := []string{}
	wechat.Log.Printf("the initcontact:%+v", wechat.InitContactList)
	for _, member := range wechat.InitContactList {
		itemList = append(itemList, member.NickName)
	}
	userList := []string{}
	for _, member := range wechat.PublicUserList {
		userList = append(userList, member.NickName)
	}

	msgIn := make(chan chat.Message, maxChanSize)
	textOut := make(chan string, maxChanSize)
	chatIn := make(chan chat.Message, maxChanSize)

	layout := ui.NewLayout(itemList, userList, chatIn, msgIn, textOut)
	layout.Init()
}
