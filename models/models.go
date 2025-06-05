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
	ID          *bson.ObjectID `json:"_id,omitempty"`
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	DueDate     *time.Time     `json:"dueDate,omitempty"`
	Priority    string         `json:"priority,omitempty"`
	Completed   bool           `json:"completed"`
}

type PROJECT struct {
	ID       *bson.ObjectID `json:"_id,omitempty"`
	ProjName string         `json:"projname"`
	Tasks    []TODO         `json:"tasks"`
}
