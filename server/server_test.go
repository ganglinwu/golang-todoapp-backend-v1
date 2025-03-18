package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetTodoByID(t *testing.T) {
	testname1 := "Get todo ID 1"
	t.Run(testname1, func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/todo/1", nil)
		response := httptest.NewRecorder()

		TodoServer(response, request)

		got := response.Body.String()
		want := "Hello there!"

		assertTodoText(t, got, want)
	})
}
