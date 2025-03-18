package inmemorystore

import "github.com/ganglinwu/todoapp-backend-v1/models"

type InMemoryStore struct {
	Store map[string]string
}

func (i *InMemoryStore) GetTodoByID(ID string) models.TODO {
	for key, value := range i.Store {
		if key == ID {
			return models.TODO{ID: key, Description: value}
		}
	}
	return models.TODO{}
}
