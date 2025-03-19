// types.go
package octypes

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strconv"
	"time"
)

// Pagination represents pagination details.
type Pagination struct {
	PageNo         int `json:"page_no"`
	ResultsPerPage int `json:"results_per_page"`
	PageMax        int `json:"page_max"`
	Count          int `json:"count"`
}

// CustomTime extends sql.NullTime to handle custom time formats.
type CustomTime struct {
	sql.NullTime
}

// TimeResponse represents various time formats for JSON marshalling.
type TimeResponse struct {
	ISO    string `json:"iso"`
	TZ     string `json:"tz"`
	Unix   int64  `json:"unix"`
	UnixMS int64  `json:"unixms"`
	US     int64  `json:"us"`
	Full   int64  `json:"full,omitempty,string"`
}

// NewCustomTimeNull creates a new CustomTime with a null value.
func NewCustomTimeNull() *CustomTime {
	return &CustomTime{}
}

// NewCustomTime creates a new CustomTime from time.Time.
func NewCustomTime(t time.Time) *CustomTime {
	return &CustomTime{
		NullTime: sql.NullTime{
			Time:  t,
			Valid: true,
		},
	}
}

// NewCustomTimeInt64 creates a new CustomTime from int64 timestamp (milliseconds).
func NewCustomTimeInt64(int64Time int64) *CustomTime {
	return NewCustomTime(time.Unix(0, int64Time*int64(time.Millisecond)))
}

// NewCustomTimeFloat64 creates a new CustomTime from float64 timestamp (milliseconds).
func NewCustomTimeFloat64(float64Time float64) *CustomTime {
	return NewCustomTime(time.Unix(0, int64(float64Time)*int64(time.Millisecond)))
}

// Scan implements the sql.Scanner interface.
func (ct *CustomTime) Scan(value interface{}) error {
	if value == nil {
		*ct = CustomTime{}
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		ct.Time = v
		ct.Valid = true
	case string:
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			return err
		}
		ct.Time = t
		ct.Valid = true
	default:
		return ct.NullTime.Scan(value)
	}
	return nil
}

// Value implements the driver.Valuer interface.
func (ct CustomTime) Value() (driver.Value, error) {
	if !ct.Valid {
		return nil, nil
	}
	return ct.Time, nil
}

