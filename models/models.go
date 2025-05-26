package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Option func(*TODO) error

const (
  TodoDefaultName = "Name of task/todo"
  TodoDefaultDescription = "Description of task/todo"
  TodoDefaultCompleted = false
)
var (
  TodoDefaultDueDate = func() {return time.Now().Add(3*time.Hours)}()
)

/*
* omitempty does not work on
* bson.ObjectID
* time.Time
* because they are structs with zero values
* whereas omitempty will only omit nil structs

	type TODO struct {
		ID          bson.ObjectID `json:"_id,omitempty"`
		Name        string        `json:"name"`
		Description string        `json:"description,omitempty"`
		DueDate     time.Time     `json:"dueDate,omitempty"`
	}
*/
type TODO struct {
	ID          *bson.ObjectID `json:"_id,omitempty"`
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	DueDate     *time.Time     `json:"dueDate,omitempty"`
	Priority    string         `json:"priority,omitempty"`
	Completed   bool           `json:"completed,omitempty`
}

type PROJECT struct {
	ID       *bson.ObjectID `json:"_id,omitempty"`
	ProjName string         `json:"projname"`
	Tasks    []TODO         `json:"tasks"`
}

func WithName(name string) func(*TODO) error {
  return func(todo *TODO) error {
    if name == "" {
      return errs.ErrTodoFieldEmpty
    } else if len(name) > 255 {
      return errs.ErrTodoNameTooLong
    }
    todo.Name = name
    return nil
  }
}

func WithDescription(description string) func(*TODO) error {
  return func(todo *TODO) error {
    todo.Description = description
    return nil
  }
}

func WithDueDate(dateString string) func(*TODO) error {
  return func (todo *TODO) error {
    duedate, err := time.Parse(time.RFC3339, dateString)
    if err != nil {
      return err
    }
    roundedDueDate := dueDate.Round(time.Day)
    roundedDefaultDueDate := TodoDefaultDueDate.Round(time.Day)
    // Compare returns -1 if receiver(i.e. dueDate) is before argument(i.e. TodoDefaultDueDate)
    if roundedDueDate.Compare(roundedDefaultDueDate) == -1 {
      return err.ErrTodoDueDateInThePast
    }
  }
}

func WithCompleted(completed bool) func(*TODO) error {
  return func(todo *TODO) error {
    todo.Completed = completed
  }
}

func NewTodo(opts.. Option) (TODO, error) {
  todo := TODO{
    Name: TodoDefaultName,
    Description: TodoDefaultDescription,
    Completed: TodoDefaultCompleted,
  }

  for _, opt := range opts {
    if err := opt(&todo); err != nil {
      return TODO{}, fmt.Errorf("Failed to construct todo: \n %w", err)
    }
  }
  return todo, nil
}
