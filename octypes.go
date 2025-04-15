// types.go
package octypes

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"sync"
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

// Pre-allocated JSON values to avoid repeated allocations
var (
	// Common values
	nullJSON  = []byte("null")
	trueJSON  = []byte("true")
	falseJSON = []byte("false")
	
	// Digits
	digit0JSON = []byte("0")
	digit1JSON = []byte("1")
	digit2JSON = []byte("2")
	digit3JSON = []byte("3")
	digit4JSON = []byte("4")
	digit5JSON = []byte("5")
	digit6JSON = []byte("6")
	digit7JSON = []byte("7")
	digit8JSON = []byte("8")
	digit9JSON = []byte("9")
	
	// Empty values
	emptyStringJSON = []byte(`""`)
	emptyArrayJSON  = []byte("[]")
	emptyObjectJSON = []byte("{}")
	
	// Common patterns
	commaJSON     = []byte(",")
	colonJSON     = []byte(":")
	quoteJSON     = []byte(`"`)
	leftBraceJSON = []byte("{")
	rightBraceJSON = []byte("}")
)

// Pre-allocated digit map for quick integer lookups (0-99)
var digitMap [100][]byte

// Resource pools for reducing allocations
var timeResponsePool = sync.Pool{
	New: func() interface{} {
		return &TimeResponse{}
	},
}

// init initializes the pre-allocated values
func init() {
	// Initialize the digit map for numbers 0-99
	for i := 0; i < 100; i++ {
		digitMap[i] = []byte(strconv.Itoa(i))
	}
}

// NewCustomTimeNull creates a new CustomTime with a null value.
func NewCustomTimeNull() *CustomTime {
	return &CustomTime{NullTime: sql.NullTime{Valid: false}}
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
	t := time.Unix(0, int64Time*int64(time.Millisecond))
	return &CustomTime{
		NullTime: sql.NullTime{
			Time:  t,
			Valid: true,
		},
	}
}

// NewCustomTimeFloat64 creates a new CustomTime from float64 timestamp (milliseconds).
func NewCustomTimeFloat64(float64Time float64) *CustomTime {
	t := time.Unix(0, int64(float64Time)*int64(time.Millisecond))
	return &CustomTime{
		NullTime: sql.NullTime{
			Time:  t,
			Valid: true,
		},
	}
}

// Scan implements the sql.Scanner interface.
func (ct *CustomTime) Scan(value interface{}) error {
	if value == nil {
		ct.Valid = false
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		ct.Time = v
		ct.Valid = true
		return nil
	case string:
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			return err
		}
		ct.Time = t
		ct.Valid = true
		return nil
	default:
		return ct.NullTime.Scan(value)
	}
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
	// Use optimized implementation internally
	opt := OptimizedCustomTime{
		Time:  ct.Time,
		Valid: ct.Valid,
	}
	return opt.MarshalJSON()
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (ct *CustomTime) UnmarshalJSON(b []byte) error {
	// Use optimized implementation internally
	var opt OptimizedCustomTime
	err := opt.UnmarshalJSON(b)
	
	// Copy the values back
	ct.Time = opt.Time
	ct.Valid = opt.Valid
	
	return err
}

// WriteTo implements the io.WriterTo interface for binary serialization.
func (ct CustomTime) WriteTo(w io.Writer) (n int64, err error) {
	// Use optimized implementation internally
	opt := OptimizedCustomTime{
		Time:  ct.Time,
		Valid: ct.Valid,
	}
	return opt.WriteTo(w)
}

// ReadFrom implements the io.ReaderFrom interface for binary deserialization.
func (ct *CustomTime) ReadFrom(r io.Reader) (n int64, err error) {
	// Use optimized implementation internally
	var opt OptimizedCustomTime
	n, err = opt.ReadFrom(r)
	
	// Copy the values back
	ct.Time = opt.Time
	ct.Valid = opt.Valid
	
	return n, err
}

// NullString extends sql.NullString to handle JSON marshalling.
type NullString struct {
	sql.NullString
}

// NewNullStringNull creates a new NullString with an explicit null value.
func NewNullStringNull() *NullString {
	return &NullString{sql.NullString{Valid: false}}
}

