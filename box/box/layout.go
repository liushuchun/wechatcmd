package box

import (
	ui "github.com/gizak/termui"
	"github.com/sebgl/chatbox/chat"
)

type Layout struct {
	msgInBox        *ui.Par
	inputBox        *ui.Par
	currentMsgCount int
	maxMsgCount     int
	msgIn           chan chat.Message
	textOut         chan string
}

func NewLayout(msgIn chan chat.Message, textOut chan string) *Layout {
	msgInBox := ui.NewPar("")
	msgInBox.Height = 20
	msgInBox.TextFgColor = ui.ColorWhite
	msgInBox.BorderLabel = "聊天室"
	msgInBox.BorderFg = ui.ColorCyan

	inputBox := ui.NewPar("")
	inputBox.Height = 3
	inputBox.TextFgColor = ui.ColorWhite
	inputBox.BorderLabel = "输入框"
	inputBox.BorderFg = ui.ColorCyan

	return &Layout{
		msgInBox:        msgInBox,
		inputBox:        inputBox,
		msgIn:           msgIn,
		textOut:         textOut,
		currentMsgCount: 0,
		maxMsgCount:     18,
	}
}

func (l *Layout) Init() {
	err := ui.Init()
	if err != nil {
		panic(err)
	}

	defer ui.Close()

	ui.Body.AddRows(
		ui.NewRow(
			ui.NewCol(12, 0, l.msgInBox)),
		ui.NewRow(
			ui.NewCol(12, 0, l.inputBox)),
	)

	ui.Body.Align()

	ui.Handle("/sys/kbd/C-c", func(ui.Event) {
		ui.StopLoop()
	})
	ui.Handle("/sys/kbd/C-d", func(ui.Event) {
		ui.StopLoop()
	})
	ui.Handle("/sys/kbd/<enter>", func(ui.Event) {
		appendToPar(l.msgInBox, l.inputBox.Text)
		resetPar(l.inputBox)

	})
	ui.Handle("/sys/kbd/<space>", func(ui.Event) {
		appendToPar(l.inputBox, " ")
	})
	ui.Handle("/sys/kbd", func(e ui.Event) {
		k, ok := e.Data.(ui.EvtKbd)
		if ok {
			appendToPar(l.inputBox, k.KeyStr)
		}
	})
	ui.Handle("/sys/wnd/resize", func(e ui.Event) {
		ui.Body.Width = ui.TermWidth()
		ui.Body.Align()
		ui.Render(ui.Body)
	})

	go l.displayMsgIn()

	ui.Render(l.inputBox, l.msgInBox)
	ui.Loop()
}

func (l *Layout) displayMsgIn() {
	for m := range l.msgIn {
		if l.currentMsgCount >= l.maxMsgCount {
			resetPar(l.msgInBox)
			l.currentMsgCount = 0
		}
		formattedMsg := "(" + m.Timestamp + ") " + m.User + ": " + m.Data + "\n"
		l.currentMsgCount++
		appendToPar(l.msgInBox, formattedMsg)
	}
}

func (l *Layout) sendText(text string) {

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
