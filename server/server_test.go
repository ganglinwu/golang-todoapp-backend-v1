package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
	"time"

	"github.com/ganglinwu/todoapp-backend-v1/errs"
	"github.com/ganglinwu/todoapp-backend-v1/models"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// initialize test suite struct
type TestSuite struct {
	suite.Suite
	server *TodoServer
}

type StubTodoStore struct {
	store []models.PROJECT
}

var (
	objID1, _ = bson.ObjectIDFromHex("67bc5c4f1e8db0c9a17efca0")
	objID2, _ = bson.ObjectIDFromHex("67e0c98b2c3e82a398cdbb16")
	objID3, _ = bson.ObjectIDFromHex("682571d1dafbee2eecbf4913")
	objID4, _ = bson.ObjectIDFromHex("682996bc78d219298228c10a")
	objID5, _ = bson.ObjectIDFromHex("68299585e7b6718ddf79b567")
	dueDate1  = time.Now().AddDate(0, 3, 0)
	dueDate2  = time.Now().AddDate(0, 0, 3)
	dueDate4  = time.Now().AddDate(0, 0, 3)

	// seed data
	todos1 = []models.TODO{
		{ID: &objID1, Name: "Water Plants", Description: "Not too much water for aloe vera", DueDate: &dueDate1},
		{ID: &objID2, Name: "Buy socks", Description: "No show socks", DueDate: &dueDate2},
	}
	todos2 = []models.TODO{
		{ID: &objID4, Name: "Test task 3", Description: "test description", DueDate: &dueDate4},
	}
	proj1 = models.PROJECT{ID: &objID3, ProjName: "proj1", Tasks: todos1}
	proj2 = models.PROJECT{ID: &objID5, ProjName: "proj2", Tasks: todos2}

	store = []models.PROJECT{proj1, proj2}
)

func TestServerTestSuite(t *testing.T) {
	suite.Run(t, &TestSuite{})
}

// This runs only once per suite
func (ts *TestSuite) SetupTest() {
	// initialize
	objID1, _ := bson.ObjectIDFromHex("67bc5c4f1e8db0c9a17efca0")
	objID2, _ := bson.ObjectIDFromHex("67e0c98b2c3e82a398cdbb16")
	objID3, _ := bson.ObjectIDFromHex("682571d1dafbee2eecbf4913")
	objID4, _ := bson.ObjectIDFromHex("682996bc78d219298228c10a")
	objID5, _ := bson.ObjectIDFromHex("68299585e7b6718ddf79b567")
	dueDate1 := time.Now().AddDate(0, 3, 0)
	dueDate2 := time.Now().AddDate(0, 0, 3)
	dueDate4 := time.Now().AddDate(0, 0, 3)

	// seed data
	todos1 := []models.TODO{
		{ID: &objID1, Name: "Water Plants", Description: "Not too much water for aloe vera", DueDate: &dueDate1},
		{ID: &objID2, Name: "Buy socks", Description: "No show socks", DueDate: &dueDate2},
	}
	todos2 := []models.TODO{
		{ID: &objID4, Name: "Test task 3", Description: "test description", DueDate: &dueDate4},
	}
	proj1 := models.PROJECT{ID: &objID3, ProjName: "proj1", Tasks: todos1}
	proj2 := models.PROJECT{ID: &objID5, ProjName: "proj2", Tasks: todos2}

	store := []models.PROJECT{proj1, proj2}
	ts.server = NewTodoServer(&StubTodoStore{store})
}

func (s *StubTodoStore) GetAllProjs() ([]models.PROJECT, error) {
	if len(s.store) == 0 {
		return []models.PROJECT{}, errs.ErrNotFound
	}
	return s.store, nil
}

func (s *StubTodoStore) GetAllTodos() ([]models.TODO, error) {
	if len(s.store) == 0 {
		return []models.TODO{}, errs.ErrNotFound
	}
	todos := []models.TODO{}

	for i := range s.store {
		for _, todo := range s.store[i].Tasks {
			todos = append(todos, todo)
		}
	}
	return todos, nil
}

