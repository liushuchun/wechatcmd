package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	ct "github.com/daviddengcn/go-colortext"
	ui "github.com/gizak/termui"
	"github.com/liushuchun/wechatcmd/wechat"
)

func main() {
	ct.Foreground(ct.Yellow, true)
	flag.Parse()
	logger := log.New(os.Stdout, "[**AI**]:", log.LstdFlags)

	logger.Println("天启元年")
	wechat := wechat.NewWechat(logger)

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

	for index, member := range wechat.MemberList {
		item := fmt.Sprintf("[%d] %s  ", index, member.NickName)
		itemList = append(itemList, item)

	}

	err := ui.Init()
	if err != nil {
		panic(err)
	}

	defer ui.Close()
	p := ui.NewPar("欢迎进入微信聊天")
	p.Height = 3
	p.Width = 100
	p.TextFgColor = ui.ColorGreen
	ui.Render(p)
	p.BorderLabel = "welcome"
	p.BorderFg = ui.ColorCyan

	listNum := (len(itemList)-1)/50 + 1
	for i := 0; i < listNum-1; i++ {
		list := ui.NewList()
		list.Items = itemList[i*50 : i*50+50]
		list.ItemFgColor = ui.ColorYellow
		list.BorderRight = false
		list.Height = 200
		list.Width = 20
		list.X = i * 20
		list.Y = 3
		ui.Render(list)
	}

	ui.Handle("/sys/kbd/q", func(e ui.Event) {
		ui.StopLoop()
	})
	ui.Handle("/sys/kbd", func(e ui.Event) {
		ui.Close()
	})
	ui.Loop()
}
