package ui

import (
	"fmt"
	"log"
	"strings"

	ui "github.com/hawklithm/termui"
	"github.com/hawklithm/termui/widgets"
	"github.com/hawklithm/wechatcmd/wechat"
)

const (
	CurMark  = "(bg-red)"
	PageSize = 45
)

type Layout struct {
	chatBox         *widgets.Paragraph //聊天窗口
	msgInBox        *widgets.Paragraph //消息窗口
	editBox         *widgets.Paragraph // 输入框
	userNickListBox *widgets.List
	userIDList      []string
	curUserIndex    int
	masterName      string // 主人的名字
	masterID        string //主人的id
	currentMsgCount int
	maxMsgCount     int
	userIn          chan []string          // 用户的刷新
	msgIn           chan wechat.Message    // 消息刷新
	msgOut          chan wechat.MessageOut //  消息输出
	closeChan       chan int
	autoReply       chan int
	showUserList    []string
	userCount       int //用户总数，这里有重复,后面会修改
	pageCount       int // page总数。
	userCur         int // 当前page中所选中的用户
	curPage         int // 当前所在页
	pageSize        int // page的size默认是50
	curUserId       string
	userMap         map[string]string
	logger          *log.Logger
	userChatLog     map[string][]*wechat.MessageRecord
	groupMemberMap  map[string]map[string]string
}

func NewLayout(userNickList []string, userIDList []string,
	groupMemberList []wechat.Member, myName, myID string,
	msgIn chan wechat.Message, msgOut chan wechat.MessageOut, closeChan, autoReply chan int, logger *log.Logger) {

	//	chinese := false
	err := ui.Init()
	if err != nil {
		panic(err)
	}
	defer ui.Close()

	//用户列表框
	userMap := make(map[string]string)
	userChatLog := make(map[string][]*wechat.MessageRecord)
	groupMemberMap := make(map[string]map[string]string)

	size := len(userNickList)

	for i := 0; i < size; i++ {
		userMap[userIDList[i]] = userNickList[i]
	}
	userMap[myID] = myName

	for _, m := range groupMemberList {
		if groupMemberMap[m.UserName] == nil {
			groupMemberMap[m.UserName] = make(map[string]string)
		}
		for _, user := range m.MemberList {
			groupMemberMap[m.UserName][user.UserName] = user.NickName
		}
	}

	userNickListBox := widgets.NewList()
	userNickListBox.Title = "用户列表"
	//userNickListBox.BorderStyle = ui.NewStyle(ui.ColorMagenta)
	//userNickListBox.Border = true
	userNickListBox.TextStyle = ui.NewStyle(ui.ColorYellow)
	userNickListBox.WrapText = false
	userNickListBox.SelectedRowStyle = ui.NewStyle(ui.ColorWhite, ui.ColorRed)

	width, height := ui.TerminalDimensions()

	userNickListBox.SetRect(0, 0, width*2/10, height)

	userNickListBox.Rows = userNickList

	chatBox := widgets.NewParagraph()
	chatBox.SetRect(width*2/10, 0, width*6/10, height*8/10)

	chatBox.TextStyle = ui.NewStyle(ui.ColorRed)
	chatBox.Title = "to:" + userNickList[0]
	chatBox.BorderStyle = ui.NewStyle(ui.ColorMagenta)

	msgInBox := widgets.NewParagraph()

	msgInBox.SetRect(width*6/10, 0, width, height*8/10)

	msgInBox.TextStyle = ui.NewStyle(ui.ColorWhite)
	msgInBox.Title = "消息窗"
	msgInBox.BorderStyle = ui.NewStyle(ui.ColorCyan)

	editBox := widgets.NewParagraph()
	editBox.SetRect(width*2/10, height*8/10, width, height)

	editBox.TextStyle = ui.NewStyle(ui.ColorWhite)
	editBox.Title = "输入框"
	editBox.BorderStyle = ui.NewStyle(ui.ColorCyan)

	pageCount := len(userNickList) / PageSize
	if len(userNickList)%PageSize != 0 {
		pageCount++
	}
	l := &Layout{
		showUserList:    userNickList,
		userCur:         0,
		curPage:         0,
		msgInBox:        msgInBox,
		userNickListBox: userNickListBox,
		userIDList:      userIDList,
		chatBox:         chatBox,
		editBox:         editBox,
		msgIn:           msgIn,
		msgOut:          msgOut,
		closeChan:       closeChan,
		currentMsgCount: 0,
		maxMsgCount:     18,
		userCount:       len(userNickList),
		pageCount:       pageCount,
		pageSize:        PageSize,
		curUserIndex:    0,
		userMap:         userMap,
		masterID:        myID,
		masterName:      myName,
		logger:          logger,
		userChatLog:     userChatLog,
		groupMemberMap:  groupMemberMap,
	}

	go l.displayMsgIn()

	// 注册各个组件
	ui.Render(l.msgInBox, l.chatBox, l.editBox, l.userNickListBox)

	uiEvents := ui.PollEvents()
	for {
		e := <-uiEvents
		switch e.ID {
		case "q", "<C-c>", "<C-d>":
			return
		case "<Enter>":
			appendToPar(l.chatBox, l.masterName+"->"+DelBgColor(l.chatBox.
				Title)+":"+l.editBox.Text+"\n")
			l.logger.Println(l.editBox.Text)
			if l.editBox.Text != "" {
				l.SendText(l.editBox.Text)
			}
			resetPar(l.editBox)
		case "<C-1>":
			l.autoReply <- 1 //开启自动回复
		case "<C-2>":
			l.autoReply <- 0 //关闭自动回复
		case "<C-3>":
			l.autoReply <- 3 //开启机器人自动回复
		case "<C-n>":
			l.NextUser()
		case "<C-p>":
			l.PrevUser()
		case "<Space>":
			appendToPar(l.editBox, " ")
		case "<Backspace>":
			if l.editBox.Text != "" {
				runslice := []rune(l.editBox.Text)
				if len(runslice) == 0 {
					return
				} else {
					l.editBox.Text = string(runslice[0 : len(runslice)-1])
					setPar(l.editBox)
				}
			}
		default:
			//logger.Println("default event received, payload=", e.Payload,
			//	"id=", e.ID, "type=", e.Type)
			if e.Type == ui.KeyboardEvent {
				k := e.ID
				appendToPar(l.editBox, k)
			} else if e.Type == ui.ResizeEvent {
				logger.Println("resize event received, payload=", e.Payload,
					"id=", e.ID)
			}
		}

	}

}

