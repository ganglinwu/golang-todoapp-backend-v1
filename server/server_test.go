package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/ganglinwu/todoapp-backend-v1/errs"
	"github.com/ganglinwu/todoapp-backend-v1/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type StubTodoStore struct {
	store []models.TODO
}

func (s *StubTodoStore) GetTodoByID(ID string) (models.TODO, error) {
	for _, todo := range s.store {
		if todo.ID.Hex() == ID {
			return todo, nil
		}
	}
	return models.TODO{}, errs.ErrNotFound
}

func (s *StubTodoStore) CreateTodo(Name, Description string, DueDate time.Time) (*bson.ObjectID, error) {
	createdTodo := models.TODO{
		ID:          &objID3,
		Name:        Name,
		Description: Description,
		DueDate:     &DueDate,
	}
	s.store = append(s.store, createdTodo)
	return createdTodo.ID, nil
}

func (s *StubTodoStore) GetAllTodos() ([]models.TODO, error) {
	if len(s.store) == 0 {
		return nil, errs.ErrNotFound
	}
	return s.store, nil
}

func (s *StubTodoStore) UpdateTodoByID(ID string, todo models.TODO) error {
	if len(s.store) == 0 {
		return errs.ErrNotFound
	}
	updatedTODO := todo
	objID, err := bson.ObjectIDFromHex(ID)
	if err != nil {
		return err
	}
	updatedTODO.ID = &objID

	existingTODO, err := s.GetTodoByID(ID)
	if err != nil {
		return err
	}

	if updatedTODO.Name == "" {
		updatedTODO.Name = existingTODO.Name
	}
	if updatedTODO.Description == "" {
		updatedTODO.Description = existingTODO.Description
	}
	if updatedTODO.DueDate == nil {
		updatedTODO.DueDate = existingTODO.DueDate
	}

	for i, todo := range s.store {
		if todo.ID.Hex() == ID {
			store = slices.Replace(store, i, i+1, updatedTODO)
			return nil
		}
	}
	return errs.ErrNotFound
}

func (s *StubTodoStore) DeleteTodoByID(ID string) (*mongo.DeleteResult, error) {
	if len(s.store) == 0 {
		return nil, errs.ErrNotFound
	}
	for i, todo := range s.store {
		if todo.ID.Hex() == ID {
			s.store = slices.Delete(s.store, i, i+1)
			return nil, nil
		}
	}
	return nil, errs.ErrNotFound
}

var (
	ID1       = "67bc5c4f1e8db0c9a17efca0"
	ID2       = "67e0c98b2c3e82a398cdbb16"
	ID3       = "67e0c98b2c3e82a398cdbb17"
	objID1, _ = bson.ObjectIDFromHex(ID1)
	objID2, _ = bson.ObjectIDFromHex(ID2)
	objID3, _ = bson.ObjectIDFromHex(ID3)
	dueDate1  = time.Now().AddDate(0, 3, 0).Truncate(time.Second)
	dueDate2  = time.Now().AddDate(0, 0, 3).Truncate(time.Second)
	store     = []models.TODO{
		{ID: &objID1, Name: "Water Plants", Description: "Not too much water for aloe vera", DueDate: &dueDate1},
		{ID: &objID2, Name: "Buy socks", Description: "No show socks", DueDate: &dueDate2},
	}
	s = NewTodoServer(&StubTodoStore{store})
)

