package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/ganglinwu/todoapp-backend-v1/errs"
	"github.com/ganglinwu/todoapp-backend-v1/models"
)

type TodoStore interface {
	GetTodoByID(ID string) (models.TODO, error)
	CreateTodoByID(ID, description string) error
	GetAllTodos() ([]models.TODO, error)
	UpdateTodoByID(ID, newDescription string) (models.TODO, error)
	DeleteTodoByID(ID string) error
}

/*
type MockTodoStore interface {
	GetTodoByID(ID string) (models.MockTODO, error)
	CreateTodoByID(ID, description string) error
	GetAllTodos() ([]models.MockTODO, error)
	UpdateTodoByID(ID, newDescription string) (models.MockTODO, error)
	DeleteTodoByID(ID string) error
}
*/

type TodoServer struct {
	TodoStore TodoStore
	http.Handler
}

func NewTodoServer(store TodoStore) *TodoServer {
	r := http.NewServeMux()
	ts := &TodoServer{}
	ts.Handler = r
	ts.TodoStore = store

	r.HandleFunc("GET /todo", ts.handleGetAllTodos)
	r.HandleFunc("GET /todo/{ID}", ts.handleGetTodoByID)
	r.HandleFunc("POST /todo/{ID}", ts.handlePostTodoByID)
	r.HandleFunc("PATCH /todo/{ID}", ts.handleUpdateTodoByID)
	r.HandleFunc("DELETE /todo/{ID}", ts.handleDeleteTodoByID)

	return ts
}

func (ts TodoServer) handleGetAllTodos(w http.ResponseWriter, r *http.Request) {
	todos, err := ts.TodoStore.GetAllTodos()

	// since our mock store is a map..
	// it may not be arranged by ID
	// in a real database we can return sort by ID
	// so let's sort our slice
	sort.Slice(todos, func(i, j int) bool {
		ID1, _ := strconv.Atoi(todos[i].ID)
		ID2, _ := strconv.Atoi(todos[j].ID)
		return ID1 < ID2
	})

	switch err {
	case errs.ErrNotFound:
		w.WriteHeader(http.StatusNotFound)
	case nil:
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(todos)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Something went wrong on our side. err: %s", err.Error())
			return
		}
		w.WriteHeader(http.StatusOK)
	default:
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "err: %s", err.Error())
	}
}

func (ts TodoServer) handleGetTodoByID(w http.ResponseWriter, r *http.Request) {
	ID := r.PathValue("ID")
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

func (ts TodoServer) handlePostTodoByID(w http.ResponseWriter, r *http.Request) {
	ID := r.PathValue("ID")
	_, err := ts.TodoStore.GetTodoByID(ID)
	switch err {
	case nil:
		w.WriteHeader(http.StatusBadRequest)
	case errs.ErrNotFound:

		read, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		description := strings.TrimPrefix(string(read), "Description=")

		err = ts.TodoStore.CreateTodoByID(ID, description)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "Sucessfully created todo ID %s: %s", ID, description)
	}
}

func (ts TodoServer) handleUpdateTodoByID(w http.ResponseWriter, r *http.Request) {
	ID := r.PathValue("ID")
	_, err := ts.TodoStore.GetTodoByID(ID)
	switch err {
	case errs.ErrNotFound:
		w.WriteHeader(http.StatusBadRequest)
	case nil:
		read, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		newDescription := string(read)
		todo, err := ts.TodoStore.UpdateTodoByID(ID, newDescription)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		err = json.NewEncoder(w).Encode(todo)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func (ts TodoServer) handleDeleteTodoByID(w http.ResponseWriter, r *http.Request) {
	ID := r.PathValue("ID")
	todo, err := ts.TodoStore.GetTodoByID(ID)
	switch err {
	case errs.ErrNotFound:
		w.WriteHeader(http.StatusBadRequest)
	case nil:
		err = ts.TodoStore.DeleteTodoByID(ID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "%s", err.Error())
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Sucessfully deleted todo ID %s: %s", ID, todo.Description)
	default:
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "something went wrong on our end. err: %s", err.Error())
	}
}
