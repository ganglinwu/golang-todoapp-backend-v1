package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type MockTODO struct {
	ID          string
	Description string
}

/*
* omitempty does not work on
* bson.ObjectID
* time.Time
* because they are structs with zero values
* whereas omitempty will only omit nil structs

	type TODO struct {
		ID          bson.ObjectID `json:"_id,omitempty"`
		Name        string        `json:"name"`
		Description string        `json:"description,omitempty"`
		DueDate     time.Time     `json:"dueDate,omitempty"`
	}
*/
type TODO struct {
	ID            *bson.ObjectID  `json:"_id,omitempty" db:"-"` // mongodb id
	Id            int             `json:"-" db:"id"`            // postgresql id
	Name          string          `json:"name" db:"name"`
	Description   string          `json:"description,omitempty" db:"description"`
	DueDate       *time.Time      `json:"dueDate,omitempty" db:"duedate"`
	DueDateString string          `json:"dueDateString,omitempty" db:"-"`
	Priority      string          `json:"priority,omitempty" db:"priority"`
	Completed     bool            `json:"completed" db:"completed"`
	Updated_at    *bson.Timestamp `json:"updated_at" db:"-"`
	Updated_At    *time.Time      `json:"-" db:"updated_at`
	ProjName      string          `json:"-" db:"projname"`
}

type PROJECT struct {
	ID       *bson.ObjectID `json:"_id,omitempty" db:"-"`
	Id       int            `json:"-" db:"id"`
	ProjName string         `json:"projname" db:"projname"`
	Tasks    []TODO         `json:"tasks" db:"-"`
}
