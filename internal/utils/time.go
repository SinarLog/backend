package utils

import (
	"fmt"
	"log"
	"math"
	"time"
)

var CURRENT_LOC *time.Location

func init() {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		log.Fatalf("Failed to load location: %s", err)
	}

	CURRENT_LOC = loc
}

func GetStartOfDay() time.Time {
	now := time.Now().In(CURRENT_LOC)

	year, month, day := now.Date()

	return time.Date(year, month, day, 0, 0, 0, 0, now.Location())
}

func GetEndOfDay() time.Time {
	now := time.Now().In(CURRENT_LOC)

	year, month, day := now.Date()

	return time.Date(year, month, day, 23, 59, 59, 0, now.Location())
}

// GetStartOfTheMonth returns the start of the month
// from now. If today is 10th June of 2023, it returns
// 1th June of 2023.
func GetStartOfTheMonth() time.Time {
	year, month, _ := time.Now().In(CURRENT_LOC).Date()

	return time.Date(year, month, 1, 0, 0, 0, 0, CURRENT_LOC)
}

func GetStartOfTheMonthFromMonthAndYear(month, year int) time.Time {
	return time.Date(year, time.Month(month), 1, 0, 0, 0, 0, CURRENT_LOC)
}

func GetEndOfTheMonth() time.Time {
	x := GetStartOfTheMonth()

	return x.AddDate(0, 1, -1)
}

func GetEndOfTheMonthFromMonthAndYear(month, year int) time.Time {
	return GetStartOfTheMonthFromMonthAndYear(month, year).AddDate(0, 1, -1)
}

// CountNumberOfDays counts the number of days
// from `f` time to `t` time.
func CountNumberOfDays(from, to time.Time) int {
	res := math.Ceil(to.Sub(from).Hours() / 24)

	return int(res)
}

// CountNumberOfWorkingDays calls CountNumberOfDays
// and then only count the number of days where it
// is not a weekend. For example given Friday to Monday
// it will return 2 since Saturday and Sunday is weekend.
func CountNumberOfWorkingDays(from, to time.Time) int {
	days := CountNumberOfDays(from, to)

	var res int
	for i := 0; i < days; i++ {
		day := time.Date(from.Year(), from.Month(), from.Day()+i, 0, 0, 0, 0, CURRENT_LOC)
		if day.Weekday() != time.Sunday && day.Weekday() != time.Saturday {
			res += 1
		}
	}

	return res
}

// CountNumberOfDaysFromDuration counts the number of days
// from a given time duration.
func CountNumberOfDaysFromDuration(dur time.Duration) int {
	return int(dur.Hours() / 24)
}

// GetWorkingDay return a time of a day where it is usually
// a working day such as Monday, Tuesday, to Friday. Hence,
// given Saturday, it will return Monday the next week.
func GetWorkingDay(t time.Time) time.Time {
	if t.Weekday() != time.Sunday && t.Weekday() != time.Saturday {
		return t
	}

	var i int
	for {
		t := t.Add(time.Duration(i) * 24 * time.Hour)

		if t.Weekday() != time.Sunday && t.Weekday() != time.Saturday {
			return t
		}

		i++
	}
}

func AddNumOfWorkingDays(t time.Time, days int) time.Time {
	for days != 0 {
		t = t.AddDate(0, 0, 1)
		if t.Weekday() != time.Sunday && t.Weekday() != time.Saturday {
			days--
		}
	}

	return t
}

// GetStartOfTheWeekFromDate return the start of the wwekday
// from the given date. For example, if the given date is
// Tuesday, June 20th 2023, then it will return Monday,
// June 19th 2023 00:00:00.
func GetStartOfTheWeekFromDate(date time.Time) time.Time {
	sterilizedDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, CURRENT_LOC)
	offset := (int(time.Monday) - int(sterilizedDate.Weekday()) - 7) % 7
	result := sterilizedDate.Add(time.Duration(offset*24) * time.Hour)
	return result
}

// GetStartOfTheWeekFromToday return the start of the weekday in
// todays week. For example, if today is Tuesday, June 20th 2023,
// then it will return Monday, June 19th 2023 00:00:00.
func GetStartOfTheWeekFromToday() time.Time {
	now := time.Now().In(CURRENT_LOC)
	date := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, CURRENT_LOC)
	offset := (int(time.Monday) - int(date.Weekday()) - 7) % 7
	result := date.Add(time.Duration(offset*24) * time.Hour)
	return result
}

// GetEndOfWeekdayFromToday returns the end of the weekday in
// todays week. For example, if today is Tuesday, June 20th 2023,
// then it will return Friday, June 23rd 2023 at 23:59:59
func GetEndOfWeekdayFromToday() time.Time {
	now := time.Now().In(CURRENT_LOC)
	date := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, CURRENT_LOC)
	offset := (int(time.Monday) - int(date.Weekday()) - 7) % 7
	result := date.Add(time.Duration(offset*24) * time.Hour).Add(4 * 24 * time.Hour)
	return result
}

// SanitizeDuration returns a string of the format such as
// 'x hours 5 minutes' from a given duration.
func SanitizeDuration(t time.Duration) string {
	var str string

	t = t.Round(time.Minute)
	hour := math.Floor(t.Hours())
	minutes := time.Duration(t - time.Duration(time.Duration(hour)*time.Hour)).Minutes()

	if hour > 1 {
		str += fmt.Sprintf("%.0f hours", hour)
	} else if hour > 0 {
		str += fmt.Sprintf("%0.f hour", hour)
	}

	if minutes > 1 {
		if str != "" {
			str += " "
		}
		str += fmt.Sprintf("%.0f minutes", minutes)
	} else if minutes > 0 {
		if str != "" {
			str += " "
		}
		str += fmt.Sprintf("%.0f mninute", minutes)
	}

	return str
}
