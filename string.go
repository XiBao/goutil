package util

import (
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