// NewNullString creates a new NullString.
func NewNullString(s string) *NullString {
	// Maintain compatibility with tests - empty string is not valid
	return &NullString{sql.NullString{String: s, Valid: s != ""}}
}

// NewNullStringValid creates a new NullString that is always valid, even for empty strings.
func NewNullStringValid(s string) *NullString {
	return &NullString{sql.NullString{String: s, Valid: true}}
}

// Scan implements the sql.Scanner interface.
func (ns *NullString) Scan(value interface{}) error {
	return ns.NullString.Scan(value)
}

// Value implements the driver.Valuer interface.
func (ns NullString) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return ns.String, nil
}

// MarshalJSON implements the json.Marshaler interface.
func (ns NullString) MarshalJSON() ([]byte, error) {
	// Use optimized implementation internally
	opt := OptimizedNullString{
		String: ns.String,
		Valid:  ns.Valid,
	}
	return opt.MarshalJSON()
}

// containsSpecialChars checks if a string contains characters that need escaping in JSON
func containsSpecialChars(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < 32 || s[i] == '"' || s[i] == '\\' {
			return true
		}
	}
	return false
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (ns *NullString) UnmarshalJSON(b []byte) error {
	// Use optimized implementation internally
	var opt OptimizedNullString
	err := opt.UnmarshalJSON(b)
	
	// Copy the values back
	ns.String = opt.String
	ns.Valid = opt.Valid
	
	return err
}

// WriteTo implements the io.WriterTo interface for binary serialization.
func (ns NullString) WriteTo(w io.Writer) (n int64, err error) {
	// Use optimized implementation internally
	opt := OptimizedNullString{
		String: ns.String,
		Valid:  ns.Valid,
	}
	return opt.WriteTo(w)
}

// ReadFrom implements the io.ReaderFrom interface for binary deserialization.
func (ns *NullString) ReadFrom(r io.Reader) (n int64, err error) {
	// Use optimized implementation internally
	var opt OptimizedNullString
	n, err = opt.ReadFrom(r)
	
	// Copy the values back
	ns.String = opt.String
	ns.Valid = opt.Valid
	
	return n, err
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

// UnmarshalJSON implements the json.Unmarshaler interface.
func (lt *LocalizedText) UnmarshalJSON(b []byte) error {
	// Fast path for null
	if isNullJSON(b) {
		*lt = nil
		return nil
	}
	
	// Fast path for empty object
	if len(b) <= 2 && b[0] == '{' && b[len(b)-1] == '}' {
		*lt = make(LocalizedText)
		return nil
	}
	
	// Standard unmarshal for other cases
	m := make(map[string]string)
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	
	// Create a new map to ensure we start fresh
	*lt = make(LocalizedText, len(m))
	for k, v := range m {
		(*lt)[k] = v
	}
	
	return nil
}

// NullInt64 extends sql.NullInt64 to handle JSON marshalling.
type NullInt64 struct {
	sql.NullInt64
}

// NewNullInt64Null creates a new NullInt64 with an explicit null value.
func NewNullInt64Null() *NullInt64 {
	return &NullInt64{sql.NullInt64{Valid: false}}
}

// NewNullInt64 creates a new NullInt64 with the provided value.
func NewNullInt64(i int64) *NullInt64 {
	return &NullInt64{sql.NullInt64{Int64: i, Valid: true}}
}

// NewNullInt64Zero creates a new NullInt64 with value 0 that is valid.
func NewNullInt64Zero() *NullInt64 {
	return &NullInt64{sql.NullInt64{Int64: 0, Valid: true}}
}

// NewNullInt64FromString creates a new NullInt64 from a string.
func NewNullInt64FromString(s string) *NullInt64 {
	if s == "" {
		return NewNullInt64Null()
	}
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return NewNullInt64Null()
	}
	return NewNullInt64(i)
}

// Scan implements the sql.Scanner interface.
func (ni *NullInt64) Scan(value interface{}) error {
	return ni.NullInt64.Scan(value)
}

// Value implements the driver.Valuer interface.
func (ni NullInt64) Value() (driver.Value, error) {
	if !ni.Valid {
		return nil, nil
	}
	return ni.Int64, nil
}

