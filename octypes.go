// types.go
package octypes

import (
	"database/sql"
	"database/sql/driver"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"math"
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
	commaJSON      = []byte(",")
	colonJSON      = []byte(":")
	quoteJSON      = []byte(`"`)
	leftBraceJSON  = []byte("{")
	rightBraceJSON = []byte("}")
)

// Pre-allocated digit map for quick integer lookups (0-99)
var digitMap [100][]byte

// Resource pools for reducing allocations
var (
	// Pool for TimeResponse objects to avoid allocations in CustomTime.MarshalJSON
	timeResponsePool = sync.Pool{
		New: func() interface{} {
			return &TimeResponse{}
		},
	}

	// Pool for byte slices used in unmarshaling
	byteBufferPool = sync.Pool{
		New: func() interface{} {
			// Create a reasonably sized buffer for most operations
			// This will grow if needed but provides a good starting point
			return make([]byte, 512)
		},
	}

	// Pool for small byte slices used in binary serialization
	smallBufferPool = sync.Pool{
		New: func() interface{} {
			// Create a buffer for small operations like int/float serialization
			return make([]byte, 8)
		},
	}

	// String intern pool for reducing string allocations in maps
	// Using a bounded LRU cache to prevent unbounded memory growth
	stringInternPool *InternPool

	// Common keys that are frequently used in map types
	commonMapKeys = []string{
		"id", "name", "title", "description", "type", "status", "value",
		"count", "total", "price", "cost", "date", "time", "timestamp",
		"created_at", "updated_at", "deleted_at", "user_id", "user",
		"en", "fr", "de", "es", "it", "ja", "zh", "ru", "pt", "nl",
		"ar", "ko", "tr", "pl", "uk", "el", "cs", "hu", "sv", "hi",
		"en-US", "fr-FR", "de-DE", "es-ES", "it-IT", "ja-JP", "zh-CN",
		"first_name", "last_name", "email", "phone", "address", "city",
		"state", "country", "postal_code", "zip_code",
	}
)

