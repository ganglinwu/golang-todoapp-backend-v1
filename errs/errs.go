package errs

const (
	ErrNotFound             = TodoErr("cannot find todo user that user has specified")
	ErrIdAlreadyInUse       = TodoErr("unexpected error: ID is already in use. to prevent unintentional overwrite, we have blocked this request")
	ErrEnvVarNotFound       = TodoErr("cannot find environment variable, please check .env file")
	ErrTodoNameEmpty        = TodoErr("the name of the todo/task must not be empty")
	ErrTodoDueDateInThePast = TodoErr("the due date of the todo/task is in the past")
)

type TodoErr string

// make TodoErr an error type by implementing method

func (e TodoErr) Error() string {
	return string(e)
}
