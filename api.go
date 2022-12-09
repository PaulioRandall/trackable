// Package trackerr aims to facilitate creation of referenceable errors and
// elegant stack traces.
package trackerr

import (
	"errors"
	"fmt"
	"strings"
)

// TODO: Realm.ErrorFormatter accepts `func(e error, parent error) string`
// TODO: which formats an error into a string when debug printing. Each
// TODO: error string will be printed on a line of its own so implementations
// TODO: should not prefix or suffix a linefeed unless they want gappy print
// TODO: outs.
// TODO:
// TODO: Will probably need an `IsThirdParty` func that returns true if the
// TODO: error does not implement the UntrackedError or TrackedError, i.e. it's
// TODO: not a go-trackerr error (but the reverse is not guaranteed to be true).

// TODO: Think about how to integrate file names and line numbers.
// TODO: - How, where, and when to collect them?
// TODO: - How to optimise print outs with them?
// TODO: - May have to redesign the Debug function?

var (
	globalRealm IntRealm

	// ErrTodo is a convenience tracked error for specifying a TODO.
	//
	// This can be useful if you're taking a stepwise refinement or test driven
	// approach to writing code.
	ErrTodo = Track("TODO: Implementation needed")

	// ErrBug is a convenience tracked error for use at the site of known bugs.
	ErrBug = Track("BUG: Fix needed")

	// ErrInsane is a convenience tracked error for sanity checking.
	ErrInsane = Track("Sanity check!!")
)

// Untracked returns a new error without a tracking ID.
//
// It is no different than using errors.New except it has a handy fmt.Sprintf
// signiture and a few extra receiving functions for any niche use cases one
// may encounter.
func Untracked(msg string, args ...any) *untrackedError {
	return globalRealm.Untracked(msg, args...)
}

// Wrap returns a new untracked error that wraps a cause.
func Wrap(cause error, msg string, args ...any) *untrackedError {
	e := globalRealm.Untracked(msg, args...)
	return e.Wrap(cause).(*untrackedError)
}

// Track returns a new tracked error from this package's singleton Realm.
//
// This is recommended way to use to create all trackable errors outside of
// testing.
func Track(msg string, args ...any) *trackedError {
	return globalRealm.Track(msg, args...)
}

// Checkpoint returns a new trackable checkpoint error from this package's
// singleton Realm.
//
// This is recommended way to use to create all checkpoint errors outside of
// testing.
func Checkpoint(msg string, args ...any) *trackedError {
	return globalRealm.Checkpoint(msg, args...)
}

// Debug pretty prints the error stack trace to terminal for debugging
// purposes.
//
// If e is nil then a message will be printed indicating so. This function is
// not designed for logging, just day to day manual debugging.
func Debug(e error) (int, error) {
	s := ErrorStack(e)

	if s == "" {
		return fmt.Print("[Debugging error] nil error")
	}

	return fmt.Print("[Debugging error]\n", s)
}

// DebugPanic recovers from a panic, prints out the error using the Debug
// function, and finally sets it as the catch error's pointer value.
//
// If nil is passed as the catch then the panic continues after printing.
//
// If the panic value is not an error the panic will continue!
//
// This function is not designed for logging, just day to day manual debugging.
func DebugPanic(catch *error) {
	v := recover()

	if v == nil {
		return
	}

	e, ok := v.(error)
	if !ok {
		panic(v)
	}

	Debug(e)

	if catch == nil {
		panic(e)
	}
	*catch = e
}

// HasTracked returns true if the error or one of the underlying causes are
// tracked, i.e. those created via the Error and Checkpoint functions.
func HasTracked(e error) bool {

	type wrapper interface {
		Unwrap() error
	}

	for e != nil {
		if IsTracked(e) {
			return true
		}

		if w, ok := e.(wrapper); ok {
			e = w.Unwrap()
		} else {
			e = nil
		}
	}

	return false
}

// IsTracked returns true if the error is being tracked, i.e. those created via
// the Error and Checkpoint functions.
func IsTracked(e error) bool {
	_, ok := e.(*trackedError)
	return ok
}

// IsCheckpoint returns true if the error is a trackable checkpoint.
func IsCheckpoint(e error) bool {
	if te, ok := e.(*trackedError); ok {
		return te.IsCheckpoint()
	}
	return false
}

// Is is an alias for errors.Is.
func Is(e, target error) bool {
	return errors.Is(e, target)
}

// All returns true only if errors.Is returns true for all targets.
func All(e error, targets ...error) bool {
	for _, t := range targets {
		if !errors.Is(e, t) {
			return false
		}
	}
	return true
}

// Any returns true if errors.Is returns true for any target.
func Any(e error, targets ...error) bool {
	for _, t := range targets {
		if errors.Is(e, t) {
			return true
		}
	}
	return false
}

// ErrorStack returns a human readable stack trace for the error.
func ErrorStack(e error) string {
	sb := strings.Builder{}

	for i, cause := range AsStack(e) {
		errStr := ErrorWithoutCause(cause)

		if IsCheckpoint(cause) {
			if i == 0 {
				sb.WriteString("——")
			} else {
				sb.WriteString("\n——")
			}

			sb.WriteString(errStr)
			sb.WriteString("——")
			continue
		}

		if i == 0 {
			sb.WriteString("  ")
		} else {
			sb.WriteString("\n⤷ ")
		}
		sb.WriteString(errStr)
	}

	sb.WriteString("\n")
	return sb.String()
}

// AsStack recursively unwraps the error returning a slice of errors.
//
// The passed error e will be first and root cause last.
func AsStack(e error) []error {
	var stack []error

	for e != nil {
		stack = append(stack, e)
		e = errors.Unwrap(e)
	}

	return stack
}

// ErrorWithoutCause removes the cause from error messages that use the format
// '%s: %w'. Where s is the error message and w is the cause's message.
func ErrorWithoutCause(e error) string {
	s := e.Error()

	if _, ok := e.(*untrackedError); ok {
		return s
	}

	if _, ok := e.(*trackedError); ok {
		return s
	}

	cause := errors.Unwrap(e)

	if cause == nil {
		return s
	}

	s = strings.TrimSuffix(s, cause.Error())
	s = strings.TrimSpace(s)
	return strings.TrimSuffix(s, ":")
}