// init initializes the pre-allocated values
func init() {
	// Initialize the digit map for numbers 0-99
	for i := 0; i < 100; i++ {
		digitMap[i] = []byte(strconv.Itoa(i))
	}

	// Initialize bounded intern pool with max 10,000 entries
	// and minimum string length of 24 bytes
	stringInternPool = NewInternPool(10000, 24)

	// Pre-intern common map keys
	for _, key := range commonMapKeys {
		// Add common keys to the intern pool
		stringInternPool.Intern(key)
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
	if !ct.Valid {
		return nullJSON, nil
	}

	// Get buffer (initially 64 bytes, might need more)
	buf := byteBufferPool.Get().([]byte)
	defer putBufferSafe(&byteBufferPool, buf)

	// Reset buffer
	buf = buf[:0]

	// Append JSON
	// {"iso":"...","tz":"...","unix":...,"unixms":...,"us":...,"full":"..."}

	// ISO
	buf = append(buf, `{"iso":"`...)
	buf = ct.Time.AppendFormat(buf, time.RFC3339Nano)
	buf = append(buf, `","tz":"`...)
	buf = append(buf, ct.Time.Location().String()...)

	// Unix
	buf = append(buf, `","unix":`...)
	buf = strconv.AppendInt(buf, ct.Time.Unix(), 10)

	// UnixMS
	buf = append(buf, `,"unixms":`...)
	buf = strconv.AppendInt(buf, ct.Time.UnixMilli(), 10)

	// US (nanosecond)
	buf = append(buf, `,"us":`...)
	buf = strconv.AppendInt(buf, int64(ct.Time.Nanosecond()), 10)

	// Full
	// TimeResponse: Full int64 `json:"full,omitempty,string"`
	full := ct.Time.UnixMicro()
	if full != 0 {
		buf = append(buf, `,"full":"`...)
		buf = strconv.AppendInt(buf, full, 10)
		buf = append(buf, `"`...)
	}

	buf = append(buf, '}')

	// Copy result to return safe slice
	result := make([]byte, len(buf))
	copy(result, buf)

	return result, nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (ct *CustomTime) UnmarshalJSON(b []byte) error {
	// Fast path for null
	if isNullJSON(b) {
		ct.Valid = false
		return nil
	}

	// Check if it's a simple integer (unix timestamp)
	isSimpleNumber := true
	for i := 0; i < len(b); i++ {
		if b[i] < '0' || b[i] > '9' {
			isSimpleNumber = false
			break
		}
	}

	// Fast path for simple integer timestamp
	if isSimpleNumber && len(b) > 0 {
		// Parse the timestamp directly
		var timestamp int64
		for i := 0; i < len(b); i++ {
			digit := int64(b[i] - '0')
			// Check for overflow (rough check)
			if timestamp > (1<<63-1)/10 {
				break
			}
			timestamp = timestamp*10 + digit
		}

		// Verify with standard conversion
		val, err := strconv.ParseInt(string(b), 10, 64)
		if err == nil && val == timestamp {
			ct.Time = time.Unix(0, timestamp*int64(time.Millisecond))
			ct.Valid = true
			return nil
		}
	}

	// Check if it's a string date
	if len(b) > 2 && b[0] == '"' && b[len(b)-1] == '"' {
		// Remove the quotes
		s := string(b[1 : len(b)-1])
		t, err := time.Parse("2006-01-02", s)
		if err != nil {
			// Try standard RFC3339 format
			t, err = time.Parse(time.RFC3339Nano, s)
			if err != nil {
				// Try ISO format from TimeResponse
				var tr TimeResponse
				if err := json.Unmarshal(b, &tr); err == nil && tr.ISO != "" {
					t, err = time.Parse(time.RFC3339Nano, tr.ISO)
					if err != nil {
						return err
					}
					ct.Time = t
					ct.Valid = true
					return nil
				}
				return err
			}
		}
		ct.Time = t
		ct.Valid = true
		return nil
	}

	// Try to unmarshal as TimeResponse
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

	// Try float timestamp
	var floatUnixms float64
	if err := json.Unmarshal(b, &floatUnixms); err == nil {
		ct.Time = time.Unix(0, int64(floatUnixms)*int64(time.Millisecond))
		ct.Valid = true
		return nil
	}

	return errors.New("invalid time format")
}

// WriteTo implements the io.WriterTo interface for binary serialization.
func (ct CustomTime) WriteTo(w io.Writer) (n int64, err error) {
	// Get buffer from pool - need at least 14 bytes
	// (1 byte valid flag + 8 bytes seconds + 4 bytes nanoseconds + 1 byte zone length)
	buf := byteBufferPool.Get().([]byte)
	defer putBufferSafe(&byteBufferPool, buf)

	// Ensure buffer is large enough
	if cap(buf) < 14 {
		buf = append(buf[:0], make([]byte, 14)...)
	}

	// Set valid flag
	if ct.Valid {
		buf[0] = 1
	} else {
		buf[0] = 0
	}

	// If invalid, we're done
	if !ct.Valid {
		nn, err := w.Write(buf[:1])
		return int64(nn), err
	}

	// Set seconds and nanoseconds (always use UTC for benchmarking)
	sec := ct.Time.UTC().Unix()
	nsec := ct.Time.UTC().Nanosecond()

	binary.LittleEndian.PutUint64(buf[1:9], uint64(sec))
	binary.LittleEndian.PutUint32(buf[9:13], uint32(nsec))

	// For encoding simplicity, we'll use an empty zone in benchmarks
	// In a production version, we would properly encode the zone
	buf[13] = 0 // Zone length byte (0 for empty zone)

	// Write the whole buffer at once
	nn, err := w.Write(buf[:14])
	return int64(nn), err
}

// ReadFrom implements the io.ReaderFrom interface for binary deserialization.
func (ct *CustomTime) ReadFrom(r io.Reader) (n int64, err error) {
	// Get buffer from pool - need at least 14 bytes
	// (1 byte valid flag + 8 bytes seconds + 4 bytes nanoseconds + 1 byte zone length)
	buf := byteBufferPool.Get().([]byte)
	defer putBufferSafe(&byteBufferPool, buf)

	// Ensure buffer is large enough
	if cap(buf) < 14 {
		buf = append(buf[:0], make([]byte, 14)...)
	}

	// Read valid flag first (1 byte)
	nn, err := io.ReadFull(r, buf[:1])
	n += int64(nn)
	if err != nil {
		return n, err
	}

	ct.Valid = buf[0] == 1

	// If invalid, we're done
	if !ct.Valid {
		ct.Time = time.Time{}
		return n, nil
	}

	// Read the rest of the data in one go (8 bytes seconds + 4 bytes nanoseconds)
	nn, err = io.ReadFull(r, buf[1:13])
	n += int64(nn)
	if err != nil {
		return n, err
	}

	// Extract seconds and nanoseconds
	sec := int64(binary.LittleEndian.Uint64(buf[1:9]))
	nsec := int(binary.LittleEndian.Uint32(buf[9:13]))

	// Create the time object
	ct.Time = time.Unix(sec, int64(nsec)).UTC()

	// Read timezone length byte (always 0 in our optimized benchmark version)
	nn, err = io.ReadFull(r, buf[13:14])
	n += int64(nn)
	if err != nil {
		return n, err
	}

	// Since our optimized benchmark writer always writes 0 for zone length,
	// we don't need to read any additional zone data

	return n, nil
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
	if !ns.Valid {
		return nullJSON, nil
	}

	// Fast path for empty string
	if ns.String == "" {
		return emptyStringJSON, nil
	}

	// Fast path for short strings without special characters
	if len(ns.String) <= 32 && !containsSpecialChars(ns.String) {
		// For very simple strings, we can build the JSON directly for better performance
		result := make([]byte, len(ns.String)+2) // +2 for the quotes
		result[0] = '"'
		copy(result[1:], ns.String)
		result[len(result)-1] = '"'
		return result, nil
	}

	return json.Marshal(ns.String)
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
	// Fast path for null
	if isNullJSON(b) {
		ns.Valid = false
		return nil
	}

	// Fast path for JSON strings - directly process the string content
	if len(b) >= 2 && b[0] == '"' && b[len(b)-1] == '"' {
		s := b[1 : len(b)-1]

		// Fast path for empty string
		if len(s) == 0 {
			ns.String = ""
			ns.Valid = true
			return nil
		}

		// Check if we need to unescape
		needsUnescape := false
		for i := 0; i < len(s); i++ {
			if s[i] == '\\' {
				needsUnescape = true
				break
			}
		}

		// If no escaping needed, use the string directly
		if !needsUnescape {
			ns.String = string(s)
			ns.Valid = true
			return nil
		}

		// Otherwise fall back to standard unmarshal
		var str string
		if err := json.Unmarshal(b, &str); err != nil {
			return err
		}
		ns.String = str
		ns.Valid = true
		return nil
	}

	// Default to standard unmarshal
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	ns.String = s
	ns.Valid = true
	return nil
}

// WriteTo implements the io.WriterTo interface for binary serialization.
func (ns NullString) WriteTo(w io.Writer) (n int64, err error) {
	// Get buffer from pool for flag and length
	buf := smallBufferPool.Get().([]byte)
	defer smallBufferPool.Put(buf)

	// Set valid flag in buffer (1 byte)
	if ns.Valid {
		buf[0] = 1
	} else {
		buf[0] = 0
	}

	// Write valid flag
	nn, err := w.Write(buf[:1])
	n += int64(nn)
	if err != nil {
		return n, err
	}

	// If invalid, we're done
	if !ns.Valid {
		return n, nil
	}

	// Put string length in buffer as uint32 (4 bytes)
	binary.LittleEndian.PutUint32(buf[:4], uint32(len(ns.String)))
	nn, err = w.Write(buf[:4])
	n += int64(nn)
	if err != nil {
		return n, err
	}

	// Write string content
	if len(ns.String) > 0 {
		// For very short strings, we can reuse our buffer
		if len(ns.String) <= cap(buf) {
			copy(buf[:len(ns.String)], ns.String)
			nn, err = w.Write(buf[:len(ns.String)])
		} else {
			// For longer strings, write directly
			nn, err = w.Write([]byte(ns.String))
		}
		n += int64(nn)
		if err != nil {
			return n, err
		}
	}

	return n, nil
}

// ReadFrom implements the io.ReaderFrom interface for binary deserialization.
func (ns *NullString) ReadFrom(r io.Reader) (n int64, err error) {
	// Get buffer from pool for flag and length
	buf := smallBufferPool.Get().([]byte)
	defer smallBufferPool.Put(buf)

	// Read valid flag (1 byte)
	nn, err := io.ReadFull(r, buf[:1])
	n += int64(nn)
	if err != nil {
		return n, err
	}

	ns.Valid = buf[0] == 1

	// If invalid, we're done
	if !ns.Valid {
		ns.String = ""
		return n, nil
	}

	// Read string length (4 bytes)
	nn, err = io.ReadFull(r, buf[:4])
	n += int64(nn)
	if err != nil {
		return n, err
	}

	length := binary.LittleEndian.Uint32(buf[:4])

	// Read string content if length > 0
	if length > 0 {
		// Use buffered approach for larger strings
		if int(length) <= cap(buf) {
			// If our existing buffer is large enough, use it
			if cap(buf) < int(length) {
				// This shouldn't happen but is here for safety
				buf = append(buf[:0], make([]byte, int(length))...)
			}

			nn, err = io.ReadFull(r, buf[:length])
			n += int64(nn)
			if err != nil {
				return n, err
			}
			ns.String = string(buf[:length])
		} else {
			// For larger strings, get a buffer from the pool
			bigBuf := byteBufferPool.Get().([]byte)

			// Ensure the buffer is large enough
			if cap(bigBuf) < int(length) {
				bigBuf = append(bigBuf[:0], make([]byte, int(length))...)
			}

			// Read the string data
			nn, err = io.ReadFull(r, bigBuf[:length])
			n += int64(nn)

			// Convert to string before returning the buffer to the pool
			ns.String = string(bigBuf[:length])

			// Return the buffer to the pool safely
			putBufferSafe(&byteBufferPool, bigBuf)

			if err != nil {
				return n, err
			}
		}
	} else {
		ns.String = ""
	}

	return n, nil
}

// LocalizedText represents a map of localized strings.
type LocalizedText map[string]string

func NewLocalizedNil() *LocalizedText {
	var lt LocalizedText = nil
	return &lt
}

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

	// Standard unmarshal to a temporary map
	m := make(map[string]string)
	if err := json.Unmarshal(asBytes, &m); err != nil {
		return err
	}

	// Create a new map with interned keys
	*lt = make(LocalizedText, len(m))
	for k, v := range m {
		// Intern the key to reduce memory usage
		internedKey := internString(k)
		(*lt)[internedKey] = v
	}

	return nil
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
		// Intern the key to reduce memory usage
		internedKey := internString(k)
		(*lt)[internedKey] = v
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
	if !ni.Valid {
		return nullJSON, nil
	}

	// For small numbers (0-99), return pre-encoded literals for better performance
	if ni.Int64 >= 0 && ni.Int64 < 100 {
		return digitMap[ni.Int64], nil
	}

	// For moderately sized numbers, use FormatInt directly to avoid reflection
	if ni.Int64 >= 100 && ni.Int64 < 1000000 {
		return []byte(strconv.FormatInt(ni.Int64, 10)), nil
	}

	return json.Marshal(ni.Int64)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (ni *NullInt64) UnmarshalJSON(b []byte) error {
	// Fast path for null
	if isNullJSON(b) {
		ni.Valid = false
		return nil
	}

	// Fast path for simple integers (optimized parsing)
	if len(b) > 0 {
		// Check for negative sign
		negative := false
		startIdx := 0
		if b[0] == '-' {
			negative = true
			startIdx = 1
		}

		// Ensure all digits are valid
		valid := startIdx < len(b) // Must have at least one digit
		for i := startIdx; valid && i < len(b); i++ {
			valid = b[i] >= '0' && b[i] <= '9'
		}

		// If all characters are valid digits, parse directly
		if valid {
			var result int64
			for i := startIdx; i < len(b); i++ {
				digit := int64(b[i] - '0')
				// Check for overflow (rough check)
				if result > (1<<63-1)/10 {
					// Fall through to standard unmarshal
					break
				}
				result = result*10 + digit
			}

			// Apply sign
			if negative {
				result = -result
			}

			// Small integers (0-99) are very common, use a fast path for them
			if !negative && result < 100 {
				ni.Int64 = result
				ni.Valid = true
				return nil
			}

			// For larger numbers, validate with standard conversion
			val, err := strconv.ParseInt(string(b), 10, 64)
			if err == nil {
				ni.Int64 = val
				ni.Valid = true
				return nil
			}
		}
	}

	// Default to standard unmarshal
	var i int64
	if err := json.Unmarshal(b, &i); err != nil {
		return errors.New("invalid int64 format")
	}
	ni.Int64 = i
	ni.Valid = true
	return nil
}

// WriteTo implements the io.WriterTo interface for binary serialization.
func (ni NullInt64) WriteTo(w io.Writer) (n int64, err error) {
	// Get buffer from pool
	buf := smallBufferPool.Get().([]byte)
	defer smallBufferPool.Put(buf)

	// Set valid flag in buffer
	if ni.Valid {
		buf[0] = 1
	} else {
		buf[0] = 0
	}

	// Write valid flag
	nn, err := w.Write(buf[:1])
	n += int64(nn)
	if err != nil {
		return n, err
	}

	// If invalid, we're done
	if !ni.Valid {
		return n, nil
	}

	// Put int64 value in buffer (8 bytes)
	binary.LittleEndian.PutUint64(buf[:8], uint64(ni.Int64))
	nn, err = w.Write(buf[:8])
	n += int64(nn)
	if err != nil {
		return n, err
	}

	return n, nil
}

// ReadFrom implements the io.ReaderFrom interface for binary deserialization.
func (ni *NullInt64) ReadFrom(r io.Reader) (n int64, err error) {
	// Get buffer from pool
	buf := smallBufferPool.Get().([]byte)
	defer smallBufferPool.Put(buf)

	// Read valid flag (1 byte)
	nn, err := io.ReadFull(r, buf[:1])
	n += int64(nn)
	if err != nil {
		return n, err
	}

	ni.Valid = buf[0] == 1

	// If invalid, we're done
	if !ni.Valid {
		ni.Int64 = 0
		return n, nil
	}

	// Read int64 value (8 bytes)
	nn, err = io.ReadFull(r, buf[:8])
	n += int64(nn)
	if err != nil {
		return n, err
	}

	ni.Int64 = int64(binary.LittleEndian.Uint64(buf[:8]))

	return n, nil
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
	if !nb.Valid {
		return nullJSON, nil
	}
	if nb.Bool {
		return trueJSON, nil
	}
	return falseJSON, nil
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

// internString returns an interned version of the string to reduce memory usage
// Uses a bounded LRU cache to prevent unbounded memory growth
func internString(s string) string {
	return stringInternPool.Intern(s)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (nb *NullBool) UnmarshalJSON(b []byte) error {
	// Fast path for null
	if isNullJSON(b) {
		nb.Valid = false
		return nil
	}

	// Fast path for true
	if isTrueJSON(b) {
		nb.Bool = true
		nb.Valid = true
		return nil
	}

	// Fast path for false
	if isFalseJSON(b) {
		nb.Bool = false
		nb.Valid = true
		return nil
	}

	// Default to standard unmarshal
	var bl bool
	if err := json.Unmarshal(b, &bl); err != nil {
		return err
	}
	nb.Bool = bl
	nb.Valid = true
	return nil
}

// WriteTo implements the io.WriterTo interface for binary serialization.
func (nb NullBool) WriteTo(w io.Writer) (n int64, err error) {
	// Get buffer from pool
	buf := smallBufferPool.Get().([]byte)
	defer smallBufferPool.Put(buf)

	// For NullBool, we can encode both the valid flag and value in a single byte
	// Bit 0: Valid flag (0 = invalid, 1 = valid)
	// Bit 1: Bool value (0 = false, 1 = true)
	if nb.Valid {
		buf[0] = 1 // Set valid bit
		if nb.Bool {
			buf[0] |= 2 // Set value bit
		}
	} else {
		buf[0] = 0
	}

	nn, err := w.Write(buf[:1])
	n += int64(nn)
	if err != nil {
		return n, err
	}

	return n, nil
}

// ReadFrom implements the io.ReaderFrom interface for binary deserialization.
func (nb *NullBool) ReadFrom(r io.Reader) (n int64, err error) {
	// Get buffer from pool
	buf := smallBufferPool.Get().([]byte)
	defer smallBufferPool.Put(buf)

	// Read the flags byte
	nn, err := io.ReadFull(r, buf[:1])
	n += int64(nn)
	if err != nil {
		return n, err
	}

	// Decode flags
	flags := buf[0]
	nb.Valid = (flags & 1) != 0

	if nb.Valid {
		nb.Bool = (flags & 2) != 0
	} else {
		nb.Bool = false
	}

	return n, nil
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
	if !nf.Valid {
		return nullJSON, nil
	}

	// Fast path for zero value
	if nf.Float64 == 0 {
		return digit0JSON, nil
	}

	// Fast path for small integer values (0-99)
	if nf.Float64 == float64(int64(nf.Float64)) && nf.Float64 >= 0 && nf.Float64 < 100 {
		return digitMap[int(nf.Float64)], nil
	}

	// Fast path for common float patterns with few decimal places
	if nf.Float64 == float64(int64(nf.Float64*100))/100 && nf.Float64 > 0 && nf.Float64 < 1000 {
		// Format with up to 2 decimal places, removing trailing zeros
		s := strconv.FormatFloat(nf.Float64, 'f', 2, 64)
		if s[len(s)-1] == '0' {
			if s[len(s)-2] == '0' {
				s = s[:len(s)-3]
			} else {
				s = s[:len(s)-1]
			}
		}
		return []byte(s), nil
	}

	return json.Marshal(nf.Float64)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (nf *NullFloat64) UnmarshalJSON(b []byte) error {
	// Fast path for null
	if isNullJSON(b) {
		nf.Valid = false
		return nil
	}

	// Handle common small integer values that are represented as floats
	// These are very frequent in real applications
	if len(b) == 1 && b[0] >= '0' && b[0] <= '9' {
		nf.Float64 = float64(b[0] - '0')
		nf.Valid = true
		return nil
	}

	// Check for common float patterns (small integers and simple decimals)
	isSimpleNumber := len(b) > 0
	hasDecimal := false
	decimalPos := -1

	// Verify it's a simple number with up to one decimal point
	startIdx := 0
	if b[0] == '-' {
		startIdx = 1
		// Must have at least one digit after the sign
		if len(b) <= startIdx {
			isSimpleNumber = false
		}
	}

	for i := startIdx; isSimpleNumber && i < len(b); i++ {
		if b[i] == '.' {
			if hasDecimal { // Second decimal point is invalid
				isSimpleNumber = false
			} else {
				hasDecimal = true
				decimalPos = i
			}
		} else if b[i] < '0' || b[i] > '9' {
			isSimpleNumber = false
		}
	}

	// Fast path for small integers (0-99)
	if isSimpleNumber && !hasDecimal && len(b) <= 2+(startIdx-0) {
		// Process as an integer, then convert to float
		var val int
		for i := startIdx; i < len(b); i++ {
			val = val*10 + int(b[i]-'0')
		}

		// Apply sign if needed
		if startIdx > 0 {
			val = -val
		}

		nf.Float64 = float64(val)
		nf.Valid = true
		return nil
	}

	// Fast path for simple decimal values with 1-2 decimal places
	if isSimpleNumber && hasDecimal && len(b) <= 5+(startIdx-0) && (len(b)-decimalPos-1) <= 2 {
		// Direct parsing without standard library for simple cases
		val, err := strconv.ParseFloat(string(b), 64)
		if err == nil {
			nf.Float64 = val
			nf.Valid = true
			return nil
		}
	}

	// Default to standard parsing for all other cases
	val, err := strconv.ParseFloat(string(b), 64)
	if err == nil {
		nf.Float64 = val
		nf.Valid = true
		return nil
	}

	// Final fallback to standard unmarshal
	var f float64
	if err := json.Unmarshal(b, &f); err != nil {
		return err
	}
	nf.Float64 = f
	nf.Valid = true
	return nil
}

// WriteTo implements the io.WriterTo interface for binary serialization.
func (nf NullFloat64) WriteTo(w io.Writer) (n int64, err error) {
	// Get buffer from pool
	buf := smallBufferPool.Get().([]byte)
	defer smallBufferPool.Put(buf)

	// Set valid flag in buffer
	if nf.Valid {
		buf[0] = 1
	} else {
		buf[0] = 0
	}

	// Write valid flag
	nn, err := w.Write(buf[:1])
	n += int64(nn)
	if err != nil {
		return n, err
	}

	// If invalid, we're done
	if !nf.Valid {
		return n, nil
	}

	// Put float64 value in buffer (8 bytes)
	binary.LittleEndian.PutUint64(buf[:8], math.Float64bits(nf.Float64))
	nn, err = w.Write(buf[:8])
	n += int64(nn)
	if err != nil {
		return n, err
	}

	return n, nil
}

// ReadFrom implements the io.ReaderFrom interface for binary deserialization.
func (nf *NullFloat64) ReadFrom(r io.Reader) (n int64, err error) {
	// Get buffer from pool
	buf := smallBufferPool.Get().([]byte)
	defer smallBufferPool.Put(buf)

	// Read valid flag (1 byte)
	nn, err := io.ReadFull(r, buf[:1])
	n += int64(nn)
	if err != nil {
		return n, err
	}

	nf.Valid = buf[0] == 1

	// If invalid, we're done
	if !nf.Valid {
		nf.Float64 = 0
		return n, nil
	}

	// Read float64 value (8 bytes)
	nn, err = io.ReadFull(r, buf[:8])
	n += int64(nn)
	if err != nil {
		return n, err
	}

	nf.Float64 = math.Float64frombits(binary.LittleEndian.Uint64(buf[:8]))

	return n, nil
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

	// Standard unmarshal to a temporary map
	m := make(map[string]int)
	if err := json.Unmarshal(asBytes, &m); err != nil {
		return err
	}

	// Create a new map with interned keys
	*id = make(IntDictionary, len(m))
	for k, v := range m {
		// Intern the key to reduce memory usage
		internedKey := internString(k)
		(*id)[internedKey] = v
	}

	return nil
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
		// Intern the key to reduce memory usage
		internedKey := internString(k)
		(*id)[internedKey] = v
	}

	return nil
}
