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

// start test suite
func TestIntTestSuite(t *testing.T) {
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
	dueDate1 := time.Now().AddDate(0, 3, 0)
	dueDate2 := time.Now().AddDate(0, 0, 3)

	// seed data
	todos := []interface{}{
		models.TODO{ID: &objID1, Name: "Water Plants", Description: "Not too much water for aloe vera", DueDate: &dueDate1},
		models.TODO{ID: &objID2, Name: "Buy socks", Description: "No show socks", DueDate: &dueDate2},
	}

	_, err = ts.collection.InsertMany(ctx, todos)
	if err != nil {
		ts.FailNowf("error inserting into mongo atlas", err.Error())
	}
}

// This runs after EVERY test
func (ts *TestSuite) TearDownTest() {
}

func TestMongoTestSuite(t *testing.T) {
	suite.Run(t, &TestSuite{})
}

func (ts *TestSuite) TestGetTodoByID() {
	got, err := ts.server.store.GetTodoByID("67bc5c4f1e8db0c9a17efca0")
	dueDate1 := time.Now().AddDate(0, 3, 0)

	objID, _ := bson.ObjectIDFromHex("67bc5c4f1e8db0c9a17efca0")
	want := models.TODO{ID: &objID, Name: "Water Plants", Description: "Not too much water for aloe vera", DueDate: &dueDate1}

	if err != nil {
		ts.FailNowf("err on GetTodoByID: ", err.Error())
	}

	ts.compareTodoStructFields(got, want)
}

/*
func (ts *TestSuite) TestGetAllTodos() {
	got, err := ts.server.store.GetAllTodos()
}
*/
