package api

import (
	"fmt"

	"github.com/wakatime/wakatime-cli/pkg/exitcode"
	"github.com/wakatime/wakatime-cli/pkg/wakaerror"

	"go.uber.org/zap/zapcore"
)

// Err represents a general api error.
type Err struct {
	Err error
}

var _ wakaerror.Error = Err{}

// Error method to implement error interface.
func (e Err) Error() string {
	return e.Err.Error()
}

// ExitCode method to implement wakaerror.Error interface.
func (Err) ExitCode() int {
	return exitcode.ErrAPI
}

// Message method to implement wakaerror.Error interface.
func (e Err) Message() string {
	return fmt.Sprintf("api error: %s", e.Err)
}

// SendDiagsOnErrors method to implement wakaerror.SendDiagsOnErrors interface.
func (Err) SendDiagsOnErrors() bool {
	return false
}

// ShouldLogError method to implement wakaerror.ShouldLogError interface.
func (Err) ShouldLogError() bool {
	return true
}

// ErrAuth represents an authentication error.
type ErrAuth struct {
	Err error
}

var _ wakaerror.Error = ErrAuth{}

// Error method to implement error interface.
func (e ErrAuth) Error() string {
	return e.Err.Error()
}

// ExitCode method to implement wakaerror.Error interface.
func (ErrAuth) ExitCode() int {
	return exitcode.ErrAuth
}

// Message method to implement wakaerror.Error interface.
func (e ErrAuth) Message() string {
	return fmt.Sprintf("invalid api key... find yours at wakatime.com/api-key. %s", e.Err.Error())
}

// SendDiagsOnErrors method to implement wakaerror.SendDiagsOnErrors interface.
func (ErrAuth) SendDiagsOnErrors() bool {
	return false
}

// ShouldLogError method to implement wakaerror.ShouldLogError interface.
func (ErrAuth) ShouldLogError() bool {
	return true
}

// ErrBadRequest represents a 400 response from the API.
type ErrBadRequest struct {
	Err error
}

var _ wakaerror.Error = ErrBadRequest{}

// Error method to implement error interface.
func (e ErrBadRequest) Error() string {
	return e.Err.Error()
}

// ExitCode method to implement wakaerror.Error interface.
func (ErrBadRequest) ExitCode() int {
	return exitcode.ErrGeneric
}

// Message method to implement wakaerror.Error interface.
func (e ErrBadRequest) Message() string {
	return fmt.Sprintf("bad request: %s", e.Err)
}

// SendDiagsOnErrors method to implement wakaerror.SendDiagsOnErrors interface.
func (ErrBadRequest) SendDiagsOnErrors() bool {
	return false
}

// ShouldLogError method to implement wakaerror.ShouldLogError interface.
func (ErrBadRequest) ShouldLogError() bool {
	return true
}

// ErrBackoff means we send later because currently rate limited.
type ErrBackoff struct {
	Err error
}

var _ wakaerror.Error = ErrBackoff{}

// Error method to implement error interface.
func (e ErrBackoff) Error() string {
	return e.Err.Error()
}

// ExitCode method to implement wakaerror.Error interface.
func (ErrBackoff) ExitCode() int {
	return exitcode.ErrBackoff
}

// LogLevel method to implement wakaerror.LogLevel interface.
func (ErrBackoff) LogLevel() int8 {
	return int8(zapcore.DebugLevel)
}

// Message method to implement wakaerror.Error interface.
func (e ErrBackoff) Message() string {
	return fmt.Sprintf("rate limited: %s", e.Err)
}

// SendDiagsOnErrors method to implement wakaerror.SendDiagsOnErrors interface.
func (ErrBackoff) SendDiagsOnErrors() bool {
	return false
}

// ShouldLogError method to implement wakaerror.ShouldLogError interface.
func (ErrBackoff) ShouldLogError() bool {
	return false
}

// ErrTimeout represents a timeout error.
type ErrTimeout struct {
	Err error
}

var _ wakaerror.Error = ErrTimeout{}

// Error method to implement error interface.
func (e ErrTimeout) Error() string {
	return e.Err.Error()
}

// ExitCode method to implement wakaerror.Error interface.
func (ErrTimeout) ExitCode() int {
	return exitcode.ErrGeneric
}

// LogLevel method to implement wakaerror.LogLevel interface.
func (ErrTimeout) LogLevel() int8 {
	return int8(zapcore.DebugLevel)
}

// Message method to implement wakaerror.Error interface.
func (e ErrTimeout) Message() string {
	return fmt.Sprintf("timeout: %s", e.Err)
}

// SendDiagsOnErrors method to implement wakaerror.SendDiagsOnErrors interface.
func (ErrTimeout) SendDiagsOnErrors() bool {
	return false
}

// ShouldLogError method to implement wakaerror.ShouldLogError interface.
func (ErrTimeout) ShouldLogError() bool {
	return false
}
