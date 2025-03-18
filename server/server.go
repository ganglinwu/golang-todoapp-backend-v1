package server

import (
	"net/http"
	"strings"

	"github.com/ganglinwu/todoapp-backend-v1/models"
)

type TodoStore interface {
	GetTodoByID(ID string) models.TODO
}

type TodoServer struct {
	TodoStore TodoStore
}

func (ts *TodoServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ID := strings.TrimPrefix(r.URL.Path, "/todo/")

	todo := ts.TodoStore.GetTodoByID(ID)

	w.Write([]byte(todo.Description))
}
