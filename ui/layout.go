package ui

import (
	"log"
	"strings"

	ui "github.com/gizak/termui"
	"github.com/hawklithm/wechatcmd/wechat"
)

const (
	CurMark  = "(bg-red)"
	PageSize = 45
)

type Layout struct {
	chatBox         *ui.Par //聊天窗口
	msgInBox        *ui.Par //消息窗口
	editBox         *ui.Par // 输入框
	userNickListBox *ui.List
	userNickList    []string
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
}

func NewLayout(userNickList []string, userIDList []string, myName, myID string, msgIn chan wechat.Message, msgOut chan wechat.MessageOut, closeChan, autoReply chan int, logger *log.Logger) *Layout {
	//用户列表框
	userMap := make(map[string]string)

	size := len(userNickList)

	for i := 0; i < size; i++ {
		userMap[userIDList[i]] = userIDList[i]
	}

	offset := 45
	if size < PageSize {
		offset = size
	}
	showUserList := userNickList[0:offset]

	showUserList[0] = AddBgColor(showUserList[0])

	userNickListBox := ui.NewList()
	userNickListBox.BorderLabel = "用户列表"
	userNickListBox.BorderFg = ui.ColorMagenta
	userNickListBox.X = 0
	userNickListBox.Y = 0

	userNickListBox.Items = showUserList
	userNickListBox.ItemFgColor = ui.ColorGreen

	chatBox := ui.NewPar("")
	chatBox.X = 20
	chatBox.Y = 0

	chatBox.TextFgColor = ui.ColorRed
	chatBox.BorderLabel = "to:" + userNickList[0]
	chatBox.BorderFg = ui.ColorMagenta

	msgInBox := ui.NewPar("")
	msgInBox.X = 60
	msgInBox.Y = 0

	msgInBox.TextFgColor = ui.ColorWhite
	msgInBox.BorderLabel = "消息窗"
	msgInBox.BorderFg = ui.ColorCyan
	msgInBox.TextFgColor = ui.ColorRGB(180, 180, 90)

	editBox := ui.NewPar("")
	editBox.X = 20
	editBox.Y = 80

	editBox.TextFgColor = ui.ColorWhite
	editBox.BorderLabel = "输入框"
	editBox.BorderFg = ui.ColorCyan
	pageCount := len(userNickList) / PageSize
	if len(userNickList)%PageSize != 0 {
		pageCount++
	}
	return &Layout{
		userNickList:    userNickList,
		showUserList:    showUserList,
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
	}
}