func TestGetTodoByID(t *testing.T) {
	getTests := []struct {
		testname   string
		testpath   string
		want       models.TODO
		statusCode int
	}{
		{"get todo 1", "/todo/67bc5c4f1e8db0c9a17efca0", models.TODO{ID: &objID1, Name: "Water Plants", Description: "Not too much water for aloe vera", DueDate: &dueDate1}, http.StatusOK},
		{"get todo 2", "/todo/67e0c98b2c3e82a398cdbb16", models.TODO{ID: &objID2, Name: "Buy socks", Description: "No show socks", DueDate: &dueDate2}, http.StatusOK},
		{"get non-existent todo ID 3", "/todo/3", models.TODO{}, http.StatusNotFound},
	}

	for _, test := range getTests {
		t.Run(test.testname, func(t *testing.T) {
			request, _ := http.NewRequest(http.MethodGet, test.testpath, nil)
			responseRecorder := httptest.NewRecorder()

			s.ServeHTTP(responseRecorder, request)

			response := responseRecorder.Result()

			got := models.TODO{}

			err := json.NewDecoder(response.Body).Decode(&got)
			defer response.Body.Close()

			if err != nil {
				if !errors.Is(err, io.EOF) {
					t.Fatal(err)
				}
			}

			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("got %#v, want %#v \n", got, test.want)
			}

			assertStatusCode(t, responseRecorder.Code, test.statusCode)
		})
	}
}

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

func TestPostNewTodoByID(t *testing.T) {
	t.Run("post new todo", func(t *testing.T) {
		data := url.Values{
			"Name":        {"New Todo"},
			"Description": {"Test description"},
			"DueDate":     {dueDate1.Format(time.RFC3339)},
		}

		reader := strings.NewReader(data.Encode())

		request, _ := http.NewRequest(http.MethodPost, "/todo", reader)
		response := httptest.NewRecorder()

		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		s.ServeHTTP(response, request)

		byteGot, _ := io.ReadAll(response.Result().Body)
		got := string(byteGot)

		want := fmt.Sprintf("\n Sucessfully created todo \n ID: %s \n Name: %s \n Description: %s \n DueDate: %s", ID3, "New Todo", "Test description", dueDate1)

		assertTodoText(t, got, want)
		assertStatusCode(t, response.Code, http.StatusCreated)
	})
}

// refactor from here
func TestUpdateTodoByID(t *testing.T) {
	updateStore := []models.TODO{
		{ID: &objID1, Name: "Water Plants", Description: "Not too much water for aloe vera", DueDate: &dueDate1},
		{ID: &objID2, Name: "Buy socks", Description: "No show socks", DueDate: &dueDate2},
	}
	updateServer := NewTodoServer(&StubTodoStore{updateStore})

	updateTests := []struct {
		testname    string
		testpath    string
		updatedTODO models.TODO
		want        models.TODO
		statusCode  int
	}{
		{"update todo ID 1", "/todo/" + ID1, models.TODO{ID: &objID1, Description: "Even less for cactus", DueDate: &dueDate1}, models.TODO{ID: &objID1, Name: "Water Plants", Description: "Even less for cactus", DueDate: &dueDate1}, http.StatusOK},
		{"update todo ID 2", "/todo/" + ID2, models.TODO{ID: &objID2, Name: "Buy socks and underwear", DueDate: &dueDate2}, models.TODO{ID: &objID2, Name: "Buy socks and underwear", Description: "No show socks", DueDate: &dueDate2}, http.StatusOK},
		{"update non-existent todo ID 3", "/todo/" + ID3, models.TODO{ID: &objID3, DueDate: &dueDate1}, models.TODO{ID: &objID3, DueDate: &dueDate1}, http.StatusBadRequest},
	}

	for i, test := range updateTests {
		t.Run(test.testname, func(t *testing.T) {
			data := url.Values{
				"ID":          {test.updatedTODO.ID.Hex()},
				"Name":        {test.updatedTODO.Name},
				"Description": {test.updatedTODO.Description},
				"DueDate":     {test.updatedTODO.DueDate.Format(time.RFC3339)},
			}

			reader := strings.NewReader(data.Encode())

			request, _ := http.NewRequest(http.MethodPatch, test.testpath, reader)
			responseRecorder := httptest.NewRecorder()

			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			updateServer.ServeHTTP(responseRecorder, request)

			if responseRecorder.Result().StatusCode == http.StatusOK {

				got := store[i]

				if !reflect.DeepEqual(got, test.want) {
					t.Errorf("got %#v, want %#v", got, test.want)
				}
			}
			assertStatusCode(t, responseRecorder.Code, test.statusCode)
		})
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
