package goutil

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"unsafe"

	"github.com/jackc/numfmt"
	"github.com/sqids/sqids-go"
)

var (
	nativeEndian binary.ByteOrder
	sqidsEnc, _  = sqids.New()
)

func init() {
	buf := [2]byte{}
	*(*uint16)(unsafe.Pointer(&buf[0])) = uint16(0xABCD)

	switch buf {
	case [2]byte{0xCD, 0xAB}:
		nativeEndian = binary.LittleEndian
	case [2]byte{0xAB, 0xCD}:
		nativeEndian = binary.BigEndian
	default:
		panic("Could not determine native endianness.")
	}
}

// Uint64 support string quoted number in json
type Uint64 uint64

// UnmarshalJSON implement json Unmarshal interface
func (u64 *Uint64) UnmarshalJSON(b []byte) (err error) {
	if len(b) == 0 {
		*u64 = 0
	}
	if b[0] == '"' && b[len(b)-1] == '"' {
		b = b[1 : len(b)-1]
	}
	i, _ := strconv.ParseUint(string(b), 10, 64)
	*u64 = Uint64(i)
	return
}

// UnmarshalParam satisfies gin.BindUnmarshaler
func (u64 *Uint64) UnmarshalParam(param string) error {
	return u64.UnmarshalJSON([]byte(param))
}

func (u64 Uint64) Value() uint64 {
	return uint64(u64)
}

func (u64 Uint64) String() string {
	return strconv.FormatUint(uint64(u64), 10)
}

type JSONUint64 uint64

func (u64 JSONUint64) MarshalJSON() ([]byte, error) {
	return []byte(StringsJoin(`"`, u64.String(), `"`)), nil
}

// UnmarshalJSON implement json Unmarshal interface
func (u64 *JSONUint64) UnmarshalJSON(b []byte) (err error) {
	if len(b) == 0 {
		*u64 = 0
	}
	if b[0] == '"' && b[len(b)-1] == '"' {
		b = b[1 : len(b)-1]
	}
	i, _ := strconv.ParseUint(string(b), 10, 64)
	*u64 = JSONUint64(i)
	return
}

func (u64 JSONUint64) Value() uint64 {
	return uint64(u64)
}

func (u64 JSONUint64) String() string {
	return strconv.FormatUint(uint64(u64), 10)
}

// Int64 support string quoted number in json
type Int64 int64

// UnmarshalJSON implement json Unmarshal interface
func (i64 *Int64) UnmarshalJSON(b []byte) (err error) {
	if len(b) == 0 {
		*i64 = 0
	}
	if b[0] == '"' && b[len(b)-1] == '"' {
		b = b[1 : len(b)-1]
	}
	i, _ := strconv.ParseInt(string(b), 10, 64)
	*i64 = Int64(i)
	return
}

// UnmarshalParam satisfies gin.BindUnmarshaler
func (i64 *Int64) UnmarshalParam(param string) error {
	return i64.UnmarshalJSON([]byte(param))
}

func (i64 Int64) Value() int64 {
	return int64(i64)
}

func (i64 Int64) String() string {
	return strconv.FormatInt(int64(i64), 10)
}

type JSONInt64 int64

func (i64 JSONInt64) MarshalJSON() ([]byte, error) {
	return []byte(StringsJoin(`"`, i64.String(), `"`)), nil
}

// UnmarshalJSON implement json Unmarshal interface
func (i64 *JSONInt64) UnmarshalJSON(b []byte) (err error) {
	if len(b) == 0 {
		*i64 = 0
	}
	if b[0] == '"' && b[len(b)-1] == '"' {
		b = b[1 : len(b)-1]
	}
	i, _ := strconv.ParseInt(string(b), 10, 64)
	*i64 = JSONInt64(i)
	return
}

// UnmarshalParam satisfies gin.BindUnmarshaler
func (i64 *JSONInt64) UnmarshalParam(param string) error {
	return i64.UnmarshalJSON([]byte(param))
}

func (i64 JSONInt64) Value() int64 {
	return int64(i64)
}

func (i64 JSONInt64) String() string {
	return strconv.FormatInt(int64(i64), 10)
}

// Int support string quoted number in json
type Int int

// UnmarshalJSON implement json Unmarshal interface
func (i32 *Int) UnmarshalJSON(b []byte) (err error) {
	if len(b) == 0 {
		*i32 = 0
	}
	if b[0] == '"' && b[len(b)-1] == '"' {
		b = b[1 : len(b)-1]
	}
	i, _ := strconv.Atoi(string(b))
	*i32 = Int(i)
	return
}

// UnmarshalParam satisfies gin.BindUnmarshaler
func (i32 *Int) UnmarshalParam(param string) error {
	return i32.UnmarshalJSON([]byte(param))
}

func (i32 Int) Value() int {
	return int(i32)
}

func (i32 Int) String() string {
	return strconv.Itoa(i32.Value())
}

