package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ganglinwu/todoapp-backend-v1/models"
)

type StubTodoStore struct {
	store map[string]string
}

func (s *StubTodoStore) GetTodoByID(ID string) models.TODO {
	for key, value := range s.store {
		if key == ID {
			return models.TODO{ID: key, Description: value}
		}
	}
	return models.TODO{}
}

func TestGetTodoByID(t *testing.T) {
	store := map[string]string{
		"1": "Hello there!",
	}

	s := &TodoServer{&StubTodoStore{store}}

	testname1 := "Get todo ID 1"
	t.Run(testname1, func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/todo/1", nil)
		response := httptest.NewRecorder()

		s.ServeHTTP(response, request)

		got := response.Body.String()
		want := "Hello there!"

		assertTodoText(t, got, want)
	})
}