// MarshalJSON implements the json.Marshaler interface.
func (ct CustomTime) MarshalJSON() ([]byte, error) {
	if !ct.Valid {
		return json.Marshal(nil)
	}

	tr := TimeResponse{
		ISO:    ct.Time.Format(time.RFC3339Nano),
		TZ:     ct.Time.Location().String(),
		Unix:   ct.Time.Unix(),
		UnixMS: ct.Time.UnixMilli(),
		US:     int64(ct.Time.Nanosecond()),
		Full:   ct.Time.UnixMicro(),
	}

	return json.Marshal(tr)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (ct *CustomTime) UnmarshalJSON(b []byte) error {
	// Handle null input
	if string(b) == "null" {
		*ct = CustomTime{}
		return nil
	}

	var tr TimeResponse
	if err := json.Unmarshal(b, &tr); err == nil && tr.ISO != "" {
		t, err := time.Parse(time.RFC3339Nano, tr.ISO)
		if err != nil {
			return err
		}
		ct.Time = t
		ct.Valid = true
		return nil
	}

	var unixms int64
	if err := json.Unmarshal(b, &unixms); err == nil {
		ct.Time = time.Unix(0, unixms*int64(time.Millisecond))
		ct.Valid = true
		return nil
	}

	var floatUnixms float64
	if err := json.Unmarshal(b, &floatUnixms); err == nil {
		ct.Time = time.Unix(0, int64(floatUnixms)*int64(time.Millisecond))
		ct.Valid = true
		return nil
	}

	var ts string
	if err := json.Unmarshal(b, &ts); err == nil {
		t, err := time.Parse("2006-01-02", ts)
		if err != nil {
			return err
		}
		ct.Time = t
		ct.Valid = true
		return nil
	}

	return errors.New("invalid time format")
}

// NullString extends sql.NullString to handle JSON marshalling.
type NullString struct {
	sql.NullString
}

// NewNullString creates a new NullString.
func NewNullString(s string) *NullString {
	return &NullString{sql.NullString{String: s, Valid: s != ""}}
}

// Scan implements the sql.Scanner interface.
func (ns *NullString) Scan(value interface{}) error {
	return ns.NullString.Scan(value)
}

// Value implements the driver.Valuer interface.
func (ns NullString) Value() (driver.Value, error) {
	if ns.Valid {
		return ns.String, nil
	}
	return nil, nil
}

// MarshalJSON implements the json.Marshaler interface.
func (ns NullString) MarshalJSON() ([]byte, error) {
	if ns.Valid {
		return json.Marshal(ns.String)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (ns *NullString) UnmarshalJSON(b []byte) error {
	var s *string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	if s != nil {
		ns.String = *s
		ns.Valid = true
	} else {
		ns.Valid = false
	}
	return nil
}

// LocalizedText represents a map of localized strings.
type LocalizedText map[string]string

// Scan implements the sql.Scanner interface.
func (lt *LocalizedText) Scan(value interface{}) error {
	if value == nil {
		*lt = nil
		return nil
	}
	asBytes, ok := value.([]byte)
	if !ok {
		return errors.New("Scan source is not []byte")
	}
	// Reset lt before unmarshalling
	*lt = make(LocalizedText)
	return json.Unmarshal(asBytes, lt)
}

// Value implements the driver.Valuer interface.
func (lt LocalizedText) Value() (driver.Value, error) {
	if lt == nil {
		return nil, nil
	}
	return json.Marshal(lt)
}

// NullInt64 extends sql.NullInt64 to handle JSON marshalling.
type NullInt64 struct {
	sql.NullInt64
}

// NewNullInt64 creates a new NullInt64.
func NewNullInt64(i int64) *NullInt64 {
	return &NullInt64{sql.NullInt64{Int64: i, Valid: true}}
}

// NewNullInt64FromString creates a new NullInt64 from a string.
func NewNullInt64FromString(s string) *NullInt64 {
	if s == "" {
		return &NullInt64{}
	}
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return &NullInt64{}
	}
	return NewNullInt64(i)
}

// Scan implements the sql.Scanner interface.
func (ni *NullInt64) Scan(value interface{}) error {
	return ni.NullInt64.Scan(value)
}

// Value implements the driver.Valuer interface.
func (ni NullInt64) Value() (driver.Value, error) {
	if ni.Valid {
		return ni.Int64, nil
	}
	return nil, nil
}

// MarshalJSON implements the json.Marshaler interface.
func (ni NullInt64) MarshalJSON() ([]byte, error) {
	if ni.Valid {
		return json.Marshal(ni.Int64)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (ni *NullInt64) UnmarshalJSON(b []byte) error {
	var i *int64
	if err := json.Unmarshal(b, &i); err == nil {
		if i != nil {
			ni.Int64 = *i
			ni.Valid = true
		} else {
			ni.Valid = false
		}
		return nil
	}
	return errors.New("invalid int64 format")
}

// NullBool extends sql.NullBool to handle JSON marshalling.
type NullBool struct {
	sql.NullBool
}

// NewNullBool creates a new NullBool.
func NewNullBool(b bool) *NullBool {
	return &NullBool{sql.NullBool{Bool: b, Valid: true}}
}

// NewNullBoolFromString creates a new NullBool from a string.
func NewNullBoolFromString(s string) *NullBool {
	if s == "" {
		return &NullBool{}
	}
	b, err := strconv.ParseBool(s)
	if err != nil {
		return &NullBool{}
	}
	return NewNullBool(b)
}

// Scan implements the sql.Scanner interface.
func (nb *NullBool) Scan(value interface{}) error {
	return nb.NullBool.Scan(value)
}

// Value implements the driver.Valuer interface.
func (nb NullBool) Value() (driver.Value, error) {
	if nb.Valid {
		return nb.Bool, nil
	}
	return nil, nil
}

// MarshalJSON implements the json.Marshaler interface.
func (nb NullBool) MarshalJSON() ([]byte, error) {
	if nb.Valid {
		return json.Marshal(nb.Bool)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (nb *NullBool) UnmarshalJSON(b []byte) error {
	var bl *bool
	if err := json.Unmarshal(b, &bl); err != nil {
		return err
	}
	if bl != nil {
		nb.Bool = *bl
		nb.Valid = true
	} else {
		nb.Valid = false
	}
	return nil
}

// NullFloat64 extends sql.NullFloat64 to handle JSON marshalling.
type NullFloat64 struct {
	sql.NullFloat64
}

// NewNullFloat64 creates a new NullFloat64.
func NewNullFloat64(f float64) *NullFloat64 {
	return &NullFloat64{sql.NullFloat64{Float64: f, Valid: true}}
}

// NewNullFloat64FromString creates a new NullFloat64 from a string.
func NewNullFloat64FromString(s string) *NullFloat64 {
	if s == "" {
		return &NullFloat64{}
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return &NullFloat64{}
	}
	return NewNullFloat64(f)
}

// Scan implements the sql.Scanner interface.
func (nf *NullFloat64) Scan(value interface{}) error {
	return nf.NullFloat64.Scan(value)
}

// Value implements the driver.Valuer interface.
func (nf NullFloat64) Value() (driver.Value, error) {
	if nf.Valid {
		return nf.Float64, nil
	}
	return nil, nil
}

// MarshalJSON implements the json.Marshaler interface.
func (nf NullFloat64) MarshalJSON() ([]byte, error) {
	if nf.Valid {
		return json.Marshal(nf.Float64)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (nf *NullFloat64) UnmarshalJSON(b []byte) error {
	var f *float64
	if err := json.Unmarshal(b, &f); err != nil {
		return err
	}
	if f != nil {
		nf.Float64 = *f
		nf.Valid = true
	} else {
		nf.Valid = false
	}
	return nil
}

// IntDictionary represents a map of string to int.
type IntDictionary map[string]int

// Scan implements the sql.Scanner interface.
func (id *IntDictionary) Scan(value interface{}) error {
	if value == nil {
		*id = nil
		return nil
	}
	asBytes, ok := value.([]byte)
	if !ok {
		return errors.New("Scan source is not []byte")
	}
	// Reset id before unmarshalling
	*id = make(IntDictionary)
	return json.Unmarshal(asBytes, id)
}

// Value implements the driver.Valuer interface.
func (id IntDictionary) Value() (driver.Value, error) {
	if id == nil {
		return nil, nil
	}
	return json.Marshal(id)
}
