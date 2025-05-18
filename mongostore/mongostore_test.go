package mongostore

import (
	"context"
	"testing"
	"time"

	"github.com/ganglinwu/todoapp-backend-v1/models"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// to prevent import cycle
// we mock up a server
type MockTodoServer struct {
	store *MongoStore
}

// initialize test suite struct
type TestSuite struct {
	suite.Suite
	collection *mongo.Collection
	server     *MockTodoServer
}

func TestMongoTestSuite(t *testing.T) {
	suite.Run(t, &TestSuite{})
}

// This runs only once per suite
func (ts *TestSuite) SetupSuite() {
	// connect
	conn, err := NewConnection()
	if err != nil {
		ts.FailNowf("unable to connect to mongoDB Atlas", err.Error())
	}

	dbName, _, err := GetDBNameCollectionName()
	if err != nil {
		ts.FailNowf("unable to load env variables", err.Error())
	}

	ts.collection = conn.Database(dbName).Collection("testTodo")

	ts.server = &MockTodoServer{(&MongoStore{conn, ts.collection})}
}

// This runs before EVERY test
func (ts *TestSuite) SetupTest() {
	// clear DB
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	filter := bson.D{{}}

	_, err := ts.collection.DeleteMany(ctx, filter)
	if err != nil {
		ts.FailNowf("unable to drop all entries from database", err.Error())
	}

	objID1, _ := bson.ObjectIDFromHex("67bc5c4f1e8db0c9a17efca0")
	objID2, _ := bson.ObjectIDFromHex("67e0c98b2c3e82a398cdbb16")
	objID3, _ := bson.ObjectIDFromHex("682571d1dafbee2eecbf4913")
	objID4, _ := bson.ObjectIDFromHex("682996bc78d219298228c10a")
	objID5, _ := bson.ObjectIDFromHex("68299585e7b6718ddf79b567")
	dueDate1 := time.Now().AddDate(0, 3, 0)
	dueDate2 := time.Now().AddDate(0, 0, 3)
	dueDate4 := time.Now().AddDate(0, 0, 3)

	// seed data
	todos1 := []models.TODO{
		{ID: &objID1, Name: "Water Plants", Description: "Not too much water for aloe vera", DueDate: &dueDate1},
		{ID: &objID2, Name: "Buy socks", Description: "No show socks", DueDate: &dueDate2},
	}
	todos2 := []models.TODO{
		{ID: &objID4, Name: "Test task 3", Description: "test description", DueDate: &dueDate4},
	}
	proj1 := models.PROJECT{ID: &objID3, ProjName: "proj1", Tasks: todos1}
	proj2 := models.PROJECT{ID: &objID5, ProjName: "proj2", Tasks: todos2}

	projSlice := []models.PROJECT{proj1, proj2}

	_, err = ts.collection.InsertMany(ctx, projSlice)
	if err != nil {
		ts.FailNowf("error inserting into mongo atlas", err.Error())
	}
}

// This runs after EVERY test
func (ts *TestSuite) TearDownTest() {
}

func (ts *TestSuite) TestGetProjByID() {
	got, err := ts.server.store.GetProjByID("682571d1dafbee2eecbf4913")

	objID1, _ := bson.ObjectIDFromHex("67bc5c4f1e8db0c9a17efca0")
	objID2, _ := bson.ObjectIDFromHex("67e0c98b2c3e82a398cdbb16")
	objID3, _ := bson.ObjectIDFromHex("682571d1dafbee2eecbf4913")
	dueDate1 := time.Now().AddDate(0, 3, 0)
	dueDate2 := time.Now().AddDate(0, 0, 3)

	todos := []models.TODO{
		{ID: &objID1, Name: "Water Plants", Description: "Not too much water for aloe vera", DueDate: &dueDate1},
		{ID: &objID2, Name: "Buy socks", Description: "No show socks", DueDate: &dueDate2},
	}
	want := models.PROJECT{ID: &objID3, ProjName: "proj1", Tasks: todos}

	if err != nil {
		ts.FailNowf("err on GetProjByID: ", err.Error())
	}

	ts.compareProjStructFields(want, got)
}

func (ts *TestSuite) TestGetAllProjs() {
	got, err := ts.server.store.GetAllProjs()
	if err != nil {
		ts.FailNowf("err on GetAllProjs: ", err.Error())
	}

	objID1, _ := bson.ObjectIDFromHex("67bc5c4f1e8db0c9a17efca0")
	objID2, _ := bson.ObjectIDFromHex("67e0c98b2c3e82a398cdbb16")
	objID3, _ := bson.ObjectIDFromHex("682571d1dafbee2eecbf4913")
	objID4, _ := bson.ObjectIDFromHex("682996bc78d219298228c10a")
	objID5, _ := bson.ObjectIDFromHex("68299585e7b6718ddf79b567")
	dueDate1 := time.Now().AddDate(0, 3, 0)
	dueDate2 := time.Now().AddDate(0, 0, 3)
	dueDate4 := time.Now().AddDate(0, 0, 3)

	todos1 := []models.TODO{
		{ID: &objID1, Name: "Water Plants", Description: "Not too much water for aloe vera", DueDate: &dueDate1},
		{ID: &objID2, Name: "Buy socks", Description: "No show socks", DueDate: &dueDate2},
	}
	todos2 := []models.TODO{
		{ID: &objID4, Name: "Test task 3", Description: "test description", DueDate: &dueDate4},
	}
	proj1 := models.PROJECT{ID: &objID3, ProjName: "proj1", Tasks: todos1}
	proj2 := models.PROJECT{ID: &objID5, ProjName: "proj2", Tasks: todos2}

	want := []models.PROJECT{proj1, proj2}

	for i := 0; i < len(want); i++ {
		ts.compareProjStructFields(got[i], want[i])
	}
}

/*
func (ts *TestSuite) TestCreateTodo() {
	name := "new todo to be inserted"
	description := "test description"
	time := time.Now()
	insertedID, err := ts.server.store.CreateTodo(name, description, time)
	if err != nil {
		ts.FailNowf("err on CreateTodo: ", err.Error())
	}

	got, err := ts.server.store.GetTodoByID(insertedID.Hex())
	if err != nil {
		ts.FailNowf("failed to convert bson.ObjectID to Hex string:", err.Error())
	}

	want := models.TODO{ID: insertedID, Name: name, Description: description, DueDate: &time}

	ts.compareTodoStructFields(got, want)
}

func (ts *TestSuite) TestUpdateTodoByID() {
	ID := "67bc5c4f1e8db0c9a17efca0"
	objID1, _ := bson.ObjectIDFromHex("67bc5c4f1e8db0c9a17efca0")

	dueDate1 := time.Now().AddDate(0, 2, 0)
	todo := models.TODO{
		Name:        "Water plants and trim leaves",
		Description: "Water all plants, trim only dead leaves",
		DueDate:     &dueDate1,
	}

	err := ts.server.store.UpdateTodoByID(ID, todo)
	if err != nil {
		ts.FailNowf("err on UpdateTodoByID: ", err.Error())
	}

	got, _ := ts.server.store.GetTodoByID(ID)
	want := todo

	want.ID = &objID1

	ts.compareTodoStructFields(got, want)
}

func (ts *TestSuite) TestDeleteTodoByID() {
	ID := "67bc5c4f1e8db0c9a17efca0"
	dr, err := ts.server.store.DeleteTodoByID(ID)
	if err != nil {
		ts.FailNowf("err on DeleteTodoByID: ", err.Error())
	}
	got := dr.DeletedCount
	want := int64(1)

	ts.Equal(got, want)
}
*/
