package goutil

import (
	"errors"
	"math/big"
	"math/rand"
	"strings"
	"time"

	"github.com/cespare/xxhash/v2"

	"github.com/ziutek/mymysql/autorc"
)

func ReplaceRuneAtIndex(str []rune, replacement rune, index int) []rune {
	str[index] = replacement
	return str
}

func ReplaceStringAtIndex(str string, replacement string, index int) string {
	return StringsJoin(str[:index], replacement, str[index+1:])
}

func NormalizeSellerNick(nick string) string {
	return strings.ReplaceAll(strings.ToLower(nick), " ", "")
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var randSeed = rand.NewSource(time.Now().UnixNano())

func RandStringN(n int) string {
	if n <= 0 {
		return ""
	}
	sb := GetStringsBuilder()
	defer PutStringsBuilder(sb)
	sb.Grow(n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, randSeed.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = randSeed.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			sb.WriteByte(letterBytes[idx])
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return sb.String()
}

func StringsJoin(strs ...string) string {
	var n int
	for i := 0; i < len(strs); i++ {
		n += len(strs[i])
	}
	if n <= 0 {
		return ""
	}
	builder := GetStringsBuilder()
	defer PutStringsBuilder(builder)
	builder.Grow(n)
	for _, s := range strs {
		builder.WriteString(s)
	}
	return builder.String()
}

func DBQuote(str string, db *autorc.Conn) string {
	if db == nil {
		return StringsJoin("'", str, "'")
	}
	return StringsJoin("'", db.Escape(str), "'")
}

func StringToUint64(str string) uint64 {
	return xxhash.Sum64([]byte(str))
}

func BitStringAnd(str1 string, str2 string) (string, error) {
	return bitStringOp(str1, str2, '&')
}

func BitStringOr(str1 string, str2 string) (string, error) {
	return bitStringOp(str1, str2, '|')
}

func BitStringXor(str1 string, str2 string) (string, error) {
	return bitStringOp(str1, str2, '^')
}

func bitStringOp(str1 string, str2 string, op byte) (string, error) {
	var (
		n1 big.Int
		n2 big.Int
	)
	l := len(str1)
	if l2 := len(str2); l2 > l {
		l = l2
	}
	if _, ok := n1.SetString(str1, 2); !ok {
		return "", errors.New("invalid str1")
	}
	if _, ok := n2.SetString(str2, 2); !ok {
		return "", errors.New("invalid str2")
	}
	return bitsOp(&n1, &n2, op, l), nil
}

func bitsOp(n1 *big.Int, n2 *big.Int, op byte, l int) string {
	var ret big.Int
	switch op {
	case '^':
		ret.Xor(n1, n2)
	case '|':
		ret.Or(n1, n2)
	case '&':
		ret.And(n1, n2)
	}
	buf := GetBufferPool()
	defer PutBufferPool(buf)
	append := l - ret.BitLen()
	if append > 0 {
		for i := 0; i < append; i++ {
			buf.WriteByte('0')
		}
	}
	for i := ret.BitLen() - 1; i >= 0; i-- {
		b := ret.Bit(i)
		if b == 1 {
			buf.WriteByte('1')
		} else {
			buf.WriteByte('0')
		}
	}
	str := buf.String()
	return str
}