func (s *StubTodoStore) GetProjByID(ID string) (models.PROJECT, error) {
	for _, proj := range s.store {
		if proj.ID.Hex() == ID {
			return proj, nil
		}
	}
	return models.PROJECT{}, errs.ErrNotFound
}

func (s *StubTodoStore) CreateProj(Name string, Tasks []models.TODO) (string, error) {
	randomObjID := bson.NewObjectID()
	IDstr := randomObjID.Hex()
	s.store = append(s.store, models.PROJECT{ID: &randomObjID, ProjName: Name, Tasks: Tasks})
	return IDstr, nil
}

func (s *StubTodoStore) CreateTodo(projID string, newTodoWithoutID models.TODO) (string, error) {
	if len(s.store) == 0 {
		return "", errs.ErrNotFound
	}
	for projIndex, proj := range s.store {
		if proj.ID.Hex() == projID {
			taskID := bson.NewObjectID()
			upsertedID := taskID.Hex()
			s.store[projIndex].Tasks = append(s.store[projIndex].Tasks, models.TODO{
				ID:          &taskID,
				Name:        newTodoWithoutID.Name,
				Description: newTodoWithoutID.Description,
				DueDate:     newTodoWithoutID.DueDate,
				Priority:    newTodoWithoutID.Priority,
			})
			return upsertedID, nil
		}
	}
	return "", errs.ErrNotFound
}

func (s *StubTodoStore) UpdateProjNameByID(ID, NewName string) error {
	if len(s.store) == 0 {
		return errs.ErrNotFound
	}
	for index, proj := range s.store {
		IDStr := proj.ID.Hex()
		if IDStr == ID {
			s.store[index].ProjName = NewName
		}
	}
	return nil
}

func (s *StubTodoStore) DeleteProjByID(ID string) (int, error) {
	if len(s.store) == 0 {
		return 0, errs.ErrNotFound
	}
	for i, proj := range s.store {
		if proj.ID.Hex() == ID {
			s.store = slices.Delete(s.store, i, i+1)
			return 1, nil
		}
	}
	return 0, errs.ErrNotFound
}

func (s *StubTodoStore) DeleteTodoByID(todoID string) (int, error) {
	if len(s.store) == 0 {
		return 0, errs.ErrNotFound
	}
	for projIndex, proj := range s.store {
		for taskIndex, task := range proj.Tasks {
			if task.ID.Hex() == todoID {
				s.store[projIndex].Tasks = slices.Delete(s.store[projIndex].Tasks, taskIndex, taskIndex+1)
				return 1, nil
			}
		}
	}
	return 0, errs.ErrNotFound
}

func (s *StubTodoStore) GetTodoByID(todoID string) (models.TODO, error) {
	if len(s.store) == 0 {
		return models.TODO{}, errs.ErrNotFound
	}
	for _, proj := range s.store {
		for _, task := range proj.Tasks {
			if task.ID.Hex() == todoID {
				return task, nil
			}
		}
	}
	return models.TODO{}, errs.ErrNotFound
}

func (s *StubTodoStore) UpdateTodoByID(ID string, newTodoWithoutID models.TODO) error {
	for projIndex, proj := range s.store {
		for taskIndex, task := range proj.Tasks {
			if task.ID.Hex() == ID {
				taskID, err := bson.ObjectIDFromHex(ID)
				if err != nil {
					return err
				}
				s.store[projIndex].Tasks[taskIndex].ID = &taskID
				s.store[projIndex].Tasks[taskIndex].Name = newTodoWithoutID.Name
				s.store[projIndex].Tasks[taskIndex].Description = newTodoWithoutID.Description
				s.store[projIndex].Tasks[taskIndex].DueDate = newTodoWithoutID.DueDate
				s.store[projIndex].Tasks[taskIndex].Priority = newTodoWithoutID.Priority
			}
		}
	}
	return nil
}

func (ts *TestSuite) TestGetAllProjs() {
	request, _ := http.NewRequest(http.MethodGet, "/proj", nil)
	responseRecorder := httptest.NewRecorder()

	ts.server.ServeHTTP(responseRecorder, request)

	response := responseRecorder.Result()

	got := []models.PROJECT{}

	err := json.NewDecoder(response.Body).Decode(&got)
	defer response.Body.Close()

	if err != nil {
		ts.FailNow(err.Error())
	}

	for index, proj := range got {
		ts.compareProjStructFields(store[index], proj)
	}

	ts.assertStatusCode(200, responseRecorder.Code)
}

