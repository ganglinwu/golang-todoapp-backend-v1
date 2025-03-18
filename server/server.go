package server

import (
	"net/http"
	"strings"

	"github.com/ganglinwu/todoapp-backend-v1/errs"
	"github.com/ganglinwu/todoapp-backend-v1/models"
)

type TodoStore interface {
	GetTodoByID(ID string) (models.TODO, error)
}

type TodoServer struct {
	TodoStore TodoStore
}

func (ts *TodoServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ID := strings.TrimPrefix(r.URL.Path, "/todo/")

	todo, err := ts.TodoStore.GetTodoByID(ID)
	switch err {
	case errs.ErrNotFound:
		w.WriteHeader(http.StatusNotFound)
	case nil:
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(todo.Description))
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
}
