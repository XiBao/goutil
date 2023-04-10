package base62

import (
	"bytes"
	"math"
)

const (
	// Base62 is a string respresentation of every possible base62 character
	Base62 = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

var (
	base62Len = int64(len(Base62))
)

type Encoder uint64

// Encode encodes the token into a base62 string
func (t Encoder) Encode() string {
	bs, _ := t.MarshalText()
	return string(bs)
}

// Decode returns a token from a 1-12 character base62 encoded string
func Decode(token string) (Encoder, error) {
	var t Encoder
	err := (&t).UnmarshalText([]byte(token))
	return t, err
}

// UnmarshalText implements the `encoding.TextUnmarshaler` interface
func (t *Encoder) UnmarshalText(data []byte) error {

	number := uint64(0)
	idx := 0.0
	chars := []byte(Base62)

	charsLength := float64(len(chars))
	tokenLength := float64(len(data))

	for _, c := range data {
		power := tokenLength - (idx + 1)
		index := bytes.IndexByte(chars, c)
		if index < 0 {
			return ErrInvalidCharacter
		}
		number += uint64(index) * uint64(math.Pow(charsLength, power))
		idx++
	}

	// the token was successfully decoded
	*t = Encoder(number)
	return nil
}

// MarshalText implements the `encoding.TextMarsheler` interface
func (t Encoder) MarshalText() ([]byte, error) {
	number := int64(t)
	var chars []byte

	if number == 0 {
		return chars, nil
	}

	for number > 0 {
		result := number / base62Len
		remainder := number % base62Len
		chars = append(chars, Base62[remainder])
		number = result
	}

	for i, j := 0, len(chars)-1; i < j; i, j = i+1, j-1 {
		chars[i], chars[j] = chars[j], chars[i]
	}

	return chars, nil
}
