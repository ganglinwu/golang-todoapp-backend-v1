package mongostore

import (
	"github.com/ganglinwu/todoapp-backend-v1/models"
)

func (ts *TestSuite) compareTodoStructFields(want, got models.TODO) {
	ts.T().Helper()
	ts.Equal(want.ID, got.ID)
	ts.Equal(want.Name, got.Name)
	ts.Equal(want.Description, got.Description)

	ts.Equal(want.DueDate.Unix(), got.DueDate.Unix())
}

func (ts *TestSuite) compareProjStructFields(want, got models.PROJECT) {
	ts.T().Helper()
	ts.Equal(want.ID, got.ID)
	ts.Equal(want.ProjName, got.ProjName)
	for i, todo := range got.Tasks {
		ts.compareTodoStructFields(want.Tasks[i], todo)
	}
}
