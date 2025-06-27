package postgres_store

import (
	"github.com/ganglinwu/todoapp-backend-v1/models"
)

func (ts *TestSuite) compareTodoStructFields(want, got models.TODO) {
	ts.T().Helper()
	ts.Equal(want.Id, got.Id)
	ts.Equal(want.Name, got.Name)
	ts.Equal(want.Description, got.Description)
	ts.Equal(want.Completed, got.Completed)
	ts.Equal(want.Priority, got.Priority)
	ts.Equal(want.ProjName, got.ProjName)

	if got.DueDate != nil {
		if want.DueDate != nil {
			ts.InDelta(want.DueDate.Unix(), got.DueDate.Unix(), 5)
		} else {
			ts.FailNow("got DueDate but did not want one")
		}
	}
}

func (ts *TestSuite) compareProjStructFields(want, got models.PROJECT) {
	ts.T().Helper()
	ts.Equal(want.Id, got.Id)
	ts.Equal(want.ProjName, got.ProjName)
}
