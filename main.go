package main

import (
	"flag"
	"log"
	"os"

	"github.com/liushuchun/wechatcmd/wechat"
)

var (
	DebugMode = flag.Bool("debug", false, "是否为 debug 模式")
)

func main() {
	flag.Parse()
	logger := log.New(os.Stdout, "[robot]", log.LstdFlags)

	logger.Println("天启元年")
	wechat := wechat.NewWechat(logger)

	if err := wechat.WaitForLogin(); err != nil {
		wechat.Log.Fatalf("等待失败：%s\n", err.Error())
		return
	}

	wechat.Log.Printf("登陆...")
	if err := wechat.Login(); err != nil {
		wechat.Log.Println("登陆失败：%v\n", err)
		return
	}

	wechat.Log.Println("成功")

}
