package mongostore

import (
	"github.com/ganglinwu/todoapp-backend-v1/models"
)

func (ts *TestSuite) compareTodoStructFields(got, want models.TODO) {
	ts.T().Helper()
	ts.Equal(got.ID, want.ID)
	ts.Equal(got.Name, want.Name)
	ts.Equal(got.Description, want.Description)

	ts.Equal(got.DueDate.Unix(), want.DueDate.Unix())
}
