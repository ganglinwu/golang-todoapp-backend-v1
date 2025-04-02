package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ganglinwu/todoapp-backend-v1/errs"
	"github.com/ganglinwu/todoapp-backend-v1/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type TodoStore interface {
	GetTodoByID(ID string) (models.TODO, error)
	CreateTodo(Name, Description string, DueDate time.Time) (*bson.ObjectID, error)
	GetAllTodos() ([]models.TODO, error)
	UpdateTodoByID(ID string, todo models.TODO) error
	DeleteTodoByID(ID string) (*mongo.DeleteResult, error)
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
	r.HandleFunc("POST /todo", ts.handlePostTodoByID)
	r.HandleFunc("PATCH /todo/{ID}", ts.handleUpdateTodoByID)
	r.HandleFunc("DELETE /todo/{ID}", ts.handleDeleteTodoByID)

	return ts
}

func (ts TodoServer) handleGetAllTodos(w http.ResponseWriter, r *http.Request) {
	todos, err := ts.TodoStore.GetAllTodos()

	switch err {
	case errs.ErrNotFound:
		w.WriteHeader(http.StatusNotFound)
	case nil:
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(todos)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "%s", err.Error())
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
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(todo)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "%s", err.Error())
			return
		}

		w.WriteHeader(http.StatusOK)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (ts TodoServer) handlePostTodoByID(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", err.Error())
		return
	}

	ID := r.FormValue("ID")
	_, err = ts.TodoStore.GetTodoByID(ID)
	switch err {
	case nil:
		w.WriteHeader(http.StatusBadRequest)
		return
	case errs.ErrNotFound:

		duedate, err := time.Parse(time.RFC3339, r.FormValue("DueDate"))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "%s", err.Error())
			return
		}

		name := r.FormValue("Name")
		description := r.FormValue("Description")

		insertedID, err := ts.TodoStore.CreateTodo(name, description, duedate)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "%s", err.Error())
			return
		}

		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "\n Sucessfully created todo \n ID: %s \n Name: %s \n Description: %s \n DueDate: %s", insertedID.Hex(), name, description, duedate)
	}
}

func (ts TodoServer) handleUpdateTodoByID(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", err.Error())
		return
	}

	todo := models.TODO{}
	ID := r.PathValue("ID")

	objID, err := bson.ObjectIDFromHex(ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", err.Error())
		return
	}

	duedate, err := time.Parse(time.RFC3339, r.FormValue("DueDate"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", err.Error())
		return
	}

	todo.ID = &objID
	todo.Name = r.FormValue("Name")
	todo.Description = r.FormValue("Description")
	todo.DueDate = &duedate

	_, err = ts.TodoStore.GetTodoByID(ID)
	switch err {
	case errs.ErrNotFound:
		w.WriteHeader(http.StatusBadRequest)
		return
	case nil:

		err = ts.TodoStore.UpdateTodoByID(ID, todo)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "%s", err.Error())
			return
		}

		err = json.NewEncoder(w).Encode(todo)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "%s", err.Error())
			return
		}

		w.WriteHeader(http.StatusOK)
		return
	}
}

func (ts TodoServer) handleDeleteTodoByID(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", err.Error())
		return
	}

	ID := r.PathValue("ID")

	todo, err := ts.TodoStore.GetTodoByID(ID)
	switch err {
	case errs.ErrNotFound:
		w.WriteHeader(http.StatusBadRequest)
	case nil:
		_, err = ts.TodoStore.DeleteTodoByID(ID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "%s", err.Error())
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "\n Sucessfully deleted todo \n ID: %s \n Name: %s \n Description: %s \n DueDate: %s", ID, todo.Name, todo.Description, todo.DueDate)
	default:
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "something went wrong on our end. err: %s", err.Error())
	}
}

/*
func handleErrAsHTTP501(w http.ResponseWriter, e error) {
	if e != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", e.Error())
		return
	}
}

func handleErrAsHTTP400(w http.ResponseWriter, e error) {
	if e != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

*/
