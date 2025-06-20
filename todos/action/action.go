package action

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/biisal/godo/todos/models"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var (
	TodoFilePath string
	TodosCount   = 0
	FocusForm    = false
	Instructions = []string{
		"Use Right > or Left < to toggle between form and list",
		"Use Tab to toggle between form inputs or todos",
		"Use Enter on todos to Toggle Done",
		"For delete enter id of todo and press Del",
		"Esc/Ctrl+C to quit",
	}
)

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		panic("Failed to get user home directory: " + err.Error())
	}
	TodoFilePath = home + "/.local/share/godo/todos.json"
}

type TodoUI struct {
	App          *tview.Application
	Form         *tview.Form
	TodoList     *tview.List
	Instructions *tview.TextView
	Description  *tview.TextView
}

func GetTodos() ([]models.Todo, error) {
	if _, err := os.Stat(TodoFilePath); os.IsNotExist(err) {
		dir := filepath.Dir(TodoFilePath)
		if err = os.MkdirAll(dir, os.ModePerm); err != nil {
			return nil, err
		}
		if _, err = os.Create(TodoFilePath); err != nil {
			return nil, err
		}
	}
	file, err := os.Open(TodoFilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var todos []models.Todo
	err = json.NewDecoder(file).Decode(&todos)
	if err != nil {
		if err == io.EOF {
			TodosCount = 0
			return []models.Todo{}, nil
		}
		return nil, err
	}
	TodosCount = len(todos)
	return todos, nil
}

func WriteTodos(todos []models.Todo) error {
	file, err := os.Create(TodoFilePath)
	if err != nil {
		return err
	}
	defer file.Close()
	return json.NewEncoder(file).Encode(todos)
}

func (b *TodoUI) MarkDone(id int) error {
	todos, err := GetTodos()
	if err != nil {
		return err
	}
	if TodosCount < id {
		return nil
	}
	todos[id-1].Done = !todos[id-1].Done
	err = WriteTodos(todos)
	if err != nil {
		return err
	}
	b.RefreshTodoList()
	return nil
}

func (b *TodoUI) RefreshTodoList() {
	b.TodoList.Clear()
	todos, err := GetTodos()
	if err != nil {
		return
	}
	for index, todo := range todos {
		doneText := "⨯"
		if todo.Done {
			doneText = "✓"
		}
		b.TodoList.AddItem(fmt.Sprintf("[%d][%s] %s", todo.ID, doneText, todo.Title), "", 0, func() {
			b.MarkDone(todo.ID)
			b.TodoList.SetCurrentItem(index)

		})
	}

}

func (b *TodoUI) DeleteTodo(id int) error {
	todos, err := GetTodos()
	if err != nil {
		return err
	}
	if TodosCount < id {
		return nil
	}
	todos = slices.Delete(todos, id-1, id)
	for i := range todos {
		todos[i].ID = i + 1
	}
	err = WriteTodos(todos)
	if err != nil {
		return err
	}
	b.RefreshTodoList()
	return nil
}

func (b *TodoUI) AddTodo(title, description string) error {
	if title == "" || description == "" {
		return fmt.Errorf("empty title and description cannot be your todo :)")
	}
	todo := models.Todo{
		ID:          TodosCount + 1,
		Title:       title,
		Description: description,
		Done:        false,
	}

	todos, err := GetTodos()
	if err != nil {
		return err
	}
	todos = append(todos, todo)
	err = WriteTodos(todos)
	if err != nil {
		return err
	}
	b.RefreshTodoList()
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
		err := b.AddTodo(titleInput.GetText(), descriptionInput.GetText())
		if err == nil {
			titleInput.SetText("")
			descriptionInput.SetText("")
			b.SetUpInstructions("")
			b.App.SetFocus(b.TodoList)
			b.TodoList.SetCurrentItem(TodosCount)
			FocusForm = !FocusForm

		} else {
			if strings.HasPrefix(err.Error(), "empty") {
				b.SetUpInstructions("\n" + err.Error())
			}
		}
	}).AddFormItem(deleteIdInput).AddButton("Delete", func() {
		id, err := strconv.Atoi(deleteIdInput.GetText())
		if err != nil {
			return
		}
		b.DeleteTodo(id)
	}).AddButton("Clear", func() {
		err := os.Remove(TodoFilePath)
		if err != nil {
			b.SetUpInstructions("\n" + err.Error())
		}
		b.RefreshTodoList()
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
			todos, _ := GetTodos()
			if index < len(todos) {
				b.Description.SetText(todos[index].Description)
			}
		}).SetBackgroundColor(tcell.ColorDefault).SetBorder(true).SetTitle(" WE HAVE TO DO THIS ")
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
