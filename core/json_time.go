package core

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	timeFmt   = "2006-01-02 15:04:05"
	timeRegex = `-|:|T|\.`
)

type JSONTime time.Time

func (t *JSONTime) String() string {
	return time.Time(*t).Format(timeFmt)
}

func NewJSONTime(t time.Time) *JSONTime {
	now := JSONTime(t.UTC())

	return &now
}

// This method for mapping JSONTime to datetime data type in sql
func (t *JSONTime) Value() (driver.Value, error) {
	if t == nil {
		return nil, nil
	}
	return time.Time(*t).Format(timeFmt), nil
}

// This method for scanning JSONTime from datetime data type in sql
func (t *JSONTime) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	if v, ok := value.(time.Time); ok {
		*t = JSONTime(v)
		return nil
	}

	return errors.New("invalid Scan Source")
}

// Implement method MarshalJSON to output time with in formatted
func (t JSONTime) MarshalJSON() ([]byte, error) {
	stamp := fmt.Sprintf("\"%s\"", time.Time(t).Format(timeFmt))
	return []byte(stamp), nil
}

func (t *JSONTime) UnmarshalJSON(data []byte) error {
	value := string(data)

	matched, err := regexp.MatchString(timeRegex, value)
	if err != nil {
		return err
	}

	if matched {
		ti, err := time.Parse(timeFmt, strings.Replace(string(data), "\"", "", -1))
		if err != nil {
			return err
		}
		*t = JSONTime(ti)

		return nil
	}

	i, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return err
	}
	*t = JSONTime(time.UnixMilli(i))

	return nil
}
