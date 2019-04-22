package ui

import (
	"fmt"
	"github.com/skratchdot/open-golang/open"
	"image"
	"image/png"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	ui "github.com/hawklithm/termui"
	"github.com/hawklithm/termui/widgets"
	"github.com/hawklithm/wechatcmd/wechat"
)

const (
	SelectedMark   = "(bg:red)"
	UnSelectedMark = "(bg:blue)"
	PageSize       = 45
)

type Layout struct {
	chatBox         *widgets.ImageList //聊天窗口
	msgInBox        *widgets.Paragraph //消息窗口
	editBox         *widgets.Paragraph // 输入框
	userNickListBox *widgets.List
	userIDList      []string
	curUserIndex    int
	masterName      string // 主人的名字
	masterID        string //主人的id
	currentMsgCount int
	maxMsgCount     int
	userIn          chan []string            // 用户的刷新
	msgIn           chan wechat.Message      // 消息刷新
	msgOut          chan wechat.MessageOut   //  消息输出
	imageIn         chan wechat.MessageImage //  图片消息
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
	imageMap        map[string]image.Image
	msgIdList       map[string][]string
	selectedMsgId   string
}

func NewLayout(userNickList []string, userIDList []string,
	groupMemberList []wechat.Member, myName, myID string,
	msgIn chan wechat.Message, msgOut chan wechat.MessageOut,
	imageIn chan wechat.MessageImage,
	closeChan, autoReply chan int, logger *log.Logger) {

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
	imageMap := make(map[string]image.Image)

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

	chatBox := widgets.NewImageList()
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
		imageIn:         imageIn,
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
		imageMap:        imageMap,
		msgIdList:       make(map[string][]string),
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
			appendToList(l.chatBox, l.masterName+"->"+l.chatBox.
				Title+":"+l.editBox.Text+"\n")
			l.logger.Println(l.editBox.Text)
			if l.editBox.Text != "" {
				l.SendText(l.editBox.Text)
			}
			resetPar(l.editBox)
		case "<C-w>":
			l.showDetail()
		case "<C-2>":
			l.autoReply <- 0 //关闭自动回复
		case "<C-3>":
			l.autoReply <- 3 //开启机器人自动回复
		case "<C-n>":
			l.NextUser()
		case "<C-p>":
			l.PrevUser()
		case "<C-j>":
			l.NextSelect()
		case "<C-k>":
			l.PrevSelect()
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
			} else if e.Type == ui.MouseEvent {
				logger.Println("mouse event received, payload=", e.Payload,
					"id=", e.ID)
			}
		}

	}

}

