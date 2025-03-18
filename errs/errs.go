package errs

const (
	ErrNotFound       = TodoErr("cannot find todo user that user has specified")
	ErrIdAlreadyInUse = TodoErr("unexpected error: ID is already in use. to prevent unintentional overwrite, we have blocked this request")
)

type TodoErr string

// make TodoErr an error type by implementing method

func (e TodoErr) Error() string {
	return string(e)
}