func (l *Layout) displayMsgIn() {
	var (
		msg wechat.Message
	)

	for {
		select {

		case msg = <-l.msgIn:

			text := msg.String()

			if l.masterID == msg.FromUserName {
				l.apendOut(wechat.MessageOut{ToUserName: msg.ToUserName, Content: msg.Content, Type: msg.MsgType})
			} else {
				l.apendIn(msg)
			}

			l.logger.Println("message receive, from=", msg.FromUserName, "to=",
				msg.ToUserName, "content=", msg.Content)

			appendToPar(l.msgInBox, text)

			if msg.FromUserName == l.userIDList[l.userCur] {

				appendToPar(l.chatBox, text)
			}

		case <-l.closeChan:
			break
		}

	}
	return
}

func (l *Layout) PrevUser() {
	l.userNickListBox.ScrollUp()
	l.userCur = l.userNickListBox.SelectedRow
	l.chatBox.Title = DelBgColor(l.userNickListBox.Rows[l.userNickListBox.SelectedRow])
	l.logger.Println("title=", l.chatBox.Title, "content=",
		l.userChatLog[l.userIDList[l.userNickListBox.SelectedRow]])
	l.chatBox.Text = convertChatLogToText(l.userChatLog[l.userIDList[l.userNickListBox.SelectedRow]])
	ui.Render(l.userNickListBox, l.chatBox)
}