func (ts *TestSuite) TestGetAllTodos() {
	request, _ := http.NewRequest(http.MethodGet, "/todo", nil)
	responseRecorder := httptest.NewRecorder()

	ts.server.ServeHTTP(responseRecorder, request)

	response := responseRecorder.Result()

	got := []models.TODO{}

	err := json.NewDecoder(response.Body).Decode(&got)
	defer response.Body.Close()

	if err != nil {
		ts.FailNow(err.Error())
	}
	objID1, _ := bson.ObjectIDFromHex("67bc5c4f1e8db0c9a17efca0")
	objID2, _ := bson.ObjectIDFromHex("67e0c98b2c3e82a398cdbb16")
	objID4, _ := bson.ObjectIDFromHex("682996bc78d219298228c10a")
	dueDate1 := time.Now().AddDate(0, 3, 0)
	dueDate2 := time.Now().AddDate(0, 0, 3)
	dueDate4 := time.Now().AddDate(0, 0, 3)

	want := []models.TODO{
		{ID: &objID1, Name: "Water Plants", Description: "Not too much water for aloe vera", DueDate: &dueDate1},
		{ID: &objID2, Name: "Buy socks", Description: "No show socks", DueDate: &dueDate2},
		{ID: &objID4, Name: "Test task 3", Description: "test description", DueDate: &dueDate4},
	}
	for i, todo := range want {
		ts.compareTodoStructFields(todo, got[i])
	}
}

func (ts *TestSuite) TestGetProjByID() {
	getTests := []struct {
		testname   string
		testpath   string
		want       models.PROJECT
		statusCode int
	}{
		{"get proj 1", "/proj/682571d1dafbee2eecbf4913", proj1, http.StatusOK},
		{"get proj 2", "/proj/68299585e7b6718ddf79b567", proj2, http.StatusOK},
		{"get non-existent proj 3", "/proj/3", models.PROJECT{}, http.StatusNotFound},
	}

	for _, test := range getTests {
		request, _ := http.NewRequest(http.MethodGet, test.testpath, nil)
		responseRecorder := httptest.NewRecorder()

		ts.server.ServeHTTP(responseRecorder, request)

		response := responseRecorder.Result()

		got := models.PROJECT{}

		err := json.NewDecoder(response.Body).Decode(&got)
		defer response.Body.Close()

		if err != nil {
			if !errors.Is(err, io.EOF) {
				ts.FailNow(err.Error())
			}
		}

		ts.compareProjStructFields(test.want, got)

		ts.assertStatusCode(test.statusCode, responseRecorder.Code)
	}
}

func (ts *TestSuite) TestCreateNewProj() {
	project := models.PROJECT{
		ProjName: "Test Project Name",
		Tasks:    []models.TODO{},
	}

	jsonData, err := json.Marshal(project)
	if err != nil {
		ts.FailNow(err.Error())
	}

	request, _ := http.NewRequest(http.MethodPost, "/proj/", bytes.NewBuffer(jsonData))
	response := httptest.NewRecorder()

	request.Header.Set("Content-Type", "application/json")

	ts.server.ServeHTTP(response, request)

	byteGot, _ := io.ReadAll(response.Result().Body)
	insertedIDString := string(byteGot)[:24]

	got, err := ts.server.TodoStore.GetProjByID(insertedIDString)
	if err != nil {
		ts.FailNow(err.Error())
	}

	insertedObjID, err := bson.ObjectIDFromHex(insertedIDString)
	if err != nil {
		ts.FailNow(err.Error())
	}
	want := models.PROJECT{ID: &insertedObjID, ProjName: "Test Project Name", Tasks: []models.TODO{}}

	ts.compareProjStructFields(want, got)
}

