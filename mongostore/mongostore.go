package mongostore

import (
	"context"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/ganglinwu/todoapp-backend-v1/errs"
	"github.com/ganglinwu/todoapp-backend-v1/models"
	"github.com/joho/godotenv"
)

type MongoStore struct {
	Conn       *mongo.Client
	Collection *mongo.Collection
}

func NewConnection() (*mongo.Client, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	connString, exist := os.LookupEnv("CONNECTION_STRING")
	if !exist {
		return nil, errs.ErrEnvVarNotFound
	}

	bsonOpts := &options.BSONOptions{
		ObjectIDAsHexString: true,
		UseJSONStructTags:   true,
		UseLocalTimeZone:    true,
	}

	conn, err := mongo.Connect(options.Client().ApplyURI(connString).SetBSONOptions(bsonOpts))
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func GetDBNameCollectionName() (string, string, error) {
	err := godotenv.Load()
	if err != nil {
		return "", "", err
	}

	dbName, exists := os.LookupEnv("DATABASE_NAME")
	if !exists {
		return "", "", errs.ErrEnvVarNotFound
	}
	collName, exists := os.LookupEnv("COLLECTION_NAME")
	if !exists {
		return dbName, "", errs.ErrEnvVarNotFound
	}
	return dbName, collName, nil
}

func (ms *MongoStore) GetTodoByID(ID string) (models.TODO, error) {
	objectID, err := bson.ObjectIDFromHex(ID)
	if err != nil {
		return models.TODO{}, err
	}
	filter := bson.D{{Key: "_id", Value: &objectID}}

	todo := models.TODO{}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	err = ms.Collection.FindOne(ctx, filter).Decode(&todo)
	if err != nil {
		return models.TODO{}, err
	}

	return todo, nil
}

func (ms *MongoStore) GetAllTodos() ([]models.TODO, error) {
	todos := []models.TODO{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	filter := bson.D{{}}

	cursor, err := ms.Collection.Find(ctx, filter)
	if err != nil {
		return []models.TODO{}, err
	}

	err = cursor.All(ctx, &todos)
	if err != nil {
		return []models.TODO{}, err
	}

	return todos, nil
}

/*
func (ms *MongoStore) CreateTodoByID(ID, description string) error {
	_, exists := i.Store[ID]
	if exists {
		return errs.ErrIdAlreadyInUse
	} else {
		i.Store[ID] = description
		return nil
	}
}


func (ms *MongoStore) UpdateTodoByID(ID, description string) (models.TODO, error) {
	if len(i.Store) == 0 {
		return models.TODO{}, errs.ErrNotFound
	}
	for key := range i.Store {
		if key == ID {
			i.Store[key] = description
			return models.TODO{ID: key, Description: description}, nil
		}
	}
	return models.TODO{}, errs.ErrNotFound
}

func (ms *MongoStore) DeleteTodoByID(ID string) error {
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
*/
