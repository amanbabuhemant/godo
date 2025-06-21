package main

import (
	"github.com/biisal/godo/todos/action"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var ()

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
		Instructions: tview.NewTextView(),
		Description:  tview.NewTextView(),
	}

	todoUI.SetUpDescription()
	todoUI.SetUpList()
	todoUI.SetUpForm()
	todoUI.SetUpInstructions("")

	leftSide := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(todoUI.TodoList, 0, 2, true).
		AddItem(todoUI.Description, 0, 1, false)
	rightSide := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(todoUI.Form, 0, 1, false).AddItem(todoUI.Instructions, 0, 1, false)

	todoUI.App.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRight || event.Key() == tcell.KeyLeft {
			if action.FocusForm {
				todoUI.App.SetFocus(todoUI.TodoList)
			} else {
				todoUI.App.SetFocus(todoUI.Form)
			}
			action.FocusForm = !action.FocusForm
		} else if event.Key() == tcell.KeyEsc {
			todoUI.App.Stop()
		} else if event.Key() == tcell.KeyDelete {
			todoUI.DeleteTodo(todoUI.TodoList.GetCurrentItem() + 1)
		}
		return event
	})

	root := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(title, 1, 1, false).
		AddItem(tview.NewFlex().
			AddItem(leftSide, 0, 2, true).
			AddItem(rightSide, 0, 1, false), 0, 1, true)
	todoUI.App.SetRoot(root, true)

	todoUI.RefreshTodoList()

	if err := todoUI.App.Run(); err != nil {
		panic(err)
	}
}
