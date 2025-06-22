package action

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strconv"

	"github.com/biisal/godo/todos/models"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var (
	TodoFilePath  string
	NotesFilePath string
	TodoMode      = "todoMod"
	NoteMode      = "noteMod"
	FormMode      = "formMod"
	Modes         = []string{TodoMode, NoteMode, FormMode}
	CurrentFocus  = 0
	TodosCount    = 0
	NotesCount    = 0
	FocusForm     = false
	Instructions  = []string{
		"Use Right > or Left < to toggle between form and list",
		"Use Tab to toggle between form inputs or todos",
		"Use Enter on todos to Toggle Done",
		"For delete enter id of todo and press Del",
		"Esc/Ctrl+C to quit",
		"",
		"",
		"Repo: https://github.com/biisal/godo",
	}
	ErrorEmpty = errors.New("Empty title or description cannot be your todo :)")
)

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		panic("Failed to get user home directory: " + err.Error())
	}
	TodoFilePath = home + "/.local/share/godo/todos.json"
	NotesFilePath = home + "/.local/share/godo/notes.json"
}

type TodoUI struct {
	App          *tview.Application
	Form         *tview.Form
	TodoList     *tview.List
	Instructions *tview.TextView
	Description  *tview.TextView
	NoteList     *tview.List
}

