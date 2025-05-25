package mongostore

import (
	"context"
	"fmt"
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

func (ms *MongoStore) GetAllTodos() ([]models.TODO, error) {
	projs, err := ms.GetAllProjs()
	if err != nil {
		return []models.TODO{}, err
	}
	todos := []models.TODO{}
	for i := range projs {
		for _, todo := range projs[i].Tasks {
			todos = append(todos, todo)
		}
	}
	return todos, nil
}

func (ms *MongoStore) CreateTodo(projID string, newTodoWithoutID models.TODO) (*mongo.UpdateResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	query := bson.D{{"_id", projID}}

	update := bson.D{{Key: "$set", Value: bson.D{{"tasks", newTodoWithoutID}}}}

	opts := options.UpdateOne().SetUpsert(true)

	result, err := ms.Collection.UpdateOne(ctx, query, update, opts)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (ms *MongoStore) CreateProj(ProjName string, Tasks []models.TODO) (*bson.ObjectID, error) {
	// TODO: check if duplicate proj exists
	proj := models.PROJECT{ProjName: ProjName, Tasks: Tasks}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	result, err := ms.Collection.InsertOne(ctx, proj)
	if err != nil {
		return &bson.ObjectID{}, err
	}

	objID := result.InsertedID.(bson.ObjectID)

	return &objID, nil
}

func (ms *MongoStore) UpdateTodoByID(ID string, newTodoWithoutID models.TODO) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	objID, err := bson.ObjectIDFromHex(ID)
	if err != nil {
		return err
	}

	query := bson.D{{"tasks._id", &objID}}

	// we need to add in ID
	// else we will be updating with an object without ID!
	newTodoWithoutID.ID = &objID

	update := bson.D{{Key: "$set", Value: bson.D{{"tasks.$", newTodoWithoutID}}}}

	result, err := ms.Collection.UpdateOne(ctx, query, update)
	if err != nil {
		return err
	}
	// TODO: check result output
	// e.g. 0 0 0 <ID>
	// - MatchedCount, ModifiedCount, UpsertedCount, ObjectID of upserted document
	fmt.Println(result)
	return nil
}

func (ms *MongoStore) UpdateProjNameByID(ID, newProjName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	projID, err := bson.ObjectIDFromHex(ID)

	query := bson.D{{"_id", &projID}}

	update := bson.D{{"$set", bson.D{{"projname", newProjName}}}}

	result, err := ms.Collection.UpdateOne(ctx, query, update)
	if err != nil {
		return err
	}
	// TODO: check result output
	// e.g. 0 0 0 <ID>
	// - MatchedCount, ModifiedCount, UpsertedCount, ObjectID of upserted document
	fmt.Println(result)
	return nil
}

func (ms *MongoStore) DeleteProjByID(ID string) (*mongo.DeleteResult, error) {
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

func (ms *MongoStore) DeleteTodoByID(ProjID, TodoID string) (*mongo.UpdateResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	projID, err := bson.ObjectIDFromHex(ProjID)
	if err != nil {
		return nil, err
	}

	todoID, err := bson.ObjectIDFromHex(TodoID)
	if err != nil {
		return nil, err
	}

	query := bson.D{{"_id", &projID}}

	update := bson.D{{"$pull", bson.D{{"tasks", bson.D{{"_id", &todoID}}}}}}

	updateResult, err := ms.Collection.UpdateOne(ctx, query, update)
	if err != nil {
		return nil, err
	}
	return updateResult, nil
}