func (ts *TestSuite) TestCreateTodo() {
	// reset seeded data
	ts.SetupTest()

	timestamp := bson.Timestamp{T: uint32(time.Now().Unix())}

	todoToCreate := models.TODO{
		Name:          "Newly Created Task",
		Description:   "Newly Created Description",
		DueDateString: "2020-03-20T02:00:00+08:00",
		Priority:      "high",
		Completed:     false,
		Updated_at:    &timestamp,
	}

	jsonData, err := json.Marshal(todoToCreate)
	if err != nil {
		ts.FailNow(err.Error())
	}

	request, _ := http.NewRequest(http.MethodPost, "/proj/68299585e7b6718ddf79b567", bytes.NewBuffer(jsonData))
	responseRecorder := httptest.NewRecorder()

	request.Header.Set("Content-Type", "application/json")

	ts.server.ServeHTTP(responseRecorder, request)

	response := responseRecorder.Result()

	byteGot, err := io.ReadAll(response.Body)
	if err != nil {
		if err != io.EOF {
			ts.FailNow(err.Error())
		}
	}
	defer response.Body.Close()

	ts.assertStatusCode(201, responseRecorder.Code)

	insertedIDString := string(byteGot)[:24]
	log.Println("debug bytegot", string(byteGot))

	insertedObjID, err := bson.ObjectIDFromHex(insertedIDString)
	if err != nil {
		ts.FailNow(err.Error())
	}
	dueDate, err := time.Parse(time.RFC3339, "2020-03-20T02:00:00+08:00")
	if err != nil {
		ts.FailNow(err.Error())
	}

	got, err := ts.server.TodoStore.GetProjByID("68299585e7b6718ddf79b567")
	if err != nil {
		ts.FailNow(err.Error())
	}

	want := models.PROJECT{ID: &objID5, ProjName: "proj2", Tasks: todos2}
	want.Tasks = append(want.Tasks, models.TODO{
		ID:          &insertedObjID,
		Name:        "Newly Created Task",
		Description: "Newly Created Description",
		DueDate:     &dueDate,
		Priority:    "high",
		Completed:   false,
		Updated_at:  &timestamp,
	})

	ts.compareProjStructFields(want, got)
}

func (ts *TestSuite) TestUpdateProjNameByID() {
	updatedProj := models.PROJECT{
		ProjName: "Updated Proj Name",
	}

	jsonData, err := json.Marshal(updatedProj)
	if err != nil {
		ts.FailNow(err.Error())
	}

	request, _ := http.NewRequest(http.MethodPatch, "/proj/68299585e7b6718ddf79b567", bytes.NewBuffer(jsonData))
	responseRecorder := httptest.NewRecorder()

	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	ts.server.ServeHTTP(responseRecorder, request)

	got, err := ts.server.TodoStore.GetProjByID("68299585e7b6718ddf79b567")
	if err != nil {
		ts.FailNow(err.Error())
	}

	want := models.PROJECT{ID: &objID5, ProjName: "Updated Proj Name", Tasks: todos2}

	ts.compareProjStructFields(want, got)
}

func (ts *TestSuite) TestUpdateTodoByID() {
	// reset seeded data
	ts.SetupTest()

	timestamp := bson.Timestamp{T: uint32(time.Now().Unix())}

	todoToUpdate := models.TODO{
		Name:          "Updated Task",
		Description:   "Updated Description",
		DueDateString: "2025-03-20T02:00:00+08:00",
		Priority:      "low",
		Completed:     false,
		Updated_at:    &timestamp,
	}
	jsonData, err := json.Marshal(todoToUpdate)
	if err != nil {
		ts.FailNow(err.Error())
	}

	request, _ := http.NewRequest(http.MethodPatch, "/todo/682996bc78d219298228c10a", bytes.NewBuffer(jsonData))
	responseRecorder := httptest.NewRecorder()

	request.Header.Set("Content-Type", "application/json")

	ts.server.ServeHTTP(responseRecorder, request)

	got, err := ts.server.TodoStore.GetProjByID("68299585e7b6718ddf79b567")
	if err != nil {
		ts.FailNow(err.Error())
	}

	wantDueDate, err := time.Parse(time.RFC3339, "2025-03-20T02:00:00+08:00")
	if err != nil {
		ts.FailNow(err.Error())
	}

	want := models.PROJECT{
		ID:       &objID5,
		ProjName: "proj2",
		Tasks: []models.TODO{
			{ID: &objID4, Name: "Updated Task", Description: "Updated Description", DueDate: &wantDueDate, Priority: "low", Updated_at: &timestamp},
		},
	}
	ts.compareProjStructFields(want, got)
}

