package util

import "time"

func TimeToDate(ts time.Time) time.Time {
	loc := ts.Location()
	return time.Date(ts.Year(), ts.Month(), ts.Day(), 0, 0, 0, 0, loc)
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
