package inmemorystore

import (
	"github.com/ganglinwu/todoapp-backend-v1/errs"
	"github.com/ganglinwu/todoapp-backend-v1/models"
)

type InMemoryStore struct {
	Store map[string]string
}

func (i *InMemoryStore) GetTodoByID(ID string) (models.MockTODO, error) {
	for key, value := range i.Store {
		if key == ID {
			return models.MockTODO{ID: key, Description: value}, nil
		}
	}
	return models.MockTODO{}, errs.ErrNotFound
}

func (i *InMemoryStore) CreateTodoByID(ID, description string) error {
	_, exists := i.Store[ID]
	if exists {
		return errs.ErrIdAlreadyInUse
	} else {
		i.Store[ID] = description
		return nil
	}
}

func (i *InMemoryStore) GetAllTodos() ([]models.MockTODO, error) {
	todos := []models.MockTODO{}
	if len(i.Store) == 0 {
		return nil, errs.ErrNotFound
	}
	for key, value := range i.Store {
		todo := models.MockTODO{ID: key, Description: value}
		todos = append(todos, todo)
	}
	return todos, nil
}

func (i *InMemoryStore) UpdateTodoByID(ID, description string) (models.MockTODO, error) {
	if len(i.Store) == 0 {
		return models.MockTODO{}, errs.ErrNotFound
	}
	for key := range i.Store {
		if key == ID {
			i.Store[key] = description
			return models.MockTODO{ID: key, Description: description}, nil
		}
	}
	return models.MockTODO{}, errs.ErrNotFound
}

func (i *InMemoryStore) DeleteTodoByID(ID string) error {
	if len(i.Store) == 0 {
		return errs.ErrNotFound
	}
	for key := range i.Store {
		if key == ID {
			delete(i.Store, key)
			return nil
		}
	}
	return errs.ErrNotFound
}