func (l *Layout) NextUser() {
	l.userNickListBox.ScrollDown()
	l.userCur = l.userNickListBox.SelectedRow
	l.chatBox.Title = DelBgColor(l.userNickListBox.Rows[l.userNickListBox.SelectedRow])
	l.logger.Println("title=", l.chatBox.Title, "content=",
		l.userChatLog[l.userIDList[l.userNickListBox.SelectedRow]])
	l.chatBox.Text = convertChatLogToText(l.userChatLog[l.userIDList[l.userNickListBox.SelectedRow]])
	ui.Render(l.userNickListBox, l.chatBox)
}

func (l *Layout) SendText(text string) {
	msg := wechat.MessageOut{}
	msg.Content = text
	msg.ToUserName = l.userIDList[l.userCur]
	//appendToPar(l.msgInBox, fmt.Sprintf(text))

	l.apendOut(msg)

	l.msgOut <- msg
}

func (l *Layout) apendOut(msg wechat.MessageOut) {
	if l.userChatLog[msg.ToUserName] == nil {
		l.userChatLog[msg.ToUserName] = []*wechat.MessageRecord{}
	}

	newMsg := wechat.NewMessageRecordOut(l.masterID,
		msg)

	if l.groupMemberMap[newMsg.From] != nil {
		newMsg.Content = l.getUserIdFromContent(newMsg.Content,
			l.groupMemberMap[newMsg.From])
	}

	if l.userMap[newMsg.To] != "" {
		newMsg.To = l.userMap[newMsg.To]
	}

	if l.userMap[newMsg.From] != "" {
		newMsg.From = l.userMap[newMsg.From]
	}

	l.userChatLog[msg.ToUserName] = append(l.userChatLog[msg.ToUserName],
		newMsg)
}

func (l *Layout) getUserIdFromContent(content string,
	userMap map[string]string) string {
	s := strings.Split(content, ":")
	if len(s) > 0 && userMap[s[0]] != "" {
		s[0] = userMap[s[0]]
	}
	l.logger.Println("groupMap=", userMap, "s=", s)
	builder := strings.Builder{}
	for _, sub := range s {
		builder.WriteString(sub)
	}
	return builder.String()
}

func (l *Layout) apendIn(msg wechat.Message) {
	if l.userChatLog[msg.FromUserName] == nil {
		l.userChatLog[msg.FromUserName] = []*wechat.MessageRecord{}
	}

	newMsg := wechat.NewMessageRecordIn(msg)

	if l.groupMemberMap[newMsg.From] != nil {
		newMsg.Content = l.getUserIdFromContent(newMsg.Content,
			l.groupMemberMap[newMsg.From])
	}

	if l.userMap[newMsg.To] != "" {
		newMsg.To = l.userMap[newMsg.To]
	}

	if l.userMap[newMsg.From] != "" {
		newMsg.From = l.userMap[newMsg.From]
	}

	l.userChatLog[msg.FromUserName] = append(l.userChatLog[msg.
		FromUserName], newMsg)

}

func AddBgColor(msg string) string {
	if strings.HasPrefix(msg, "[") {
		return msg
	}
	return "[" + msg + "]" + CurMark
}
func DelBgColor(msg string) string {

	if !strings.HasPrefix(msg, "[") {
		return msg
	}
	return msg[1 : len(msg)-9]
}

func appendToPar(p *widgets.Paragraph, k string) {
	if strings.Count(p.Text, "\n") >= 20 {
		p.Text = ""
	}
	p.Text += k
	ui.Render(p)
}

func resetPar(p *widgets.Paragraph) {
	p.Text = ""
	ui.Render(p)
}

func setPar(p *widgets.Paragraph) {
	ui.Render(p)
}

func convertChatLogToText(records []*wechat.MessageRecord) string {
	var b strings.Builder
	for _, i := range records {
		_, _ = fmt.Fprint(&b, i.From+"->"+i.To+": "+i.Content+"\n")
	}
	return b.String()
}
