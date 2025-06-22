package main

import (
	"log"

	"github.com/biisal/godo/todos/action"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func main() {
	tview.Styles.PrimitiveBackgroundColor = tcell.ColorDefault
	tview.Styles.ContrastBackgroundColor = tcell.ColorGray
	tview.Styles.MoreContrastBackgroundColor = tcell.ColorBlack
	tview.Styles.BorderColor = tcell.NewHexColor(0x00f5ff)
	tview.Styles.PrimaryTextColor = tcell.NewHexColor(0x00f5ff)
	title := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetText("GoDo - Go And Do :))").
		SetDynamicColors(true)

	todoUI := &action.TodoUI{
		App:          tview.NewApplication(),
		Form:         tview.NewForm(),
		TodoList:     tview.NewList(),
		NoteList:     tview.NewList(),
		Instructions: tview.NewTextView(),
		Description:  tview.NewTextView(),
	}
	var focusModes = []tview.Primitive{todoUI.TodoList, todoUI.NoteList, todoUI.Form}

	todoUI.SetUpList()
	todoUI.SetUpNoteList()
	todoUI.SetUpDescription()
	todoUI.SetUpForm()
	todoUI.SetUpInstructions("")

	leftTop := tview.NewFlex().SetDirection(tview.FlexColumn).AddItem(todoUI.TodoList, 0, 1, true).AddItem(todoUI.NoteList, 0, 1, false)

	leftSide := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(leftTop, 0, 2, true).
		AddItem(todoUI.Description, 0, 1, false)
	rightSide := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(todoUI.Form, 0, 1, false).AddItem(todoUI.Instructions, 0, 1, false)

	todoUI.App.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {

		prevFocus := action.CurrentFocus
		ctrlKey := event.Modifiers()&tcell.ModCtrl != 0

		switch {
		case event.Key() == tcell.KeyRight && ctrlKey:
			action.CurrentFocus = (action.CurrentFocus + 1) % len(focusModes)
		case event.Key() == tcell.KeyLeft && ctrlKey:
			action.CurrentFocus = (action.CurrentFocus - 1 + len(focusModes)) % len(focusModes)
		case event.Key() == tcell.KeyEscape:
			todoUI.App.Stop()
		case event.Key() == tcell.KeyEnter:
			if action.CurrentFocus == 0 {
				todoUI.DeleteItem(action.TodoMode, todoUI.TodoList.GetCurrentItem())
			} else {
				todoUI.DeleteItem(action.NoteMode, todoUI.NoteList.GetCurrentItem()+1)
			}
		}

		if prevFocus != action.CurrentFocus {
			todoUI.SetDescriptionOnFocus()
			todoUI.App.SetFocus(focusModes[action.CurrentFocus])
		}

		return event
	})

	root := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(title, 1, 1, false).
		AddItem(tview.NewFlex().
			AddItem(leftSide, 0, 2, true).
			AddItem(rightSide, 0, 1, false), 0, 1, true)
	todoUI.App.SetRoot(root, true)

	todoUI.RefreshItemList(action.TodoMode)
	todoUI.RefreshItemList(action.NoteMode)

	if err := todoUI.App.Run(); err != nil {
		log.Fatal(err)
	}
}
