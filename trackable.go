package trackable

import (
	"fmt"
)

// trackable represents an error that uses an integer as an identifier.
//
// If an ID is less than or equal to zero then the error is categorised as
// untracked according to errors.Is.
type trackable struct {
	id    int
	msg   string
	cause error
}

// Track returns a new trackable error, that is, one with a tracking ID.
//
// This function is designed to be called during package initialisation only.
// This means it should only be used to initialise package global variables,
// within init functions, or as part of a test.
func Track(msg string, args ...any) *trackable {
	return &trackable{
		id:  newId(),
		msg: fmt.Sprintf(msg, args...),
	}
}

// Untracked returns a new error without a tracking ID.
func Untracked(msg string, args ...any) *trackable {
	return &trackable{
		msg: fmt.Sprintf(msg, args...),
	}
}

// Wrap returns a new error, without a tracking ID, that wraps a cause.
//
// It's an alternative to fmt.Errorf where the cause does not have to form part
// of the error message.
func Wrap(cause error, msg string, args ...any) *trackable {
	return &trackable{
		msg:   fmt.Sprintf(msg, args...),
		cause: cause,
	}
}

func (e *trackable) Error() string {
	return ErrorStack(e.msg, e.cause)
}

func (e trackable) Unwrap() error {
	return e.cause
}

func (e trackable) Is(target error) bool {
	if e.id <= 0 {
		return false
	}

	if it, ok := target.(*trackable); ok {
		return e.id == it.id
	}

	return false
}

func (e trackable) Wrap(cause error) error {
	e.cause = cause
	return &e
}

func (e trackable) Because(msg string, args ...any) error {
	e.cause = Untracked(msg, args...)
	return &e
}

func (e trackable) BecauseOf(cause error, msg string, args ...any) error {
	e.cause = Wrap(cause, msg, args...)
	return &e
}
