package postgres_store

import (
	"database/sql"
	"os"
	"strconv"

	"github.com/joho/godotenv"

	"github.com/ganglinwu/todoapp-backend-v1/errs"
	"github.com/ganglinwu/todoapp-backend-v1/models"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostGresStore struct {
	DB *sql.DB
}

func NewConnection(connString string) (*sql.DB, error) {
	err := godotenv.Load(".env")
	if err != nil {
		return nil, err
	}

	if connString == "" {
		connStringEnv, exist := os.LookupEnv("POSTGRES_CONNECTION_STRING")
		if !exist {
			return nil, errs.ErrEnvVarNotFound
		}
		connString = connStringEnv
	}

	db, err := sql.Open("pgx", connString)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (pg *PostGresStore) GetAllProjs() ([]models.PROJECT, error) {
	projects := &[]models.PROJECT{}

	stmt := "select * from projects"

	rows, err := pg.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		project := models.PROJECT{}

		err := rows.Scan(&project.Id, &project.ProjName)
		if err != nil {
			return nil, err
		}
		*projects = append(*projects, project)
	}

	return *projects, nil
}

func (pg *PostGresStore) GetAllTodos() ([]models.TODO, error) {
	todos := []models.TODO{}

	stmt := "select * from todos"

	rows, err := pg.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		todo := models.TODO{}
		err := rows.Scan(&todo.Id, &todo.Name, &todo.Description, &todo.DueDate, &todo.Priority, &todo.Completed, &todo.Updated_At, &todo.ProjName)
		if err != nil {
			return nil, err
		}
		todos = append(todos, todo)
	}
	return todos, nil
}

func (pg *PostGresStore) GetProjByID(ID string) (models.PROJECT, error) {
	project := models.PROJECT{}

	IDint, err := strconv.Atoi(ID)
	if err != nil {
		return models.PROJECT{}, err
	}

	stmt := "select * from projects where id = $1"

	row := pg.DB.QueryRow(stmt, IDint)

	err = row.Scan(&project.Id, &project.ProjName)
	if err != nil {
		return models.PROJECT{}, err
	}
	return project, nil
}

func (pg *PostGresStore) GetTodoByID(todoID string) (models.TODO, error) {
	todo := models.TODO{}

	intID, err := strconv.Atoi(todoID)
	if err != nil {
		return models.TODO{}, err
	}

	stmt := `SELECT * FROM todos WHERE id=$1`

	err = pg.DB.QueryRow(stmt, intID).Scan(&todo.Id, &todo.Name, &todo.Description, &todo.DueDate, &todo.Priority, &todo.Completed, &todo.Updated_At, &todo.ProjName)
	if err != nil {
		return models.TODO{}, err
	}
	return todo, nil
}

func (pg *PostGresStore) CreateProj(Name string, Tasks []models.TODO) (string, error) {
	stmt := `insert into projects (projname) values ($1) returning id;`

	var id int

	err := pg.DB.QueryRow(stmt, Name).Scan(&id)
	if err != nil {
		return "", err
	}

	stringID := strconv.Itoa(id)

	return stringID, nil
}

func (pg *PostGresStore) CreateTodo(projID string, newTodoWithoutID models.TODO) (string, error) {
	// first run a query to get the projname from the projID
	intProjID, err := strconv.Atoi(projID)
	if err != nil {
		return "", err
	}

	projName := ""

	row := pg.DB.QueryRow(`select projname from projects where id = $1`, intProjID)

	err = row.Scan(&projName)
	if err != nil {
		return "", err
	}

	// server method handleCreateTodo needs to handle empty inputs!
	stmt := `INSERT INTO todos (name, description, duedate, priority, completed, projname) VALUES($1, $2, $3, $4, $5, $6) RETURNING id;`

	row = pg.DB.QueryRow(stmt, newTodoWithoutID.Name, newTodoWithoutID.Description, newTodoWithoutID.DueDate, newTodoWithoutID.Priority, newTodoWithoutID.Completed, projName)

	var insertedID int

	err = row.Scan(&insertedID)
	if err != nil {
		return "", err
	}
	stringID := strconv.Itoa(insertedID)
	return stringID, nil
}

func (pg *PostGresStore) UpdateProjNameByID(ID, newName string) error {
	stmt := `UPDATE projects SET projname = $1 WHERE id = $2;`

	intID, err := strconv.Atoi(ID)
	if err != nil {
		return err
	}

	_, err = pg.DB.Exec(stmt, newName, intID)
	if err != nil {
		return err
	}

	return nil
}

func (pg *PostGresStore) UpdateTodoByID(todoID string, newTodoWithoutID models.TODO) error {
	stmt := `UPDATE todos SET name = $1, description = $2, duedate = $3, priority = $4, completed = $5, projname = $6 WHERE id = $7`

	_, err := pg.DB.Exec(stmt, newTodoWithoutID.Name, newTodoWithoutID.Description, newTodoWithoutID.DueDate, newTodoWithoutID.Priority, newTodoWithoutID.Completed, newTodoWithoutID.ProjName, todoID)
	if err != nil {
		return err
	}

	return nil
}

func (pg *PostGresStore) DeleteProjByID(projID string) (int, error) {
	stmt := `DELETE FROM projects WHERE id = $1`

	intProjID, err := strconv.Atoi(projID)
	if err != nil {
		return 0, err
	}

	result, err := pg.DB.Exec(stmt, intProjID)
	if err != nil {
		return 0, err
	}

	deletedCount, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return int(deletedCount), nil
}

func (pg *PostGresStore) DeleteTodoByID(todoID string) (int, error) {
	stmt := `DELETE FROM todos WHERE id = $1`

	intTodoID, err := strconv.Atoi(todoID)
	if err != nil {
		return 0, err
	}

	result, err := pg.DB.Exec(stmt, intTodoID)
	if err != nil {
		return 0, nil
	}

	deleteCount, err := result.RowsAffected()
	if err != nil {
		return 0, nil
	}

	return int(deleteCount), nil
}

/*
* methods to implement

	type TodoStore interface {
		GetAllProjs() ([]models.PROJECT, error)
		GetAllTodos() ([]models.TODO, error)
		GetProjByID(ID string) (models.PROJECT, error)
		CreateProj(Name string, Tasks []models.TODO) (string, error)
		CreateTodo(projID string, newTodoWithoutID models.TODO) (interface{}, error)
		UpdateProjNameByID(ID, newName string) error
		UpdateTodoByID(todoID string, newTodoWithoutID models.TODO) error
		DeleteProjByID(ID string) (int, error)
		DeleteTodoByID(todoID string) (int, error)
		GetTodoByID(todoID string) (models.TODO, error)
	}
*/
