// package errors provide errors
package errors

import (
	"fmt" // for Sprintf()
)

// Error is interface error
type Error interface {
	Error() string
	Name() string
}

// CustomError is custom error
type CustomError struct{ msg string }

// ErrValueNotAvaiable is error about trying to use not available value
type ErrValueNotAvailable struct {
	key any
}

type ErrNilBucket struct{}

// ErrValueDelete is error about trying to use deleting value
type ErrValueDelete struct {
	key any
}

// ErrOutOfRange is error about out of range
type ErrOutOfRange struct {
	index any
}

// New functions creating error

func ToError(err error) Error {
	if err == nil {
		return nil
	}
	return NewErrorf(err.Error())
}

// NewErrorf create CustomError
func NewErrorf(format string, values ...any) Error {
	return NewCustomError(fmt.Sprintf(format, values...))
}

// NewCustomError create CustomError
func NewCustomError(msg string) CustomError {
	return CustomError{msg}
}

// NewErrValueDelete create ErrValueDelete
func NewErrValueDelete(key any) ErrValueDelete {
	return ErrValueDelete{key}
}

// NewErrValueNotAvaiable create ErrValueNotAvaiable
func NewErrValueNotAvailable(key any) ErrValueNotAvailable {
	return ErrValueNotAvailable{key}
}

// NewErrNilBucket create ErrNilBucket
func NewErrNilBucket() ErrNilBucket {
	return ErrNilBucket{}
}

// NewErrOutOfRange create ErrOutOfRange
func NewErrOutOfRange(index uint) ErrOutOfRange {
	return ErrOutOfRange{index}
}

// Name functions return error's names

// Name return "CustomError"
func (err CustomError) Name() string {
	return "CustomError"
}

// Name return "ErrValueDelete"
func (err ErrValueDelete) Name() string {
	return "ErrValueDelete"
}

// Name return "ErrValueNotAvaiable"
func (err ErrValueNotAvailable) Name() string {
	return "ErrValueNotAvailable"
}

// Name return "ErrNilBucket"
func (err ErrNilBucket) Name() string {
	return "ErrNilBucket"
}

// Name return "ErrIndexOutOfRange"
func (err ErrOutOfRange) Name() string {
	return "ErrIndexOutOfRange"
}

// Error functions return string error

// Error return string error
func (err CustomError) Error() string {
	return err.msg
}

// Error return string error
func (err ErrValueDelete) Error() string {
	return fmt.Sprintf("value of key `%v` is delete", err.key)
}

// Error return string error
func (err ErrValueNotAvailable) Error() string {
	return fmt.Sprintf("key `%v` is not available", err.key)
}

// Error return string error
func (err ErrNilBucket) Error() string {
	return "bucket is nil"
}

// Error return string error
func (err ErrOutOfRange) Error() string {
	return fmt.Sprintf("index `%v` does not exists", err.index)
}