func GetItems(mode string) ([]models.Item, error) {
	path := TodoFilePath
	if mode == NoteMode {
		path = NotesFilePath
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		dir := filepath.Dir(path)
		if err = os.MkdirAll(dir, os.ModePerm); err != nil {
			return nil, err
		}
		if _, err = os.Create(path); err != nil {
			return nil, err
		}
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var items []models.Item
	err = json.NewDecoder(file).Decode(&items)
	if err != nil {
		if err == io.EOF {
			TodosCount = 0
			return []models.Item{}, nil
		}
		return nil, err
	}
	if mode == TodoMode {
		TodosCount = len(items)
	} else {
		NotesCount = len(items)
	}
	return items, nil
}

func WriteItem(mode string, items []models.Item) error {
	path := TodoFilePath
	if mode == NoteMode {
		path = NotesFilePath
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return json.NewEncoder(file).Encode(items)
}

func (b *TodoUI) MarkDone(id int) error {
	todos, err := GetItems(TodoMode)
	if err != nil {
		return err
	}
	if TodosCount < id {
		return nil
	}
	todos[id-1].Done = !todos[id-1].Done
	err = WriteItem(TodoMode, todos)
	if err != nil {
		return err
	}
	b.RefreshItemList(TodoMode)
	return nil
}

func (b *TodoUI) RefreshItemList(mode string) {
	list := b.TodoList
	if mode == NoteMode {
		list = b.NoteList
	}
	list.Clear()
	todos, err := GetItems(mode)
	if err != nil {
		return
	}
	maxIDLen := len(strconv.Itoa(TodosCount))
	for index, todo := range todos {
		idStr := fmt.Sprintf("%0*d", maxIDLen, todo.ID)

		doneText := "⨯"
		doneColor := ""
		if todo.Done {
			doneText = "✓"
			doneColor = "[#508878]"
		}
		list.AddItem(fmt.Sprintf("%s[%s] [%s] │ %s", doneColor, idStr, doneText, todo.Title), "", 0, func() {
			b.MarkDone(todo.ID)
			list.SetCurrentItem(index)
		})
	}
}

func (b *TodoUI) DeleteItem(mode string, id int) error {
	todos, err := GetItems(mode)
	if err != nil {
		return err
	}
	if (mode == TodoMode && TodosCount < id) || (mode == NoteMode && NotesCount < id) {
		return nil
	}
	todos = slices.Delete(todos, id-1, id)
	for i := range todos {
		todos[i].ID = i + 1
	}
	err = WriteItem(mode, todos)
	if err != nil {
		return err
	}
	b.RefreshItemList(mode)
	return nil
}

func (b *TodoUI) AddItem(mode, title, description string) error {
	if title == "" || description == "" {
		return ErrorEmpty
	}
	newId := TodosCount + 1
	if mode == NoteMode {
		newId = NotesCount + 1
	}
	todo := models.Item{
		ID:          newId,
		Title:       title,
		Description: description,
		Done:        false,
	}

	todos, err := GetItems(mode)
	if err != nil {
		return err
	}
	todos = append(todos, todo)
	err = WriteItem(mode, todos)
	if err != nil {
		return err
	}
	b.RefreshItemList(mode)
	return nil
}

func (b *TodoUI) SetUpForm() {
	titleInput := tview.NewInputField().
		SetLabel("Title: ").SetFieldBackgroundColor(tcell.ColorRed)
	titleInput.SetFieldTextColor(tcell.ColorWhite).SetBackgroundColor(tcell.ColorRed)

	descriptionInput := tview.NewInputField().SetLabel("Description: ").SetFieldBackgroundColor(tcell.ColorBlack)
	descriptionInput.SetFieldTextColor(tcell.ColorWhite)

	deleteIdInput := tview.NewInputField().SetLabel("ID (for delete) : ").SetFieldBackgroundColor(tcell.ColorBlack)
	deleteIdInput.SetFieldTextColor(tcell.ColorWhite)

	b.Form.AddFormItem(titleInput).AddFormItem(descriptionInput).AddButton("Add", func() {
		err := b.AddItem(TodoMode, titleInput.GetText(), descriptionInput.GetText())
		if err == nil {
			titleInput.SetText("")
			descriptionInput.SetText("")
			b.SetUpInstructions("")
			b.App.SetFocus(b.TodoList)
			b.TodoList.SetCurrentItem(TodosCount)
			FocusForm = !FocusForm

		} else {
			if errors.Is(err, ErrorEmpty) {
				b.SetUpInstructions("\n" + err.Error())
			}
		}
	}).AddFormItem(deleteIdInput).AddButton("Delete", func() {
		id, err := strconv.Atoi(deleteIdInput.GetText())
		if err != nil {
			return
		}
		b.DeleteItem(TodoMode, id)
	}).AddButton("Add Note", func() {
		err := b.AddItem(NoteMode, titleInput.GetText(), descriptionInput.GetText())
		if err == nil {
			titleInput.SetText("")
			descriptionInput.SetText("")
			b.SetUpInstructions("")
			b.App.SetFocus(b.TodoList)
			b.TodoList.SetCurrentItem(TodosCount)
			FocusForm = !FocusForm

		} else {
			if errors.Is(err, ErrorEmpty) {
				b.SetUpInstructions("\n" + err.Error())
			}
		}
	}).
		AddButton("Clear", func() {
			err := os.Remove(TodoFilePath)
			if err != nil {
				b.SetUpInstructions("\n" + err.Error())
			}
			b.RefreshItemList(TodoMode)
		})

	b.Form.SetBackgroundColor(tcell.ColorDefault).SetBorder(true).SetTitle(" ADD OR DEL A TODO ")
}

func (b *TodoUI) SetUpList() {
	b.TodoList.SetSelectedTextColor(tcell.NewRGBColor(255, 255, 255)).SetSelectedBackgroundColor(tcell.ColorDefault).
		SetChangedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
			if index < 0 {
				return
			}
			b.Description.Clear()
			todos, _ := GetItems(TodoMode)
			if index < len(todos) {
				b.Description.SetText(todos[index].Description)
			}
		}).SetBackgroundColor(tcell.ColorDefault).SetBorder(true).SetTitle(" NOTES ")
	b.TodoList.ShowSecondaryText(false)
	b.TodoList.SetHighlightFullLine(true)
	b.TodoList.SetSelectedBackgroundColor(tcell.NewHexColor(0x00f5ff))
	b.TodoList.SetSelectedTextColor(tcell.NewHexColor(0x000000))

}
func (b *TodoUI) SetUpNoteList() {
	b.NoteList.SetSelectedTextColor(tcell.NewRGBColor(255, 255, 255)).SetSelectedBackgroundColor(tcell.ColorDefault).
		SetChangedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
			if index < 0 {
				return
			}
			b.Description.Clear()
			todos, _ := GetItems(NoteMode)
			if index < len(todos) {
				b.Description.SetText(todos[index].Description)
			}
		}).SetBackgroundColor(tcell.ColorDefault).SetBorder(true).SetTitle(" WE HAVE TO DO THIS ")
	b.NoteList.ShowSecondaryText(false)
	b.NoteList.SetHighlightFullLine(true)
	b.NoteList.SetSelectedBackgroundColor(tcell.NewHexColor(0x00f5ff))
	b.NoteList.SetSelectedTextColor(tcell.NewHexColor(0x000000))

}
func (b *TodoUI) SetUpInstructions(appendText string) {
	b.Instructions.
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(true)
	b.Instructions.Clear()
	b.Instructions.SetBackgroundColor(tcell.ColorDefault).SetBorder(true).SetTitle(" INSTRUCTIONS ")
	for _, value := range Instructions {
		fmt.Fprintln(b.Instructions, value)
	}
	if appendText != "" {
		fmt.Fprintln(b.Instructions, appendText)
	}
}

func (b *TodoUI) SetUpDescription() {
	b.Description.
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(true).SetBackgroundColor(tcell.ColorDefault).SetBorder(true).SetTitle(" DESCRIPTION ")
}
