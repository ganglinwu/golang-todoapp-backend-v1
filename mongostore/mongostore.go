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

func NewConnection(connString *string) (*mongo.Client, error) {
	if *connString == "" {
		err := godotenv.Load()
		if err != nil {
			return nil, err
		}

		mongoDSN, exist := os.LookupEnv("MONGO_CONNECTION_STRING")
		if !exist {
			return nil, errs.ErrEnvVarNotFound
		}
		*connString = mongoDSN
	}

	bsonOpts := &options.BSONOptions{
		ObjectIDAsHexString: true,
		UseJSONStructTags:   true,
		UseLocalTimeZone:    true,
	}

	conn, err := mongo.Connect(options.Client().ApplyURI(*connString).SetBSONOptions(bsonOpts))
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func GetDBNameCollectionName(dbName, collName *string) (*string, *string, error) {
	if *dbName == "" {
		err := godotenv.Load()
		if err != nil {
			return nil, nil, err
		}

		name, exists := os.LookupEnv("DATABASE_NAME")
		if !exists {
			return nil, nil, errs.ErrEnvVarNotFound
		}
		*dbName = name
	}

	if *collName == "" {
		name, exists := os.LookupEnv("COLLECTION_NAME")
		if !exists {
			return dbName, nil, errs.ErrEnvVarNotFound
		}
		*collName = name
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
		todos = append(todos, projs[i].Tasks...)
	}
	return todos, nil
}

func (ms *MongoStore) CreateTodo(projID string, newTodoWithoutID models.TODO) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	objID, err := bson.ObjectIDFromHex(projID)
	if err != nil {
		return "", err
	}

	query := bson.D{{Key: "_id", Value: &objID}}

	// generate new ObjectID for created todo
	todoID := bson.NewObjectID()

	upsertedID := ""

	// add ID onto newTodoWithoutID
	if newTodoWithoutID.ID == nil {
		newTodoWithoutID.ID = &todoID
		upsertedID = todoID.Hex()
	} else {
		upsertedID = newTodoWithoutID.ID.Hex()
	}

	update := bson.D{{Key: "$push", Value: bson.D{{Key: "tasks", Value: newTodoWithoutID}}}}

	opts := options.UpdateOne().SetUpsert(true)

	_, err = ms.Collection.UpdateOne(ctx, query, update, opts)
	if err != nil {
		return "", err
	}

	return upsertedID, nil
}

func (ms *MongoStore) CreateProj(ProjName string, Tasks []models.TODO) (string, error) {
	// TODO: check if duplicate proj exists
	proj := models.PROJECT{ProjName: ProjName, Tasks: Tasks}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	result, err := ms.Collection.InsertOne(ctx, proj)
	if err != nil {
		return "", err
	}

	objID := result.InsertedID.(bson.ObjectID)
	IDstr := objID.Hex()

	return IDstr, nil
}

func (ms *MongoStore) UpdateTodoByID(ID string, newTodoWithoutID models.TODO) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	objID, err := bson.ObjectIDFromHex(ID)
	if err != nil {
		return err
	}

	query := bson.D{{Key: "tasks._id", Value: &objID}}

	// we need to add in ID
	// else we will be updating with an object without ID!
	newTodoWithoutID.ID = &objID

	update := bson.D{{Key: "$set", Value: bson.D{{Key: "tasks.$", Value: newTodoWithoutID}}}}

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
	if err != nil {
		return err
	}

	query := bson.D{{Key: "_id", Value: &projID}}

	update := bson.D{{Key: "$set", Value: bson.D{{Key: "projname", Value: newProjName}}}}

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

func (ms *MongoStore) DeleteProjByID(ID string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	objID, err := bson.ObjectIDFromHex(ID)
	if err != nil {
		return 0, err
	}

	dr, err := ms.Collection.DeleteOne(ctx, bson.D{{Key: "_id", Value: objID}})
	if err != nil {
		return 0, err
	}
	return int(dr.DeletedCount), nil
}

func (ms *MongoStore) DeleteTodoByID(TodoID string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	todoID, err := bson.ObjectIDFromHex(TodoID)
	if err != nil {
		return 0, err
	}

	query := bson.D{{Key: "tasks._id", Value: &todoID}}

	update := bson.D{{Key: "$pull", Value: bson.D{{Key: "tasks", Value: bson.D{{Key: "_id", Value: &todoID}}}}}}

	updateResult, err := ms.Collection.UpdateOne(ctx, query, update)
	if err != nil {
		return 0, err
	}

	deletedCount := updateResult.ModifiedCount

	return int(deletedCount), nil
}

func (ms *MongoStore) GetTodoByID(TodoID string) (models.TODO, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	todoID, err := bson.ObjectIDFromHex(TodoID)
	if err != nil {
		return models.TODO{}, err
	}

	projThatContainsTodo := models.PROJECT{}

	query := bson.D{{Key: "tasks._id", Value: &todoID}}

	err = ms.Collection.FindOne(ctx, query).Decode(&projThatContainsTodo)
	if err != nil {
		return models.TODO{}, err
	}

	for _, todo := range projThatContainsTodo.Tasks {
		if todo.ID.Hex() == TodoID {
			return todo, nil
		}
	}
	return models.TODO{}, errs.ErrNotFound
}