// Float64 support string quoted number in json
type Float64 float64

// UnmarshalJSON implement json Unmarshal interface
func (f64 *Float64) UnmarshalJSON(b []byte) (err error) {
	if len(b) == 0 {
		*f64 = 0
	}
	if b[0] == '"' && b[len(b)-1] == '"' {
		b = b[1 : len(b)-1]
	}
	i, _ := strconv.ParseFloat(string(b), 64)
	*f64 = Float64(i)
	return
}

// UnmarshalParam satisfies gin.BindUnmarshaler
func (f64 *Float64) UnmarshalParam(param string) error {
	return f64.UnmarshalJSON([]byte(param))
}

func (f64 Float64) Value() float64 {
	return float64(f64)
}

func BitStringToUintSlice(s string, size int) []uint64 {
	var (
		b     uint64
		parts []uint64
		idx   = 1
	)
	for _, v := range s {
		if v == 49 {
			b = b | 1<<(size-idx)
		}
		if idx%48 == 0 {
			parts = append(parts, b)
			b = 0
			idx = 1
		} else {
			idx += 1
		}
	}
	return parts
}

func UintToBitString(v uint64, size int) string {
	s := fmt.Sprintf("%b", v)
	if len(s) >= size {
		return s
	}
	var (
		i    int
		bits = make([]string, size)
	)
	for i < size {
		var (
			pos = size - i - 1
			c   = "0"
		)
		if v&(1<<pos) > 0 {
			c = "1"
		}
		bits[i] = c
		i += 1
	}
	return strings.Join(bits, "")
}

func NumberFormatter() *numfmt.Formatter {
	return &numfmt.Formatter{}
}

func RoundNumberFormatter(round int32) *numfmt.Formatter {
	return &numfmt.Formatter{
		Rounder: &numfmt.Rounder{
			Places: round,
		},
	}
}

func RangeRandInt(from int, to int) int {
	if from == to {
		return to
	}
	if from > to {
		from, to = to, from
	}
	return rand.Intn(to+1-from) + from
}

func IntToBytes(i int) []byte {
	return Uint64ToBytes(uint64(i))
}

func Int64ToBytes(i int64) []byte {
	return Uint64ToBytes(uint64(i))
}

func Uint64ToBytes(i uint64) []byte {
	b := make([]byte, 8)
	nativeEndian.PutUint64(b, i)
	return b
}

func Int64FromBytes(b []byte) int64 {
	return int64(Uint64FromBytes(b))
}

func Uint64FromBytes(b []byte) uint64 {
	return nativeEndian.Uint64(b)
}

func IntFromFloat[T ~uint64 | ~int64 | ~int32 | ~uint32 | ~int | ~uint](f64 float64, mul float64) T {
	return T(math.Round(f64 * mul))
}

func IntSliceToString[T ~int32 | ~int64 | ~int](arr []T, spliter string) string {
	l := len(arr)
	if l == 0 {
		return ""
	}
	builder := GetStringsBuilder()
	defer PutStringsBuilder(builder)
	for idx, v := range arr {
		builder.WriteString(strconv.FormatInt(int64(v), 10))
		if idx < l-1 {
			builder.WriteString(spliter)
		}
	}
	return builder.String()
}

func IntSliceFromString[T ~int32 | ~int64 | ~int](str string, spliter string) []T {
	if str == "" {
		return nil
	}
	arr := strings.Split(str, spliter)
	ret := make([]T, 0, len(arr))
	for _, s := range arr {
		if v, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64); err == nil {
			ret = append(ret, T(v))
		}
	}
	return ret
}

func UintSliceToString[T ~uint32 | ~uint64 | ~uint](arr []T, spliter string) string {
	l := len(arr)
	if l == 0 {
		return ""
	}
	builder := GetStringsBuilder()
	defer PutStringsBuilder(builder)
	for idx, v := range arr {
		builder.WriteString(strconv.FormatUint(uint64(v), 10))
		if idx < l-1 {
			builder.WriteString(spliter)
		}
	}
	return builder.String()
}

func UintSliceFromString[T ~uint32 | ~uint64 | ~uint](str string, spliter string) []T {
	if str == "" {
		return nil
	}
	arr := strings.Split(str, spliter)
	ret := make([]T, 0, len(arr))
	for _, s := range arr {
		if v, err := strconv.ParseUint(strings.TrimSpace(s), 10, 64); err == nil {
			ret = append(ret, T(v))
		}
	}
	return ret
}

func NumbersToUint64s[T ~int | ~int32 | ~uint | ~uint32](arr []T) []uint64 {
	ret := make([]uint64, 0, len(arr))
	for _, v := range arr {
		ret = append(ret, uint64(v))
	}
	return ret
}

func EncodeUint64s(ids ...uint64) string {
	ret, _ := sqidsEnc.Encode(ids)
	return ret
}

func DecodeUint64s(req string) []uint64 {
	return sqidsEnc.Decode(req)
}
