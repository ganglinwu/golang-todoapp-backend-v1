package postgres_store

import (
	"log"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/suite"

	"github.com/ganglinwu/todoapp-backend-v1/models"
)

/*
* Test wide variables
*
 */
var (
	dueDate1 = time.Now().AddDate(0, 3, 0)
	dueDate2 = time.Now().AddDate(0, 0, 3)
	dueDate3 = time.Now().AddDate(0, 0, 3)

	todo1 = models.TODO{
		Id:          1,
		Name:        "Water Plants",
		Description: "Not too much water for aloe vera",
		DueDate:     &dueDate1,
		Priority:    "low",
		Completed:   false,
		ProjName:    "proj1",
	}
	todo2 = models.TODO{
		Id:          2,
		Name:        "Buy socks",
		Description: "No show socks",
		Priority:    "mid",
		DueDate:     &dueDate2,
		Completed:   false,
		ProjName:    "proj1",
	}

	todo3 = models.TODO{
		Id:          3,
		Name:        "Test task 3",
		Description: "test description",
		Priority:    "hi",
		DueDate:     &dueDate3,
		Completed:   false,
		ProjName:    "proj2",
	}

	proj1 = models.PROJECT{
		Id:       1,
		ProjName: "proj1",
	}
	proj2 = models.PROJECT{
		Id:       2,
		ProjName: "proj2",
	}
)

/*
* Test wide variables
*
 */

var MockPostGresStore = PostGresStore{}

type TestSuite struct {
	suite.Suite
	store *PostGresStore
}

func TestPostGresSuite(t *testing.T) {
	suite.Run(t, &TestSuite{})
}

// This runs only once per suite
func (ts *TestSuite) SetupSuite() {
	// fetch connection string for test database
	err := godotenv.Load(".env")
	if err != nil {
		ts.FailNow("unable to load .env")
	}

	connString, ok := os.LookupEnv("POSTGRES_CONNECTION_STRING_TEST")
	if !ok {
		ts.FailNow("unable to load connString from .env")
	}

	// connect
	db, err := NewConnection(connString)
	if err != nil {
		ts.FailNowf("unable to connect to azure postgres", err.Error())
	}

	ts.store = &PostGresStore{
		DB: db,
	}

	err = ts.store.DB.Ping()
	if err != nil {
		ts.FailNowf("failed to ping azure postgres", err.Error())
	}
}

// This runs before EVERY test
func (ts *TestSuite) SetupTest() {
	// clear DB
	_, err := ts.store.DB.Exec(`truncate table todos;`)
	if err != nil {
		log.Fatal("exec 1:", err.Error())
	}

	_, err = ts.store.DB.Exec(`drop table todos;`)
	if err != nil {
		log.Fatal("exec 2:", err.Error())
	}

	_, err = ts.store.DB.Exec(`drop table projects;`)
	if err != nil {
		log.Fatal("exec 2:", err.Error())
	}

	_, err = ts.store.DB.Exec(`create table if not exists projects (
    id SERIAL PRIMARY KEY,
    projname VARCHAR(255) NOT NULL UNIQUE
    );`)
	if err != nil {
		log.Fatal("exec 3:", err.Error())
	}

	_, err = ts.store.DB.Exec(`create table if not exists todos (
    id SERIAL PRIMARY KEY, 
    name VARCHAR(255) NOT NULL,
    description VARCHAR(255) NOT NULL,
    duedate TIMESTAMPTZ NOT NULL,
    priority VARCHAR(10) NOT NULL,
    completed BOOLEAN NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    projname VARCHAR(255) NOT NULL,
    FOREIGN KEY (projname) REFERENCES projects(projname) ON UPDATE CASCADE ON DELETE CASCADE
    );`)
	if err != nil {
		log.Fatal("exec 4:", err.Error())
	}

	ts.store.DB.Exec(`INSERT INTO projects (projname) VALUES ($1)`, proj1.ProjName)
	ts.store.DB.Exec(`INSERT INTO projects (projname) VALUES ($1)`, proj2.ProjName)

	ts.store.DB.Exec(`INSERT INTO todos (name, description, duedate, priority, completed, projname) VALUES($1, $2, $3, $4, $5, $6)`,
		todo1.Name,
		todo1.Description,
		todo1.DueDate,
		todo1.Priority,
		todo1.Completed,
		todo1.ProjName,
	)

	ts.store.DB.Exec(`INSERT INTO todos (name, description, duedate, priority, completed, projname) VALUES($1, $2, $3, $4, $5, $6)`,
		todo2.Name,
		todo2.Description,
		todo2.DueDate,
		todo2.Priority,
		todo2.Completed,
		todo2.ProjName,
	)

	ts.store.DB.Exec(`INSERT INTO todos (name, description, duedate, priority, completed, projname) VALUES($1, $2, $3, $4, $5, $6)`,
		todo3.Name,
		todo3.Description,
		todo3.DueDate,
		todo3.Priority,
		todo3.Completed,
		todo3.ProjName,
	)
}

func (ts *TestSuite) TestGetAllProjs() {
	got, err := ts.store.GetAllProjs()
	if err != nil {
		ts.FailNowf("err on GetAllProjs: ", err.Error())
	}

	want := []models.PROJECT{proj1, proj2}

	for i := 0; i < len(want); i++ {
		ts.compareProjStructFields(got[i], want[i])
	}
}

func (ts *TestSuite) TestGetAllTodos() {
	got, err := ts.store.GetAllTodos()
	if err != nil {
		ts.FailNowf("err on GetAllTodos: ", err.Error())
	}
	want := []models.TODO{todo1, todo2, todo3}

	for i := 0; i < len(want); i++ {
		ts.compareTodoStructFields(got[i], want[i])
	}
}

