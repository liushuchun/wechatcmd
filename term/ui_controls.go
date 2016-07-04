package term

import (
	"os"

	ui "github.com/gizak/termui"
)

func registerEventHandlers() {
	ui.Handle("/sys/kbd/", func(ev ui.Event) {
		handleAnyKey(ev)
	})
	ui.Handle("/sys/kbd/C-c", func(ui.Event) {
		handleQuit()
	})
	ui.Handle("/sys/wnd/resize", func(ui.Event) {
		handleResize()
	})
}

func deregisterEventHandlers() {
	ui.Handle("/sys/kbd/", func(ev ui.Event) {})
	ui.Handle("/sys/kbd/C-c", func(ev ui.Event) {})
	ui.Handle("/sys/wnd/resize", func(ev ui.Event) {})
}

func handleLabelViewKey() {
	switch page := currentPage.(type) {
	case *TicketListPage:
		q := new(LabelListPage)
		q.ActiveQuery = page.ActiveQuery
		previousPages = append(previousPages, currentPage)
		currentPage = q
		changePage()
	}
}
func handleQuit() {
	ui.Close()
	os.Exit(0)
}

func handleSortOrderKey() {
	switch currentPage.(type) {
	case *TicketListPage:
		q := new(SortOrderPage)
		previousPages = append(previousPages, currentPage)
		currentPage = q
		changePage()
	}
}

func handleRefreshKey() {
	if obj, ok := currentPage.(Refresher); ok {
		obj.Refresh()
	}
}

func handleEditKey() {
	if obj, ok := currentPage.(TicketEditer); ok {
		obj.EditTicket()
	}
}

func handleBackKey() {
	if obj, ok := currentPage.(GoBacker); ok {
		obj.GoBack()
	} else {
		ui.StopLoop()
		exitNow = true
	}
}

func handleResize() {
	changePage()
}

func handleSelectKey() {
	if obj, ok := currentPage.(ItemSelecter); ok {
		obj.SelectItem()
	}
}

func handleTopOfPageKey() {
	if obj, ok := currentPage.(PagePager); ok {
		obj.TopOfPage()
		obj.Update()
	}
}

func handleBottomOfPageKey() {
	if obj, ok := currentPage.(PagePager); ok {
		obj.BottomOfPage()
		obj.Update()
	}
}

func handleUpKey() {
	if obj, ok := currentPage.(PagePager); ok {
		obj.PreviousLine(1)
		obj.Update()
	}
}

func handleDownKey() {
	if obj, ok := currentPage.(PagePager); ok {
		obj.NextLine(1)
		obj.Update()
	}
}

func handlePageUpKey() {
	if obj, ok := currentPage.(PagePager); ok {
		obj.PreviousPage()
		obj.Update()
	}
}

func handlePageDownKey() {
	if obj, ok := currentPage.(PagePager); ok {
		obj.NextPage()
		obj.Update()
	}
}

func handleParaUpKey() {
	if obj, ok := currentPage.(PagePager); ok {
		obj.PreviousPara()
		obj.Update()
	}
}

func handleParaDownKey() {
	if obj, ok := currentPage.(PagePager); ok {
		obj.NextPara()
		obj.Update()
	}
}

func handleNextSearchKey() {
	if obj, ok := currentPage.(Searcher); ok {
		obj.Search()
	}
}

func handleNextTicketKey() {
	if obj, ok := currentPage.(NextTicketer); ok {
		obj.NextTicket()
	}
}

func handlePrevTicketKey() {
	if obj, ok := currentPage.(PrevTicketer); ok {
		obj.PrevTicket()
	}
}

func handleHelp() {
	previousPages = append(previousPages, currentPage)
	currentPage = helpPage
	changePage()
}

func handleNavigateKey(e ui.Event) {
	key := e.Data.(ui.EvtKbd).KeyStr
	switch key {
	case "L":
		handleLabelViewKey()
	case "S":
		handleSortOrderKey()
	case "C-r":
		handleRefreshKey()
	case "E":
		handleEditKey()
	case "w":
		args := []string{"add"}
		handleWatchCommand(args)
	case "W":
		args := []string{"remove"}
		handleWatchCommand(args)
	case "v":
		handleVoteCommand(true)
	case "V":
		handleVoteCommand(false)
	case "q":
		handleBackKey()
	case "<enter>":
		handleSelectKey()
	case "g":
		handleTopOfPageKey()
	case "G":
		handleBottomOfPageKey()
	case "<space>":
		handlePageDownKey()
	case "C-f":
		handlePageDownKey()
	case "C-b":
		handlePageUpKey()
	case "}":
		handleParaDownKey()
	case "{":
		handleParaUpKey()
	case "<down>":
		handleDownKey()
	case "<up>":
		handleUpKey()
	case "j":
		handleDownKey()
	case "k":
		handleUpKey()
	case ":":
		handleCommandKey(e)
	case "/":
		handleCommandKey(e)
	case "?":
		handleCommandKey(e)
	case "n":
		handleNextSearchKey()
	case "N":
		handleNextTicketKey()
	case "P":
		handlePrevTicketKey()
	case "h":
		handleHelp()
	}
}

func handleCommandKey(e ui.Event) {
	if obj, ok := currentPage.(PagePager); ok {
		if obj, ok := obj.(CommandBoxer); ok {
			obj.SetCommandMode(true)
			obj.CommandBar().Reset()
			handleAnyKey(e)
		}
	}
}

func handleEditBoxKey(obj EditPager, key string) {
	var str string
	switch {
	case len(key) == 1:
		str = key
	case key == "<space>":
		str = ` `
	case key == "<enter>":
		str = "\n"
	case key == "<backspace>" || key == "C-8":
		// C-8 == ^? == backspace on a UK macbook
		obj.DeleteRuneBackward()
		obj.Update()
		return
	default:
		return
	}
	r := decodeTermuiKbdStringToRune(str)
	obj.InsertRune(r)
	obj.Update()
	return
}

func handleAnyKey(e ui.Event) {
	key := e.Data.(ui.EvtKbd).KeyStr
	if obj, ok := currentPage.(PagePager); ok && obj.IsPopulated() {
		if obj, ok := obj.(CommandBoxer); ok {
			if !obj.CommandMode() {
				handleNavigateKey(e)
				return
			}
		} else {
			handleNavigateKey(e)
			return
		}
	}

	if obj, ok := currentPage.(EditPager); ok {
		handleEditBoxKey(obj, key)
		return
	}

	if obj, ok := currentPage.(CommandBoxer); ok && obj.CommandMode() {
		cb := obj.CommandBar()
		if key == "<enter>" {
			cb.Submit()
			return
		} else if key == "<up>" {
			cb.PreviousCommand()
			return
		} else if key == "<down>" {
			cb.NextCommand()
			return
		} else {
			handleEditBoxKey(cb, key)
			return
		}
	}

}
