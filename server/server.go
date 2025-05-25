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
	GetAllProjs() ([]models.PROJECT, error)
	GetAllTodos() ([]models.TODO, error)
	GetProjByID(ID string) (models.PROJECT, error)
	CreateProj(Name string, Tasks []models.TODO) (*bson.ObjectID, error)
	CreateTodo(projID string, newTodoWithoutID models.TODO) (*mongo.UpdateResult, error)
	UpdateProjNameByID(ID, newName string) error
	UpdateTodoByID(todoID string, newTodoWithoutID models.TODO) error
	DeleteProjByID(ID string) (*mongo.DeleteResult, error)
	DeleteTodoByID(ID string) (*mongo.UpdateResult, error)
}

type TodoServer struct {
	TodoStore TodoStore
	http.Handler
}

func NewTodoServer(store TodoStore) *TodoServer {
	r := http.NewServeMux()
	ts := &TodoServer{}
	ts.Handler = r
	ts.TodoStore = store

	r.HandleFunc("GET /proj", ts.handleGetAllProjs)
	r.HandleFunc("GET /todo", ts.handleGetAllTodos)
	r.HandleFunc("GET /proj/{ID}", ts.handleGetProjByID)
	r.HandleFunc("POST /proj", ts.handleCreateProj)
	r.HandleFunc("POST /proj/{ID}", ts.handleCreateTodo)
	r.HandleFunc("PATCH /proj/{ID}", ts.handleUpdateProjNameByID)
	r.HandleFunc("PATCH /todo/{ID}", ts.handleUpdateTodoByID)
	r.HandleFunc("DELETE /proj/{ID}", ts.handleDeleteProjByID)
	r.HandleFunc("DELETE /todo/{ID}", ts.handleDeleteTodoByID)
	return ts
}

func (ts TodoServer) handleGetAllProjs(w http.ResponseWriter, r *http.Request) {
	projs, err := ts.TodoStore.GetAllProjs()

	switch err {
	case errs.ErrNotFound:
		w.WriteHeader(http.StatusNotFound)
	case nil:
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(projs)
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

func (ts TodoServer) handleGetProjByID(w http.ResponseWriter, r *http.Request) {
	ID := r.PathValue("ID")
	todo, err := ts.TodoStore.GetProjByID(ID)

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

func (ts TodoServer) handleCreateProj(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", err.Error())
		return
	}
	projName := r.FormValue("ProjName")

	tasks := []models.TODO{}

	insertedID, err := ts.TodoStore.CreateProj(projName, tasks)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", err.Error())
		return
	}
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%s \n Sucessfully created proj \n ID: %s \n ProjName: %s \n Tasks: %#v \n", insertedID.Hex(), insertedID.Hex(), projName, tasks)
}

func (ts TodoServer) handleCreateTodo(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", err.Error())
		return
	}

	projID := r.PathValue("ID")
	dueDate, err := time.Parse(time.RFC3339, r.FormValue("DueDate"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", err.Error())
		return
	}

	newTodoWithoutID := models.TODO{
		Name:        r.FormValue("Name"),
		Description: r.FormValue("Description"),
		DueDate:     &dueDate,
		Priority:    r.FormValue("Priority"),
	}
	updateResult, err := ts.TodoStore.CreateTodo(projID, newTodoWithoutID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", err.Error())
		return
	}
	if updateResult.MatchedCount != 1 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "we could not find the project, thus was unable to create a new todo")
		return
	}
	if updateResult.UpsertedCount != 1 {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "something went wrong on our end, please try again later.")
		return
	}
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%s\n Successfully created todo \n ID: %s \n Name: %s \n Description: %s \n DueDate: %s \n Priority: %s \n", updateResult.UpsertedID, updateResult.UpsertedID, r.FormValue("Name"), r.FormValue("Description"), r.FormValue("DueDate"), r.FormValue("Priority"))
	return
}

func (ts TodoServer) handleUpdateProjNameByID(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", err.Error())
		return
	}

	ID := r.PathValue("ID")
	newProjName := r.FormValue("ProjName")

	err = ts.TodoStore.UpdateProjNameByID(ID, newProjName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "\n Sucessfully updated proj name \n ID: %s \n ProjName: %s \n", ID, newProjName)
	return
}

func (ts TodoServer) handleUpdateTodoByID(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", err.Error())
		return
	}

	ID := r.PathValue("ID")
	dueDate, err := time.Parse(time.RFC3339, r.FormValue("DueDate"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", err.Error())
		return
	}

	newTodoWithoutID := models.TODO{
		Name:        r.FormValue("Name"),
		Description: r.FormValue("Description"),
		DueDate:     &dueDate,
		Priority:    r.FormValue("Priority"),
	}

	err = ts.TodoStore.UpdateTodoByID(ID, newTodoWithoutID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "\n Sucessfully updated todo \n ID: %s \n Name: %s \n Description: %s \n DueDate: %s \n Priority: %s \n", ID, newTodoWithoutID.Name, newTodoWithoutID.Description, r.FormValue("DueDate"), newTodoWithoutID.Priority)
	return
}

func (ts TodoServer) handleDeleteProjByID(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", err.Error())
		return
	}
	ID := r.PathValue("ID")

	deleteResult, err := ts.TodoStore.DeleteProjByID(ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%d", deleteResult.DeletedCount)
	return
}

func (ts TodoServer) handleDeleteTodoByID(w http.ResponseWriter, r *http.Request) {
	todoID := r.PathValue("ID")

	updateResult, err := ts.TodoStore.DeleteTodoByID(todoID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", err.Error())
		return
	}
	if updateResult.MatchedCount != 1 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "the project/todo could not be found")
		return
	}
	if updateResult.ModifiedCount != 1 {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "we could not delete the todo. something went wrong on our end.")
		return
	}
	w.WriteHeader(http.StatusOK)
	return
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
