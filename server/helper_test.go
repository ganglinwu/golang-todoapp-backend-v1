package server

import (
	"testing"
)

func assertTodoText(t testing.TB, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func assertStatusCode(t testing.TB, gotCode, wantCode int) {
	t.Helper()
	if gotCode != wantCode {
		t.Errorf("got %d, want %d", gotCode, wantCode)
	}
}
