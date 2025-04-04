package handlers

type ValidationError struct {
	msg string
}

func (err ValidationError) Error() string {
	return err.msg
}
