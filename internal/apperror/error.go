package apperror

type AppError struct {
	Code int
	Msg  string
	Err  error
}

func New(code int, msg string) *AppError {
	return &AppError{
		Code: code,
		Msg:  msg,
	}
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}

	return e.Msg
}

func Wrap(code int, msg string, err error) *AppError {
	return &AppError{Code: code, Msg: msg, Err: err}
}

func (e *AppError) Unwrap() error {
	return e.Err
}
