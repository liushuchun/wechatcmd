package main

import (
	"strings"

	ui "github.com/gizak/termui"
	"github.com/liushuchun/wechatcmd/wechat"
)

const (
	CurMark  = "*(bg-green)"
	PageSize = 50
)

type Layout struct {
	chatBox     *ui.Par //聊天窗口
	msgInBox    *ui.Par //消息窗口
	editBox     *ui.Par // 输入框
	userListBox *ui.List
	userList    []string

	currentMsgCount int
	maxMsgCount     int
	userIn          chan []string       // 用户的刷新
	chatIn          chan wechat.Message // 聊天界面的消息刷新
	msgIn           chan wechat.Message // 消息刷新
	textOut         chan string         //  消息输出
	showUserList    []string
	userCount       int //用户总数，这里有重复,后面会修改
	pageCount       int // page总数。
	userCur         int // 当前page中所选中的用户
	curPage         int // 当前所在页
	pageSize        int // page的size默认是50
}

func NewLayout(initUserList, userList []string, chatIn, msgIn chan wechat.Message, textOut chan string) *Layout {
	//用户列表框
	userList = append(initUserList, userList...)
	size := len(userList)
	offset := 50
	if size < PageSize {
		offset = size
	}
	showUserList := userList[0:offset]

	showUserList[0] = showUserList[0] + CurMark

	userListBox := ui.NewList()
	userListBox.BorderLabel = "用户列表"
	userListBox.BorderFg = ui.ColorMagenta
	userListBox.X = 0
	userListBox.Y = 0
	userListBox.Height = 100
	userListBox.Width = 20
	userListBox.Items = showUserList
	userListBox.ItemFgColor = ui.ColorGreen

	chatBox := ui.NewPar("")
	chatBox.X = 20
	chatBox.Y = 0
	chatBox.Height = 80
	chatBox.Width = 40
	chatBox.TextFgColor = ui.ColorRed
	chatBox.BorderLabel = "to:" + userList[0]
	chatBox.BorderFg = ui.ColorMagenta

	msgInBox := ui.NewPar("")
	msgInBox.X = 60
	msgInBox.Y = 0
	msgInBox.Height = 80
	msgInBox.Width = 40
	msgInBox.TextFgColor = ui.ColorWhite
	msgInBox.BorderLabel = "消息窗"
	msgInBox.BorderFg = ui.ColorCyan
	msgInBox.TextFgColor = ui.ColorRGB(180, 180, 90)

	editBox := ui.NewPar("")
	editBox.X = 20
	editBox.Y = 80
	editBox.Height = 20
	editBox.Width = 80
	editBox.TextFgColor = ui.ColorWhite
	editBox.BorderLabel = "输入框"
	editBox.BorderFg = ui.ColorCyan
	pageCount := len(userList) / PageSize
	if len(userList)%PageSize != 0 {
		pageCount++
	}
	return &Layout{
		userList:        userList,
		showUserList:    showUserList,
		userCur:         0,
		curPage:         0,
		chatIn:          chatIn,
		msgInBox:        msgInBox,
		userListBox:     userListBox,
		chatBox:         chatBox,
		editBox:         editBox,
		msgIn:           msgIn,
		textOut:         textOut,
		currentMsgCount: 0,
		maxMsgCount:     18,
		userCount:       len(userList),
		pageCount:       pageCount,
		pageSize:        PageSize,
	}
}

func (l *Layout) Init() {
	err := ui.Init()
	if err != nil {
		panic(err)
	}
	defer ui.Close()

	height := ui.TermHeight()
	width := ui.TermWidth()
	l.userListBox.SetWidth(width * 2 / 10)

	l.msgInBox.SetWidth(width * 4 / 10)
	l.msgInBox.SetX(width * 6 / 10)
	l.msgInBox.Height = height * 8 / 10

	l.chatBox.SetX(width * 2 / 10)
	l.chatBox.Height = height * 8 / 10
	l.chatBox.SetWidth(width * 4 / 10)

	l.editBox.SetX(width * 2 / 10)
	l.editBox.SetY(height * 8 / 10)
	l.editBox.SetWidth(width * 8 / 10)
	l.editBox.Height = height * 2 / 10

	ui.Handle("/sys/kbd/C-c", func(ui.Event) {
		ui.StopLoop()
	})
	ui.Handle("/sys/kbd/C-d", func(ui.Event) {
		ui.StopLoop()
	})
	ui.Handle("/sys/kbd/<enter>", func(ui.Event) {
		appendToPar(l.msgInBox, l.editBox.Text)
		resetPar(l.editBox)

	})
	ui.Handle("/sys/kbd/<space>", func(ui.Event) {
		appendToPar(l.editBox, " ")
	})
	ui.Handle("/sys/kbd", func(e ui.Event) {
		k, ok := e.Data.(ui.EvtKbd)
		if ok {
			appendToPar(l.editBox, k.KeyStr)
		}
	})
	ui.Handle("/sys/wnd/resize", func(e ui.Event) {
		ui.Body.Width = ui.TermWidth()
		ui.Body.Align()
		ui.Render(ui.Body)
	})

	go l.displayMsgIn()
	// 注册各个组件
	ui.Render(l.msgInBox, l.chatBox, l.editBox, l.userListBox)
	ui.Loop()
}

func (l *Layout) displayMsgIn() {
	for m := range l.msgIn {
		if l.currentMsgCount >= l.maxMsgCount {
			resetPar(l.msgInBox)
			l.currentMsgCount = 0
		}
		formattedMsg := "(" + m.Timestamp + ") " + m.FromUser + ": " + m.Content + "\n"
		l.currentMsgCount++
		appendToPar(l.msgInBox, formattedMsg)
	}
}

func (l *Layout) NextUser() {
	if l.userCur+1 >= l.pageSize { //跳出了对应的下标
		l.userCur = 0
		if l.curPage+1 >= l.pageCount { //当前页是最后一页了
			l.curPage = 0
		} else {
			l.curPage++
		}

		if l.curPage == l.pageCount-1 { //最后一页，判断情况
			l.showUserList = l.userList[l.curPage*l.pageSize : l.userCount]
		} else {
			l.showUserList = l.userList[l.curPage*l.pageSize : l.curPage*l.pageSize+50]
		}
		//设定第一行是背景色
		l.showUserList[0] = l.showUserList[0] + CurMark
		l.userListBox.Items = l.showUserList
	} else {
		l.userListBox.Items[l.userCur] = strings.TrimRight(l.userListBox.Items[l.userCur], CurMark)
		l.userCur++
		l.userListBox.Items[l.userCur] += CurMark
	}

	ui.Render(l.userListBox)
}

func (l *Layout) SendText(text string) {

	appendToPar(l.msgInBox, text)

	l.textOut <- text
}

func appendToPar(p *ui.Par, k string) {
	p.Text += k
	ui.Render(p)
}

func resetPar(p *ui.Par) {
	p.Text = ""
	ui.Render(p)
}