// MarshalJSON implements the json.Marshaler interface.
func (ni NullInt64) MarshalJSON() ([]byte, error) {
	// Use optimized implementation internally
	opt := OptimizedNullInt64{
		Int64: ni.Int64,
		Valid: ni.Valid,
	}
	return opt.MarshalJSON()
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (ni *NullInt64) UnmarshalJSON(b []byte) error {
	// Use optimized implementation internally
	var opt OptimizedNullInt64
	err := opt.UnmarshalJSON(b)
	
	// Copy the values back
	ni.Int64 = opt.Int64
	ni.Valid = opt.Valid
	
	return err
}

// WriteTo implements the io.WriterTo interface for binary serialization.
func (ni NullInt64) WriteTo(w io.Writer) (n int64, err error) {
	// Use optimized implementation internally
	opt := OptimizedNullInt64{
		Int64: ni.Int64,
		Valid: ni.Valid,
	}
	return opt.WriteTo(w)
}

// ReadFrom implements the io.ReaderFrom interface for binary deserialization.
func (ni *NullInt64) ReadFrom(r io.Reader) (n int64, err error) {
	// Use optimized implementation internally
	var opt OptimizedNullInt64
	n, err = opt.ReadFrom(r)
	
	// Copy the values back
	ni.Int64 = opt.Int64
	ni.Valid = opt.Valid
	
	return n, err
}

// NullBool extends sql.NullBool to handle JSON marshalling.
type NullBool struct {
	sql.NullBool
}

// NewNullBoolNull creates a new NullBool with an explicit null value.
func NewNullBoolNull() *NullBool {
	return &NullBool{sql.NullBool{Valid: false}}
}

// NewNullBool creates a new NullBool.
func NewNullBool(b bool) *NullBool {
	return &NullBool{sql.NullBool{Bool: b, Valid: true}}
}

// NewNullBoolFalse creates a new NullBool with value false that is valid.
func NewNullBoolFalse() *NullBool {
	return &NullBool{sql.NullBool{Bool: false, Valid: true}}
}

// NewNullBoolFromString creates a new NullBool from a string.
func NewNullBoolFromString(s string) *NullBool {
	if s == "" {
		return NewNullBoolNull()
	}
	b, err := strconv.ParseBool(s)
	if err != nil {
		return NewNullBoolNull()
	}
	return NewNullBool(b)
}

// Scan implements the sql.Scanner interface.
func (nb *NullBool) Scan(value interface{}) error {
	return nb.NullBool.Scan(value)
}

// Value implements the driver.Valuer interface.
func (nb NullBool) Value() (driver.Value, error) {
	if !nb.Valid {
		return nil, nil
	}
	return nb.Bool, nil
}

// MarshalJSON implements the json.Marshaler interface.
func (nb NullBool) MarshalJSON() ([]byte, error) {
	// Use optimized implementation internally
	opt := OptimizedNullBool{
		Bool:  nb.Bool,
		Valid: nb.Valid,
	}
	return opt.MarshalJSON()
}

// isNullJSON is a fast null check
func isNullJSON(b []byte) bool {
	return len(b) == 4 && b[0] == 'n' && b[1] == 'u' && b[2] == 'l' && b[3] == 'l'
}

// isTrueJSON is a fast true check
func isTrueJSON(b []byte) bool {
	return len(b) == 4 && b[0] == 't' && b[1] == 'r' && b[2] == 'u' && b[3] == 'e'
}

// isFalseJSON is a fast false check
func isFalseJSON(b []byte) bool {
	return len(b) == 5 && b[0] == 'f' && b[1] == 'a' && b[2] == 'l' && b[3] == 's' && b[4] == 'e'
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (nb *NullBool) UnmarshalJSON(b []byte) error {
	// Use optimized implementation internally
	var opt OptimizedNullBool
	err := opt.UnmarshalJSON(b)
	
	// Copy the values back
	nb.Bool = opt.Bool
	nb.Valid = opt.Valid
	
	return err
}

// WriteTo implements the io.WriterTo interface for binary serialization.
func (nb NullBool) WriteTo(w io.Writer) (n int64, err error) {
	// Use optimized implementation internally
	opt := OptimizedNullBool{
		Bool:  nb.Bool,
		Valid: nb.Valid,
	}
	return opt.WriteTo(w)
}

// ReadFrom implements the io.ReaderFrom interface for binary deserialization.
func (nb *NullBool) ReadFrom(r io.Reader) (n int64, err error) {
	// Use optimized implementation internally
	var opt OptimizedNullBool
	n, err = opt.ReadFrom(r)
	
	// Copy the values back
	nb.Bool = opt.Bool
	nb.Valid = opt.Valid
	
	return n, err
}

// NullFloat64 extends sql.NullFloat64 to handle JSON marshalling.
type NullFloat64 struct {
	sql.NullFloat64
}

// NewNullFloat64Null creates a new NullFloat64 with an explicit null value.
func NewNullFloat64Null() *NullFloat64 {
	return &NullFloat64{sql.NullFloat64{Valid: false}}
}

// NewNullFloat64 creates a new NullFloat64.
func NewNullFloat64(f float64) *NullFloat64 {
	return &NullFloat64{sql.NullFloat64{Float64: f, Valid: true}}
}

// NewNullFloat64Zero creates a new NullFloat64 with value 0.0 that is valid.
func NewNullFloat64Zero() *NullFloat64 {
	return &NullFloat64{sql.NullFloat64{Float64: 0.0, Valid: true}}
}

// NewNullFloat64FromString creates a new NullFloat64 from a string.
func NewNullFloat64FromString(s string) *NullFloat64 {
	if s == "" {
		return NewNullFloat64Null()
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return NewNullFloat64Null()
	}
	return NewNullFloat64(f)
}

// Scan implements the sql.Scanner interface.
func (nf *NullFloat64) Scan(value interface{}) error {
	return nf.NullFloat64.Scan(value)
}

// Value implements the driver.Valuer interface.
func (nf NullFloat64) Value() (driver.Value, error) {
	if !nf.Valid {
		return nil, nil
	}
	return nf.Float64, nil
}

// MarshalJSON implements the json.Marshaler interface.
func (nf NullFloat64) MarshalJSON() ([]byte, error) {
	// Use optimized implementation internally
	opt := OptimizedNullFloat64{
		Float64: nf.Float64,
		Valid:   nf.Valid,
	}
	return opt.MarshalJSON()
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (nf *NullFloat64) UnmarshalJSON(b []byte) error {
	// Use optimized implementation internally
	var opt OptimizedNullFloat64
	err := opt.UnmarshalJSON(b)
	
	// Copy the values back
	nf.Float64 = opt.Float64
	nf.Valid = opt.Valid
	
	return err
}

// WriteTo implements the io.WriterTo interface for binary serialization.
func (nf NullFloat64) WriteTo(w io.Writer) (n int64, err error) {
	// Use optimized implementation internally
	opt := OptimizedNullFloat64{
		Float64: nf.Float64,
		Valid:   nf.Valid,
	}
	return opt.WriteTo(w)
}

// ReadFrom implements the io.ReaderFrom interface for binary deserialization.
func (nf *NullFloat64) ReadFrom(r io.Reader) (n int64, err error) {
	// Use optimized implementation internally
	var opt OptimizedNullFloat64
	n, err = opt.ReadFrom(r)
	
	// Copy the values back
	nf.Float64 = opt.Float64
	nf.Valid = opt.Valid
	
	return n, err
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

// UnmarshalJSON implements the json.Unmarshaler interface.
func (id *IntDictionary) UnmarshalJSON(b []byte) error {
	// Fast path for null
	if isNullJSON(b) {
		*id = nil
		return nil
	}
	
	// Fast path for empty object
	if len(b) <= 2 && b[0] == '{' && b[len(b)-1] == '}' {
		*id = make(IntDictionary)
		return nil
	}
	
	// Standard unmarshal for other cases
	m := make(map[string]int)
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	
	// Create a new map to ensure we start fresh
	*id = make(IntDictionary, len(m))
	for k, v := range m {
		(*id)[k] = v
	}
	
	return nil
}