package base62

// Error is a token error
type Error string

// Error implements the `errors.Error` interface
func (e Error) Error() string {
	return string(e)
}

const (
	// ErrInvalidCharacter is the error returned or panic'd when a non `Base62` string is being parsed
	ErrInvalidCharacter = Error("there was a non base62 character in the token")
)
