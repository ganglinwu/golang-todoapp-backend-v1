package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ganglinwu/todoapp-backend-v1/errs"
	"github.com/ganglinwu/todoapp-backend-v1/models"
)

type StubTodoStore struct {
	store map[string]string
}

func (s *StubTodoStore) GetTodoByID(ID string) (models.TODO, error) {
	for key, value := range s.store {
		if key == ID {
			return models.TODO{ID: key, Description: value}, nil
		}
	}
	return models.TODO{}, errs.ErrNotFound
}

func (s *StubTodoStore) CreateTodoByID(ID, description string) error {
	_, exist := s.store[ID]
	if exist {
		return errs.ErrIdAlreadyInUse
	} else {
		s.store[ID] = description
		return nil
	}
}

func TestGetTodoByID(t *testing.T) {
	store := map[string]string{
		"1": "Hello there!",
		"2": "Water plants",
	}

	s := NewTodoServer(&StubTodoStore{store})

	getTests := []struct {
		testname   string
		testpath   string
		want       string
		statusCode int
	}{
		{"get todo ID 1", "/todo/1", "Hello there!", http.StatusOK},
		{"get todo ID 2", "/todo/2", "Water plants", http.StatusOK},
		{"get non-existent todo ID 3", "/todo/3", "", http.StatusNotFound},
	}

	for _, test := range getTests {
		t.Run(test.testname, func(t *testing.T) {
			request, _ := http.NewRequest(http.MethodGet, test.testpath, nil)
			response := httptest.NewRecorder()

			s.ServeHTTP(response, request)

			got := response.Body.String()

			assertTodoText(t, got, test.want)
			assertStatusCode(t, response.Code, test.statusCode)
		})
	}
}

func TestPostNewTodoByID(t *testing.T) {
	store := map[string]string{
		"1": "Hello there!",
		"2": "Water plants",
	}

	s := NewTodoServer(&StubTodoStore{store})

	t.Run("post new todo on id 3", func(t *testing.T) {
		reader := strings.NewReader("Description=Save the earth")

		request, _ := http.NewRequest(http.MethodPost, "/todo/3", reader)
		response := httptest.NewRecorder()

		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		s.ServeHTTP(response, request)

		got := response.Body.String()
		want := "Sucessfully created todo ID 3: Save the earth"

			got := response.Body.String()

			assertTodoText(t, got, test.want)
			assertStatusCode(t, response.Code, test.statusCode)
		})
	}
}