func (ts *TestSuite) TestGetProjByID() {
	got, err := ts.store.GetProjByID("1")
	if err != nil {
		ts.FailNowf("err on GetProjByID: ", err.Error())
	}
	want := proj1

	ts.compareProjStructFields(want, got)
}

func (ts *TestSuite) TestGetTodoByID() {
	got, err := ts.store.GetTodoByID("1")
	if err != nil {
		ts.FailNowf("err on GetTodoByID: ", err.Error())
	}
	want := todo1

	ts.compareTodoStructFields(want, got)
}

func (ts *TestSuite) TestCreateProj() {
	// flush and reset table
	ts.SetupTest()

	insertedProjID, err := ts.store.CreateProj("proj3", []models.TODO{})
	if err != nil {
		ts.FailNowf("err on CreateProj: ", err.Error())
	}

	got, err := ts.store.GetProjByID(insertedProjID)
	if err != nil {
		ts.FailNowf("err on GetProjByID: ", err.Error())
	}

	intID, err := strconv.Atoi(insertedProjID)
	if err != nil {
		ts.FailNowf("err on strconv.Atoi: ", err.Error())
	}

	want := models.PROJECT{
		Id:       intID,
		ProjName: "proj3",
	}

	ts.compareProjStructFields(want, got)
}

func (ts *TestSuite) TestCreateTodo() {
	newTodo := models.TODO{
		Name:        "Inserted Todo",
		Description: "Test",
		DueDate:     &dueDate1,
		Priority:    "low",
		Completed:   false,
		ProjName:    "proj2",
	}

	// CreateTodo returns int, err
	stringID, err := ts.store.CreateTodo("2", newTodo)
	if err != nil {
		ts.FailNowf("err on CreateTodo ", err.Error())
	}

	// TODO: fetch specific todo using GetTodoByID
	got, err := ts.store.GetAllTodos()
	if err != nil {
		ts.FailNowf("err on GetAllTodos ", err.Error())
	}

	intID, err := strconv.Atoi(stringID)
	if err != nil {
		ts.FailNowf("err on strconv.Atoi ", err.Error())
	}

	newTodo.Id = intID
	want := []models.TODO{todo1, todo2, todo3, newTodo}

	for i := range len(got) {
		ts.compareTodoStructFields(want[i], got[i])
	}
}

func (ts *TestSuite) TestUpdateProjNameByID() {
	err := ts.store.UpdateProjNameByID("1", "New proj1")
	if err != nil {
		ts.FailNowf("err on UpdateProjNameByID ", err.Error())
	}

	got, err := ts.store.GetProjByID("1")
	if err != nil {
		ts.FailNowf("err on GetProjByID ", err.Error())
	}

	want := models.PROJECT{
		Id:       1,
		ProjName: "New proj1",
	}

	ts.compareProjStructFields(want, got)
}

func (ts *TestSuite) TestUpdateTodoByID() {
	newDueDate := dueDate1.Add(3 * time.Hour)

	todoToUpdate := models.TODO{
		Name:        "Updated name",
		Description: "Updated description",
		DueDate:     &newDueDate,
		Priority:    "hi",
		Completed:   true,
		ProjName:    "proj2",
	}

	err := ts.store.UpdateTodoByID("1", todoToUpdate)
	if err != nil {
		ts.FailNowf("err on UpdateTodoByID ", err.Error())
	}

	// TODO: compare specific todo instead of all
	// especially postgres does sequential writes, thus the "order" of todos will not be the same as the index(id) number
	got, err := ts.store.GetAllTodos()
	if err != nil {
		ts.FailNowf("err on GetAllTodos ", err.Error())
	}

	todoToUpdate.Id = 1

	want := []models.TODO{todo2, todo3, todoToUpdate}

	for i := range got {
		ts.compareTodoStructFields(want[i], got[i])
	}
}

func (ts *TestSuite) TestDeleteProjByID() {
	deleteCount, err := ts.store.DeleteProjByID("1")
	if err != nil {
		ts.FailNowf("err on DeleteProjByID ", err.Error())
	}

	ts.Equal(1, deleteCount, "want 1 got %d", deleteCount)

	got, err := ts.store.GetAllProjs()
	if err != nil {
		ts.FailNowf("err on GetAllProjs ", err.Error())
	}

	want := []models.PROJECT{
		{Id: 2, ProjName: "proj2"},
	}

	for i := range got {
		ts.compareProjStructFields(want[i], got[i])
	}
}

func (ts *TestSuite) TestDeleteTodoByID() {
	deleteCount, err := ts.store.DeleteTodoByID("1")
	if err != nil {
		ts.FailNowf("err on DeleteTodoByID ", err.Error())
	}

	ts.Equal(1, deleteCount, "want 1 got %d", deleteCount)

	got, err := ts.store.GetAllTodos()
	if err != nil {
		ts.FailNowf("err on GetAllTodos ", err.Error())
	}

	want := []models.TODO{todo2, todo3}

	for i := range got {
		ts.compareTodoStructFields(want[i], got[i])
	}
}

/*
* methods to implement
type TodoStore interface {
	GetAllProjs() ([]models.PROJECT, error)
	GetAllTodos() ([]models.TODO, error)
	GetProjByID(ID string) (models.PROJECT, error)
	CreateProj(Name string, Tasks []models.TODO) (string, error)
	CreateTodo(projID string, newTodoWithoutID models.TODO) (string, error)
	UpdateProjNameByID(ID, newName string) error
	UpdateTodoByID(todoID string, newTodoWithoutID models.TODO) error
	DeleteProjByID(ID string) (int, error)
	DeleteTodoByID(todoID string) (int, error)
	GetTodoByID(todoID string) (models.TODO, error)
}
*/
