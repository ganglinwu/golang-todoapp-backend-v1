package errs

const (
	ErrNotFound = TodoErr("cannot find todo user that user has specified")
)

type TodoErr string

// make TodoErr an error type by implementing method

func (e TodoErr) Error() string {
	return string(e)
}