func (l *Layout) displayMsgIn() {
	var (
		msg    wechat.Message
		imgMsg wechat.MessageImage
	)

	for {
		select {

		case imgMsg = <-l.imageIn:
			//l.imageMap[imgMsg.MsgId] = imgMsg.Img
			//if l.msgIdList[imgMsg.TargetId] == nil {
			//	l.msgIdList[imgMsg.TargetId] = []string{}
			//}
			//l.msgIdList[imgMsg.TargetId] = append(l.msgIdList[imgMsg.TargetId], imgMsg.MsgId)

			var newMsgText string

			if l.masterID == imgMsg.FromUserName {
				newMsgText = l.apendChatLogOut(wechat.MessageOut{ToUserName: imgMsg.
					ToUserName, Content: imgMsg.Content, Type: imgMsg.MsgType,
					MsgId: imgMsg.MsgId})
			} else {
				newMsgText = l.apendImageChatLogIn(imgMsg)
			}

			l.logger.Println("message receive = ", newMsgText)

			appendToPar(l.msgInBox, newMsgText+"\n")

			var targetUserName string
			if l.masterID == imgMsg.FromUserName {
				targetUserName = imgMsg.ToUserName
			} else {
				targetUserName = imgMsg.FromUserName
			}
			if targetUserName == l.userIDList[l.userCur] {
				l.logger.Println("append to current chatbox", imgMsg.FromUserName,
					"to=",
					imgMsg.ToUserName, "content=", imgMsg.Content)
				appendImageToList(l.chatBox, imgMsg.Img)
			}

		case msg = <-l.msgIn:

			var newMsgText string
			msg.Content = TranslateEmoji(ConvertToEmoji(msg.Content))

			if l.masterID == msg.FromUserName {
				newMsgText = l.apendChatLogOut(wechat.MessageOut{ToUserName: msg.
					ToUserName, Content: msg.Content, Type: msg.MsgType,
					MsgId: msg.MsgId})
			} else {
				newMsgText = l.apendChatLogIn(msg)
			}

			l.logger.Println("message receive = ", newMsgText)

			appendToPar(l.msgInBox, newMsgText+"\n")

			var targetUserName string
			if l.masterID == msg.FromUserName {
				targetUserName = msg.ToUserName
			} else {
				targetUserName = msg.FromUserName
			}
			if targetUserName == l.userIDList[l.userCur] {
				l.logger.Println("append to current chatbox", msg.FromUserName,
					"to=",
					msg.ToUserName, "content=", msg.Content)
				appendToList(l.chatBox, newMsgText)
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
	l.chatBox.Title = l.userNickListBox.Rows[l.userNickListBox.SelectedRow]
	l.logger.Println("title=", l.chatBox.Title, "content=",
		l.userChatLog[l.userIDList[l.userNickListBox.SelectedRow]])
	setRows(l.chatBox, l.userChatLog[l.userIDList[l.userNickListBox.SelectedRow]])
	l.logger.Println("prev user", "rows=", len(l.chatBox.Rows))
	ui.Render(l.userNickListBox, l.chatBox)
}

func setRows(p *widgets.ImageList, records []*wechat.MessageRecord) {
	var rows []*widgets.ImageListItem
	for _, i := range records {
		item := widgets.NewImageListItem()
		if i.ContentImg != nil {
			item.Img = i.ContentImg
		} else {
			item.Text = i.From + "->" + i.To + ": " + i.Content
		}
		rows = append(rows, item)
	}
	p.Rows = rows
	p.SelectedRow = len(p.Rows) - 1
	if p.SelectedRow < 0 {
		p.SelectedRow = 0
	}
}

func (l *Layout) PrevSelect() {
	l.chatBox.ScrollUp()
	ui.Render(l.chatBox)
}

func (l *Layout) showDetail() {
	item := l.chatBox.Rows[l.chatBox.SelectedRow]
	if item.Img == nil {
		return
	}
	//root, e := os.Getwd()
	//if e != nil {
	root := "/tmp"
	//}
	key := time.Now().UTC().UnixNano()
	builder := strings.Builder{}
	builder.WriteString(root)
	builder.WriteRune(os.PathSeparator)
	builder.WriteString(strconv.FormatInt(key, 10))
	builder.WriteString(".png")
	out, err := os.Create(builder.String())
	if err != nil {
		l.logger.Fatalln("open file failed! path=", builder.String(), err)
	}
	if err := png.Encode(out, item.Img); err != nil {
		l.logger.Fatalln("encode image failed! path=", builder.String(), err)
	} else {
		_ = open.Start(builder.String())
	}

}

func (l *Layout) NextSelect() {
	l.chatBox.ScrollDown()
	ui.Render(l.chatBox)
}

func (l *Layout) NextUser() {
	l.userNickListBox.ScrollDown()
	l.userCur = l.userNickListBox.SelectedRow
	l.chatBox.Title = l.userNickListBox.Rows[l.userNickListBox.SelectedRow]
	l.logger.Println("title=", l.chatBox.Title, "content=",
		l.userChatLog[l.userIDList[l.userNickListBox.SelectedRow]])
	setRows(l.chatBox, l.userChatLog[l.userIDList[l.userNickListBox.SelectedRow]])
	l.logger.Println("next user", "rows=", len(l.chatBox.Rows), "top row=",
		l.chatBox.TopLine())
	ui.Render(l.userNickListBox, l.chatBox)
}

func (l *Layout) SendText(text string) {
	msg := wechat.MessageOut{}
	msg.Content = text
	msg.ToUserName = l.userIDList[l.userCur]
	//appendToPar(l.msgInBox, fmt.Sprintf(text))

	l.apendChatLogOut(msg)

	l.msgOut <- msg
}

func (l *Layout) apendChatLogOut(msg wechat.MessageOut) string {
	if l.userChatLog[msg.ToUserName] == nil {
		l.userChatLog[msg.ToUserName] = []*wechat.MessageRecord{}
	}

	newMsg := wechat.NewMessageRecordOut(l.masterID,
		msg)

	if l.groupMemberMap[newMsg.From] != nil {
		if newMsg.Type == 3 {
			newMsg.Content = l.getUserIdAndConvertImgContent(newMsg.Content,
				l.groupMemberMap[newMsg.From])

		} else {
			newMsg.Content = l.getUserIdFromContent(newMsg.Content,
				l.groupMemberMap[newMsg.From])
		}
	} else {
		if newMsg.Type == 3 {
			newMsg.Content = "图片"
		}
	}

	if l.userMap[newMsg.To] != "" {
		newMsg.To = l.userMap[newMsg.To]
	}

	if l.userMap[newMsg.From] != "" {
		newMsg.From = l.userMap[newMsg.From]
	}

	l.userChatLog[msg.ToUserName] = append(l.userChatLog[msg.ToUserName],
		newMsg)

	return newMsg.String()
}

func (l *Layout) getUserIdFromContent(content string,
	userMap map[string]string) string {
	if userMap == nil {
		return content
	}
	s := strings.Split(content, ":")
	if len(s) > 0 && userMap[s[0]] != "" {
		s[0] = userMap[s[0]]
	}
	l.logger.Println("groupMap=", userMap, "s=", s)
	builder := strings.Builder{}
	for i, sub := range s {
		builder.WriteString(sub)
		if i != len(s)-1 {
			builder.WriteString(":")
		}
	}
	return builder.String()
}

func (l *Layout) getUserIdAndConvertImgContent(content string,
	userMap map[string]string) string {
	if userMap == nil {
		return content
	}
	s := strings.Split(content, ":")
	if len(s) > 0 && userMap[s[0]] != "" {
		s[0] = userMap[s[0]]
	}
	l.logger.Println("groupMap=", userMap, "s=", s)
	return s[0] + ":" + AddUnSelectedBg("图片")
}

func (l *Layout) apendChatLogIn(msg wechat.Message) string {
	if l.userChatLog[msg.FromUserName] == nil {
		l.userChatLog[msg.FromUserName] = []*wechat.MessageRecord{}
	}

	newMsg := wechat.NewMessageRecordIn(msg)

	if l.groupMemberMap[newMsg.From] != nil {
		if newMsg.Type == 3 {
			newMsg.Content = l.getUserIdAndConvertImgContent(newMsg.Content,
				l.groupMemberMap[newMsg.From])
		} else {
			newMsg.Content = l.getUserIdFromContent(newMsg.Content,
				l.groupMemberMap[newMsg.From])
		}
	} else {
		if newMsg.Type == 3 {
			newMsg.Content = "图片"
		}
	}

	if l.userMap[newMsg.To] != "" {
		newMsg.To = l.userMap[newMsg.To]
	}

	if l.userMap[newMsg.From] != "" {
		newMsg.From = l.userMap[newMsg.From]
	}

	l.userChatLog[msg.FromUserName] = append(l.userChatLog[msg.
		FromUserName], newMsg)

	return newMsg.String()

}

func (l *Layout) apendImageChatLogIn(msg wechat.MessageImage) string {
	if l.userChatLog[msg.FromUserName] == nil {
		l.userChatLog[msg.FromUserName] = []*wechat.MessageRecord{}
	}

	newMsg := wechat.NewImageMessageRecordIn(msg)

	if l.groupMemberMap[newMsg.From] != nil {
		if newMsg.Type == 3 {
			newMsg.Content = l.getUserIdAndConvertImgContent(newMsg.Content,
				l.groupMemberMap[newMsg.From])
		} else {
			newMsg.Content = l.getUserIdFromContent(newMsg.Content,
				l.groupMemberMap[newMsg.From])
		}
	} else {
		if newMsg.Type == 3 {
			newMsg.Content = "图片"
		}
	}

	if l.userMap[newMsg.To] != "" {
		newMsg.To = l.userMap[newMsg.To]
	}

	if l.userMap[newMsg.From] != "" {
		newMsg.From = l.userMap[newMsg.From]
	}

	l.userChatLog[msg.FromUserName] = append(l.userChatLog[msg.
		FromUserName], newMsg)

	return newMsg.String()

}

func AddSelectedBg(msg string) string {
	return AddBgColor(DelBgColor(msg), SelectedMark)
}

func AddUnSelectedBg(msg string) string {
	return AddBgColor(DelBgColor(msg), UnSelectedMark)
}
func AddBgColor(msg string, color string) string {
	if strings.HasPrefix(msg, "[") {
		return msg
	}
	return "[" + msg + "]" + color
}
func DelBgColor(msg string) string {

	if !strings.HasPrefix(msg, "[") {
		return msg
	}
	return msg[1 : len(msg)-9]
}

func appendToPar(p *widgets.Paragraph, k string) {
	if strings.Count(p.Text, "\n") >= 20 {
		subText := strings.Split(p.Text, "\n")
		p.Text = strings.Join(subText[len(subText)-20:], "\n")
	}
	p.Text += k
	ui.Render(p)
}

func appendToList(p *widgets.ImageList, k string) {
	item := widgets.NewImageListItem()
	item.Text = k
	p.Rows = append(p.Rows, item)
	ui.Render(p)
}

func appendImageToList(p *widgets.ImageList, k image.Image) {
	item := widgets.NewImageListItem()
	item.Img = k
	p.Rows = append(p.Rows, item)
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
	var start = 0
	if len(records) > 20 {
		start = len(records) - 20
	}
	for _, i := range records[start:] {
		_, _ = fmt.Fprint(&b, i.From+"->"+i.To+": "+i.Content+"\n")
	}
	return b.String()
}
