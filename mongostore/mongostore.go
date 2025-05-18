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

func (ms *MongoStore) GetProjByID(ID string) (models.PROJECT, error) {
	if ID == "" {
		return models.PROJECT{}, errs.ErrNotFound
	}
	objectID, err := bson.ObjectIDFromHex(ID)
	if err != nil {
		return models.PROJECT{}, err
	}
	filter := bson.D{{Key: "_id", Value: &objectID}}

	proj := models.PROJECT{}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	err = ms.Collection.FindOne(ctx, filter).Decode(&proj)
	if err != nil {
		return models.PROJECT{}, errs.ErrNotFound
	}

	return proj, nil
}

func (ms *MongoStore) GetAllProjs() ([]models.PROJECT, error) {
	projs := []models.PROJECT{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	filter := bson.D{{}}

	cursor, err := ms.Collection.Find(ctx, filter)
	if err != nil {
		return []models.PROJECT{}, err
	}

	err = cursor.All(ctx, &projs)
	if err != nil {
		return []models.PROJECT{}, err
	}

	return projs, nil
}

func (ms *MongoStore) CreateTodo(Name, Description string, DueDate time.Time) (*bson.ObjectID, error) {
	// TODO: check if duplicate todo exists
	todo := models.TODO{Name: Name, Description: Description, DueDate: &DueDate}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	result, err := ms.Collection.InsertOne(ctx, todo)
	if err != nil {
		return &bson.ObjectID{}, err
	}

	objID := result.InsertedID.(bson.ObjectID)

	return &objID, nil
}

/*
	func (ms *MongoStore) UpdateTodoByID(ID string, todo models.TODO) error {
		existingTodo, err := ms.GetProjByID(ID)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		objID, err := bson.ObjectIDFromHex(ID)
		if err != nil {
			return err
		}

		todo.ID = &objID
		if todo.Name == "" {
			todo.Name = existingTodo.Name
		}
		if todo.Description == "" {
			todo.Description = existingTodo.Description
		}
		if todo.DueDate == nil {
			todo.DueDate = existingTodo.DueDate
		}

		update := bson.D{{Key: "$set", Value: todo}}

		_, err = ms.Collection.UpdateByID(ctx, objID, update)
		if err != nil {
			return err
		}
		return nil
	}
*/
func (ms *MongoStore) DeleteTodoByID(ID string) (*mongo.DeleteResult, error) {
	_, err := ms.GetProjByID(ID)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	objID, err := bson.ObjectIDFromHex(ID)
	if err != nil {
		return nil, err
	}

	dr, err := ms.Collection.DeleteOne(ctx, bson.D{{Key: "_id", Value: objID}})
	if err != nil {
		return nil, err
	}
	return dr, nil
}