func (l *Layout) Init() {
	//	chinese := false
	err := ui.Init()
	if err != nil {
		panic(err)
	}
	defer ui.Close()
	ui.ThemeAttr("helloworld")

	height := ui.TermHeight()
	width := ui.TermWidth()
	l.userNickListBox.SetWidth(width * 2 / 10)
	l.userNickListBox.Height = height
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
		appendToPar(l.chatBox, l.masterName+"->"+DelBgColor(l.chatBox.BorderLabel)+":"+l.editBox.Text+"\n")
		l.logger.Println(l.editBox.Text)
		if l.editBox.Text != "" {

			l.SendText(l.editBox.Text)
		}
		resetPar(l.editBox)

	})
	ui.Handle("/sys/kbd/C-1", func(ui.Event) {
		l.autoReply <- 1 //开启自动回复
	})
	ui.Handle("/sys/kbd/C-2", func(ui.Event) {
		l.autoReply <- 0 //关闭自动回复
	})
	ui.Handle("/sys/kbd/C-3", func(ui.Event) {
		l.autoReply <- 3 //开启机器人自动回复
	})

	ui.Handle("/sys/kbd/C-n", func(ui.Event) {
		l.NextUser()
	})

	ui.Handle("/sys/kbd/C-p", func(ui.Event) {
		l.PrevUser()
	})

	ui.Handle("/sys/kbd/<space>", func(ui.Event) {
		appendToPar(l.editBox, " ")
	})
	ui.Handle("/sys/kbd/C-8", func(ui.Event) {
		if l.editBox.Text == "" {
			return
		}
		runslice := []rune(l.editBox.Text)
		if len(runslice) == 0 {
			return
		} else {
			l.editBox.Text = string(runslice[0 : len(runslice)-1])
			setPar(l.editBox)
		}
	})
	ui.Handle("/sys/kbd", func(e ui.Event) {

		if k, ok := e.Data.(ui.EvtKbd); ok {
			// chinese = false
			// for _, r := range k.KeyStr {
			// 	if unicode.Is(unicode.Scripts["Han"], r) {
			// 		chinese = true
			// 	}
			// }
			// if chinese && len(k.KeyStr) > 1 {
			// 	runslice := []rune(k.KeyStr)

			// 	temp := runslice[len(runslice)-1]
			// 	runslice = runslice[0 : len(runslice)-1]
			// 	runslice = append(runslice, temp)
			// }

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
	ui.Render(l.msgInBox, l.chatBox, l.editBox, l.userNickListBox)
	ui.Loop()
}

func (l *Layout) displayMsgIn() {
	var (
		msg wechat.Message
	)

	for {
		select {

		case msg = <-l.msgIn:

			text := msg.String()

			appendToPar(l.msgInBox, text)

			if msg.FromUserName == l.userIDList[l.curPage*PageSize+l.userCur] {

				appendToPar(l.chatBox, text)
			}

		case <-l.closeChan:
			break
		}

	}
	return
}

func (l *Layout) PrevUser() {
	if l.userCur-1 < 0 { //如果是第一行
		if l.curPage > 0 { //如果不是第一页
			l.userCur = PageSize - 1
			l.curPage-- //到上一页
			//刷新一下显示的内容
			l.showUserList = l.userNickList[l.curPage*l.pageSize : l.curPage*l.pageSize+l.pageSize]
		} else {
			//如果是第一页
			//跳转到最后一页

			l.userCur = (l.userCount % l.pageSize) - 1
			if l.userCur < 0 {
				l.userCur = l.pageSize - 1
			}
			l.curPage = l.pageCount - 1
			l.showUserList = l.userNickList[l.curPage*l.pageSize : l.userCount]

		}
		l.showUserList[l.userCur] = AddBgColor(l.showUserList[l.userCur])
		l.userNickListBox.Items = l.showUserList

	} else { //不是第一行，则删掉前面一行的信息，更新上一个的信息。
		l.userNickListBox.Items[l.userCur] = DelBgColor(l.userNickListBox.Items[l.userCur])
		l.userCur--
		l.userNickListBox.Items[l.userCur] = AddBgColor(l.userNickListBox.Items[l.userCur])

	}
	l.chatBox.BorderLabel = DelBgColor(l.showUserList[l.userCur])
	ui.Render(l.userNickListBox, l.chatBox)

}

func (l *Layout) NextUser() {
	if l.userCur+1 >= l.pageSize || l.userCur+1 >= len(l.showUserList) { //跳出了对应的下标
		l.userNickListBox.Items[l.userCur] = DelBgColor(l.userNickListBox.Items[l.userCur])

		l.userCur = 0
		l.userNickListBox.Items[l.userCur] = AddBgColor(l.userNickListBox.Items[l.userCur])

		if l.curPage+1 >= l.pageCount { //当前页是最后一页了
			l.curPage = 0
		} else {
			l.curPage++
		}

		if l.curPage == l.pageCount-1 { //最后一页，判断情况
			l.showUserList = l.userNickList[l.curPage*l.pageSize : l.userCount]
		} else {
			l.showUserList = l.userNickList[l.curPage*l.pageSize : l.curPage*l.pageSize+l.pageSize]
		}
		//设定第一行是背景色
		l.showUserList[0] = AddBgColor(l.showUserList[0])
		l.userNickListBox.Items = l.showUserList
	} else {
		l.userNickListBox.Items[l.userCur] = DelBgColor(l.userNickListBox.Items[l.userCur])
		l.userCur++
		l.userNickListBox.Items[l.userCur] = AddBgColor(l.userNickListBox.Items[l.userCur])
	}
	l.chatBox.BorderLabel = DelBgColor(l.userNickListBox.Items[l.userCur])

	ui.Render(l.userNickListBox, l.chatBox)

}

func (l *Layout) SendText(text string) {
	msg := wechat.MessageOut{}
	msg.Content = text
	msg.ToUserName = l.userIDList[l.curPage*PageSize+l.userCur]
	//appendToPar(l.msgInBox, fmt.Sprintf(text))

	l.msgOut <- msg
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

func appendToPar(p *ui.Par, k string) {
	if strings.Count(p.Text, "\n") >= 20 {
		p.Text = ""
	}
	p.Text += k
	ui.Render(p)
}

func resetPar(p *ui.Par) {
	p.Text = ""
	ui.Render(p)
}

func setPar(p *ui.Par) {
	ui.Render(p)
}
