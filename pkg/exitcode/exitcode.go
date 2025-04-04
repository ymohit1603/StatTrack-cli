package exitcode

import "strconv"

const (
	// Success is used when a heartbeat was sent successfully.
	Success = 0
	// ErrGeneric is used for general errors.
	ErrGeneric = 1
	// ErrAPI is when the WakaTime API returned an error.
	ErrAPI = 102
	// ErrAuth is used when the api key is invalid.
	ErrAuth = 104
	// ErrConfigFileParse is used when the ~/.wakatime.cfg config file could not be parsed.
	ErrConfigFileParse = 103
	// ErrConfigFileRead is used for errors of config read command.
	ErrConfigFileRead = 110
	// ErrConfigFileWrite is used for errors of config write command.
	ErrConfigFileWrite = 111
	// ErrBackoff is used when sending heartbeats postponed because we're currently rate limited.
	ErrBackoff = 112
)

// Err represents a type response for exit code errors. A Success response is also wrapped in this type.
type Err struct {
	Code int
}

// Error method to implement error interface.
func (e Err) Error() string {
	return strconv.Itoa(e.Code)
}
