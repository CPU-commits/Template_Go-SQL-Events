package utils

import (
	"crypto/rand"
	"encoding/base32"
	"errors"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"regexp"
	"strings"
	"time"
	"unicode"
)

func isUsername(u string) bool {
	if len(u) == 0 {
		return false
	}
	hasLetter := false
	for i := 0; i < len(u); i++ {
		c := u[i]
		switch {
		case c >= 'a' && c <= 'z':
			hasLetter = true
		case c >= '0' && c <= '9':
			// ok
		case c == '.' || c == '_':
			// ok
		default:
			return false
		}
	}
	return hasLetter
}

func DaysSinceCreation(createdAt time.Time) (int, error) {

	createdDate := time.Date(createdAt.Year(), createdAt.Month(), createdAt.Day(), 0, 0, 0, 0, time.UTC)
	today := time.Now().UTC().Truncate(24 * time.Hour)

	days := int(today.Sub(createdDate).Hours() / 24)

	return days, nil
}

func VerifyNotExpiredAt(expiration time.Time, clockType string, err error) error {
	var now time.Time

	switch clockType {
	case "utc":
		now = time.Now().UTC()
	case "local":
		now = time.Now()
	default:
		return errors.New("invalid clock type: must be 'utc' or 'local'")
	}

	if now.After(expiration) {
		return err
	}

	return nil
}

func GenerateRandomString(length int) (string, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	code := strings.ToUpper(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b))
	return code[:length], nil
}

func IterateDates[T any](from time.Time, to time.Time, fun func(d time.Time) T) (result []T) {
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		r := fun(d)
		result = append(result, r)
	}

	return result
}

func MonthStart(t time.Time) time.Time {
	loc := t.Location()

	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, loc)
}

func MonthEnd(t time.Time) time.Time {
	startNext := MonthStart(t).AddDate(0, 1, 0)

	return startNext.Add(-time.Nanosecond)
}

func CleanAndLower(s string) string {
	cleaned := strings.TrimSpace(s)
	return strings.ToLower(cleaned)
}

func Strim(s string) string {
	lines := strings.Split(s, "\n")

	for i, line := range lines {
		fields := strings.FieldsFunc(line, func(r rune) bool {
			return unicode.IsSpace(r) && r != '\n'
		})
		lines[i] = strings.Join(fields, " ")
	}

	return strings.Join(lines, "\n")
}

func NormalizeFileName(name string) string {
	name = strings.ToLower(name)

	name = strings.ReplaceAll(name, " ", "_")

	var sb strings.Builder
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '_' || r == '.':
			sb.WriteRune(r)
		default:
			if base := removeAccents(r); base != 0 {
				sb.WriteRune(base)
			}
		}
	}

	re := regexp.MustCompile(`_+`)
	name = re.ReplaceAllString(sb.String(), "_")

	return name
}
func removeAccents(r rune) rune {
	switch r {
	case 'á', 'à', 'ä', 'â':
		return 'a'
	case 'é', 'è', 'ë', 'ê':
		return 'e'
	case 'í', 'ì', 'ï', 'î':
		return 'i'
	case 'ó', 'ò', 'ö', 'ô':
		return 'o'
	case 'ú', 'ù', 'ü', 'û':
		return 'u'
	case 'ñ':
		return 'n'
	default:
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return r
		}
		return 0
	}
}
func FormatEmailDate(t *time.Time, tz string) string {
	if t == nil {
		return ""
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		loc = time.Local
	}

	tt := t.In(loc)
	return tt.Format("03:04 PM 02-01-2006")
}