func (ts *TestSuite) TestDeleteProjByID() {
	request, _ := http.NewRequest(http.MethodDelete, "/proj/682571d1dafbee2eecbf4913", nil)
	responseRecorder := httptest.NewRecorder()

	ts.server.ServeHTTP(responseRecorder, request)
	response := responseRecorder.Result()

	byteGot, err := io.ReadAll(response.Body)
	if err != nil {
		if err != io.EOF {
			ts.FailNow(err.Error())
		}
	}

	defer response.Body.Close()
	got := string(byteGot)

	ts.assertTodoText("Number of projects deleted: 1", got)
	ts.assertStatusCode(200, responseRecorder.Code)
}

func (ts *TestSuite) TestDeleteTodoByID() {
	request, _ := http.NewRequest(http.MethodDelete, "/todo/67bc5c4f1e8db0c9a17efca0", nil)
	responseRecorder := httptest.NewRecorder()

	ts.server.ServeHTTP(responseRecorder, request)
	response := responseRecorder.Result()

	byteGot, err := io.ReadAll(response.Body)
	if err != nil {
		if err != io.EOF {
			ts.FailNow(err.Error())
		}
	}

	defer response.Body.Close()
	got := string(byteGot)

	ts.assertTodoText("Number of todos deleted: 1", got)
	ts.assertStatusCode(200, responseRecorder.Code)
}

/*
func TestGetAllTodo(t *testing.T) {
	request, _ := http.NewRequest(http.MethodGet, "/todo", nil)
	response := httptest.NewRecorder()

	s.ServeHTTP(response, request)

	var marshaledResponse []models.TODO
	err := json.NewDecoder(response.Body).Decode(&marshaledResponse)
	if err != nil {
		t.Fatal(err)
	}
	want := []models.TODO{
		{ID: &objID1, Name: "Water Plants", Description: "Not too much water for aloe vera", DueDate: &dueDate1},
		{ID: &objID2, Name: "Buy socks", Description: "No show socks", DueDate: &dueDate2},
	}
	if !reflect.DeepEqual(marshaledResponse, want) {
		t.Errorf("got %#v, want %#v", marshaledResponse, want)
	}
}


func TestDeleteTodoByID(t *testing.T) {
	deleteStore := []models.TODO{
		{ID: &objID1, Name: "Water Plants", Description: "Not too much water for aloe vera", DueDate: &dueDate1},
		{ID: &objID2, Name: "Buy socks", Description: "No show socks", DueDate: &dueDate2},
	}
	deleteServer := NewTodoServer(&StubTodoStore{deleteStore})

	deleteTests := []struct {
		testname   string
		testpath   string
		want       string
		statusCode int
	}{
		{"delete todo ID 1", "/todo/" + ID1, fmt.Sprintf("\n Sucessfully deleted todo \n ID: %s \n Name: %s \n Description: %s \n DueDate: %s", ID1, "Water Plants", "Not too much water for aloe vera", dueDate1), http.StatusOK},
		{"delete todo ID 2", "/todo/" + ID2, fmt.Sprintf("\n Sucessfully deleted todo \n ID: %s \n Name: %s \n Description: %s \n DueDate: %s", ID2, "Buy socks", "No show socks", dueDate2), http.StatusOK},
		{"delete non-existent todo ID 3", "/todo/3", "", http.StatusBadRequest},
	}

	for _, test := range deleteTests {
		t.Run(test.testname, func(t *testing.T) {
			request, _ := http.NewRequest(http.MethodDelete, test.testpath, nil)
			responseRecorder := httptest.NewRecorder()

			deleteServer.ServeHTTP(responseRecorder, request)
			response := responseRecorder.Result()

			byteGot, err := io.ReadAll(response.Body)
			if err != nil {
				if err != io.EOF {
					t.Fatal(err)
				}
			}

			defer response.Body.Close()
			got := string(byteGot)

			assertTodoText(t, got, test.want)
			assertStatusCode(t, responseRecorder.Code, test.statusCode)
		})
	}
}
*/
