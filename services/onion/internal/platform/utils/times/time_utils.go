package times

import (
	"database/sql"
	"errors"
	"strings"
	"time"
)

var ErrDurationEmpty = errors.New("duration is empty")

const (
	FormatWithDateTimeZone = "2006-01-02 15:04:05 -0700 MST"
	FormatWithDateTime     = "2006-01-02 15:04:05"
	FormatWithDate         = "2006-01-02"
	FormatCardExpiry       = "01/2006"
	FormatDDMMYYYY         = "01/02/2006"
	FormatDDMMYYYYHHMMSS   = "02012006150405"
	FormatFullDate         = "January 02, 2006"
)

const (
	YYYYMMDDHHmmsss = "2006-01-02 15:04:05"
)

const (
	MillisecondsPerSecond = 1000
	SecondsPerMinute      = 60
	SecondsPerDay         = 86400
	MinutesPerHour        = 60
	HoursPerDay           = 24
	DaysPerYear           = 365
	DaysPerMonth          = 30
	DaysPerWeek           = 7
)

const (
	Day = time.Hour * HoursPerDay
)

func GetTimeStringYMDHMS(t time.Time) string {
	return t.Format(YYYYMMDDHHmmsss)
}

func ParseWithDefault(format, t string, defaultTime time.Time) time.Time {
	parsedTime, err := time.Parse(format, t)
	if err != nil {
		return defaultTime
	}

	return parsedTime
}

func durationBetween(start, end time.Time) time.Duration {
	return end.Sub(start)
}

func DaysBetween(start, end time.Time) float64 {
	const hoursInADay = 24

	return durationBetween(start, end).Hours() / hoursInADay
}

func TimeFromUnix(unix int64) time.Time {
	return time.Unix(unix, 0)
}

func UnixFromTime(t time.Time) int64 {
	if t.IsZero() {
		return 0
	}

	return t.Unix()
}

func GetDayRangeStartAndEnd() (start, end time.Time) {
	now := time.Now().UTC()
	start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	end = start.Add(24 * time.Hour).Add(-time.Nanosecond)

	return start, end
}

func GetWeekRangeStartAndEnd() (start, end time.Time) {
	now := time.Now().UTC()

	weekday := now.Weekday()
	if weekday == 0 {
		weekday = 7
	}

	start = time.Date(now.Year(), now.Month(), now.Day()-int(weekday)+1, 0, 0, 0, 0, time.UTC)
	end = start.Add(7 * 24 * time.Hour).Add(-time.Nanosecond)

	return start, end
}

func GetMonthRangeStartAndEnd() (start, end time.Time) {
	now := time.Now().UTC()
	start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	end = start.AddDate(0, 1, 0).Add(-time.Nanosecond)

	return start, end
}

func GetDurationFromHHMMSS(duration string) (time.Duration, error) {
	duration = strings.TrimSpace(duration)
	if duration == "" {
		return time.Duration(0), ErrDurationEmpty
	}

	const secondsFormat = 1

	const minutesFormat = 2

	const hoursFormat = 3

	durationParts := strings.Split(duration, ":")
	switch len(durationParts) {
	case secondsFormat:
		return time.ParseDuration(durationParts[0] + "s")
	case minutesFormat:
		return time.ParseDuration(durationParts[0] + "m" + durationParts[1] + "s")
	case hoursFormat:
		return time.ParseDuration(durationParts[0] + "h" + durationParts[1] + "m" + durationParts[2] + "s")
	default:
		return time.Duration(0), ErrDurationEmpty
	}
}

func getTimezoneByCountry(countryCode string) string {
	countries := map[string]string{
		"AE": "Asia/Dubai",
		"BH": "Asia/Bahrain",
		"DZ": "Africa/Algiers",
		"EG": "Africa/Cairo",
		"IQ": "Asia/Baghdad",
		"JO": "Asia/Amman",
		"KW": "Asia/Kuwait",
		"LB": "Asia/Beirut",
		"LY": "Africa/Tripoli",
		"MA": "Africa/Casablanca",
		"OM": "Asia/Muscat",
		"PS": "Asia/Gaza",
		"QA": "Asia/Qatar",
		"SA": "Asia/Riyadh",
		"SD": "Africa/Khartoum",
		"SS": "Africa/Juba",
		"SY": "Asia/Damascus",
		"TN": "Africa/Tunis",
	}

	timezone, ok := countries[countryCode]
	if !ok {
		return "UTC"
	}

	return timezone
}

func GetCountryTime(country string) (time.Time, error) {
	loc, err := time.LoadLocation(getTimezoneByCountry(country))
	if err != nil {
		return time.Time{}, err
	}

	locationTime := time.Now().In(loc)

	return ParseWithDefault(FormatWithDateTime, locationTime.Format(FormatWithDateTime), locationTime), nil
}

func FromNullTime(t sql.NullTime) *time.Time {
	if !t.Valid {
		return nil
	}

	return &t.Time
}

func FromTimestamp(t sql.NullTime) time.Time {
	if !t.Valid {
		return time.Time{}
	}

	return t.Time
}
