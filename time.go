package goutil

import (
	"strconv"
	"time"
)

var ZeroTime = Time(time.Unix(0, 0))

type Time time.Time

func TimeWithTime(t time.Time) Time {
	return Time(t)
}

func TimeWithUnix(ts int64) Time {
	return Time(time.Unix(ts, 0))
}

func (t Time) Time() time.Time {
	return time.Time(t)
}

func (t Time) String() string {
	return t.Time().Format("2006-01-02 15:04:05")
}

func (t Time) DateString() string {
	return t.Time().Format("2006-01-02")
}

func (t Time) CompactDateString() string {
	return t.Time().Format("20060102")
}

func (t Time) IsZero() bool {
	date := t.DateString()
	return t.Time().IsZero() || date == "0001-01-01" || date == "1970-01-01" || date == "0000-00-00"
}

func (t Time) MarshalJSON() ([]byte, error) {
	return []byte(StringsJoin(`"`, t.String(), `"`)), nil
}

func (t *Time) UnmarshalJSON(b []byte) error {
	if b[0] == '"' && b[len(b)-1] == '"' {
		b = b[1 : len(b)-1]
		val, err := time.ParseInLocation("2006-01-02 15:04:05", string(b), time.Now().Location())
		if err != nil {
			return err
		}
		*t = Time(val)
	} else {
		ts, _ := strconv.ParseInt(string(b), 10, 64)
		*t = Time(time.Unix(ts, 0))
	}
	return nil
}

func TimeToDate(ts time.Time) time.Time {
	loc := ts.Location()
	return time.Date(ts.Year(), ts.Month(), ts.Day(), 0, 0, 0, 0, loc)
}

func TimeToDateEnd(ts time.Time) time.Time {
	loc := time.Now().Location()
	return time.Date(ts.Year(), ts.Month(), ts.Day(), 23, 59, 59, 0, loc)
}

func TimeToHour(ts time.Time) time.Time {
	loc := ts.Location()
	return time.Date(ts.Year(), ts.Month(), ts.Day(), ts.Hour(), 0, 0, 0, loc)
}

func Monday(ts time.Time) time.Time {
	offset := int(time.Monday - ts.Weekday())
	if offset > 0 {
		offset = -6
	}
	return ts.AddDate(0, 0, offset)
}
