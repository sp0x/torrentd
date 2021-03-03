package utils

import (
	"fmt"
	"github.com/bcampbell/fuzzytime"
	"regexp"
	"strconv"
	"strings"
	"text/scanner"
	"time"
	"unicode"
)

const (
	filterTimeFormat = time.RFC1123Z
)

func NormalizeNumber(s string) string {
	normalized := strings.ReplaceAll(s, ",", "")

	if normalized == "" {
		normalized = "0"
	}

	return normalized
}

func NormalizeSpace(s string) string {
	return strings.TrimSpace(strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return ' '
		}
		return r
	}, s))
}

func splitDecimalStr(s string) (int, float64, error) {
	if parts := strings.SplitN(s, ".", 2); len(parts) == 2 {
		i, err := strconv.Atoi(parts[0])
		if err != nil {
			return 0, 0, err
		}
		f, err := strconv.ParseFloat("0."+parts[1], 64)
		if err != nil {
			return 0, 0, err
		}
		return i, f, nil
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0, 0, err
	}
	return i, 0, nil
}

var (
	timeAgoRegexp     = regexp.MustCompile(`(?i)\bago`)
	todayRegexp       = regexp.MustCompile(`(?i)\btoday([\s,]+|$)`)
	tomorrowRegexp    = regexp.MustCompile(`(?i)\btomorrow([\s,]+|$)`)
	yesterdayRegexp   = regexp.MustCompile(`(?i)\byesterday([\s,]+|$)`)
	missingYearRegexp = regexp.MustCompile(`^\d{1,2}-\d{1,2}\b`)
)

func parseTimeAgo(src string, now time.Time) (time.Time, error) {
	normalized := NormalizeSpace(src)
	normalized = strings.ToLower(normalized)

	var s scanner.Scanner
	s.Init(strings.NewReader(normalized))
	var tok rune
	for tok != scanner.EOF {
		tok = s.Scan()

		switch s.TokenText() {
		case ",", "ago", "", "and":
			continue
		}

		v, fraction, err := splitDecimalStr(s.TokenText())
		if err != nil {
			return now, fmt.Errorf(
				"failed to parse decimal time %q in time format at %s", s.TokenText(), s.Pos())
		}

		tok = s.Scan()
		if tok == scanner.EOF {
			return now, fmt.Errorf(
				"expected a time unit at %s:%v", s.TokenText(), s.Pos())
		}

		unit := s.TokenText()
		if unit != "s" {
			unit = strings.TrimSuffix(s.TokenText(), "s")
		}

		switch unit {
		case "year", "yr", "y":
			now = now.AddDate(-v, 0, 0)
			if fraction > 0 {
				now = now.Add(time.Duration(float64(now.AddDate(-1, 0, 0).Sub(now)) * fraction))
			}
		case "month", "mnth", "mo":
			now = now.AddDate(0, -v, 0)
			if fraction > 0 {
				now = now.Add(time.Duration(float64(now.AddDate(0, -1, 0).Sub(now)) * fraction))
			}
		case "week", "wk", "w":
			now = now.AddDate(0, 0, -7)
			if fraction > 0 {
				now = now.Add(time.Duration(float64(now.AddDate(0, 0, -7).Sub(now)) * fraction))
			}
		case "day", "d":
			now = now.AddDate(0, 0, -v)
			if fraction > 0 {
				now = now.Add(time.Minute * -time.Duration(fraction*1440))
			}
		case "hour", "hr", "h":
			now = now.Add(time.Hour * -time.Duration(v))
			if fraction > 0 {
				now = now.Add(time.Second * -time.Duration(fraction*3600))
			}
		case "minute", "min", "m":
			now = now.Add(time.Minute * -time.Duration(v))
			if fraction > 0 {
				now = now.Add(time.Second * -time.Duration(fraction*60))
			}
		case "second", "sec", "s":
			now = now.Add(time.Second * -time.Duration(v))
		default:
			return now, fmt.Errorf("unsupporting unit of time %q", unit)
		}
	}

	return now, nil
}

var todayTimeFormat = "15:04:05"

func ParseFuzzyTime(src string, now time.Time, allowPartialDate bool) (time.Time, error) {
	if okTime, err := time.Parse(todayTimeFormat, src); err == nil {
		dt := fuzzytime.DateTime{}
		dt.Time.SetHour(okTime.Hour())
		dt.Time.SetSecond(okTime.Second())
		dt.Time.SetMinute(okTime.Minute())
		dt.Date.SetYear(now.Year())
		dt.Date.SetDay(now.Day())
		dt.Date.SetMonth(int(now.Month()))
		isof := dt.ISOFormat()
		okTime, _ = time.Parse("2006-01-02T15:04:05", isof)
		return okTime, nil
	}

	if timeAgoRegexp.MatchString(src) {
		t, err := parseTimeAgo(src, now)
		if err != nil {
			return t, fmt.Errorf("error parsing time ago %q: %v", src, err)
		}
		return t, nil
	}

	normalized := NormalizeSpace(src)
	out := todayRegexp.ReplaceAllLiteralString(normalized, now.Format("2006 "))
	out = tomorrowRegexp.ReplaceAllLiteralString(out, now.AddDate(0, 0, 1).Format("Mon, 02 Jan 2006 "))
	out = yesterdayRegexp.ReplaceAllLiteralString(out, now.AddDate(0, 0, -1).Format("Mon, 02 Jan 2006 "))

	if m := missingYearRegexp.FindStringSubmatch(out); len(m) > 0 {
		out = missingYearRegexp.ReplaceAllLiteralString(src, m[0]+now.Format("-2006"))
	}

	dt, _, err := fuzzytime.USContext.Extract(out)
	if err != nil {
		return time.Time{}, fmt.Errorf("error extracting date from %q: %v", out, err)
	}

	if dt.Time.Empty() {
		dt.Time.SetHour(0)
		dt.Time.SetMinute(0)
		dtx, err := time.Parse("2006", out)
		if err != nil {
			dtx, err = time.Parse("Jan 2006", out)
			if err == nil {
				return dtx, nil
			}
		} else {
			return dtx, nil
		}
	}

	// If we dont support partial dates, we return an error
	if !allowPartialDate && !dt.HasFullDate() {
		return time.Time{}, fmt.Errorf("found only partial date %v", dt.ISOFormat())
	}

	if !dt.Time.HasSecond() {
		dt.Time.SetSecond(0)
	}

	if !dt.HasTZOffset() {
		dt.Time.SetTZOffset(0)
	}

	return time.Parse("2006-01-02T15:04:05Z07:00", dt.ISOFormat())
}

func FilterFuzzyTime(src string, now time.Time, allowPartialDate bool) (string, error) {
	t, err := ParseFuzzyTime(src, now, allowPartialDate)
	if err != nil {
		return "", fmt.Errorf("error parsing fuzzy time %q: %v", src, err)
	}
	return t.Format(filterTimeFormat), nil
}

func ParseDate(layouts []string, value string) (string, error) {
	var err error
	for _, layout := range layouts {
		var t time.Time
		if t, err = time.Parse(layout, value); err == nil {
			return t.Format(filterTimeFormat), nil
		}
	}
	return "", fmt.Errorf("no matching date pattern for %s. %s", value, err)
}

func FilterSplit(sep string, pos int, value string) (string, error) {
	frags := strings.Split(value, sep)
	if pos < 0 {
		pos = len(frags) + pos
	}
	return frags[pos], nil
}