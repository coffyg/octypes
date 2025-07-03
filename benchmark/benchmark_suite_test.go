package benchmark

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"
	"time"

	"github.com/coffyg/octypes"
)

// Sample JSON data
var (
	stringJSON      = []byte(`"test string with some content that's long enough to test performance"`)
	nullStringJSON  = []byte(`null`)
	intJSON         = []byte(`12345`)
	floatJSON       = []byte(`123.456`)
	boolJSON        = []byte(`true`)
	complexJSON     = []byte(`{
		"en": "English",
		"fr": "French",
		"de": "German",
		"es": "Spanish",
		"it": "Italian",
		"ru": "Russian",
		"zh": "Chinese",
		"ja": "Japanese",
		"ko": "Korean",
		"ar": "Arabic",
		"hi": "Hindi",
		"pt": "Portuguese",
		"bn": "Bengali",
		"id": "Indonesian",
		"ur": "Urdu",
		"ms": "Malay",
		"tr": "Turkish",
		"nl": "Dutch",
		"th": "Thai",
		"pl": "Polish",
		"sw": "Swahili",
		"vi": "Vietnamese",
		"jv": "Javanese",
		"uk": "Ukrainian",
		"fa": "Persian",
		"ro": "Romanian",
		"el": "Greek"
	}`)
)

// Complex struct for testing
type TestComplexStruct struct {
	ID          octypes.NullInt64     `json:"id"`
	Name        octypes.NullString    `json:"name"`
	Description octypes.NullString    `json:"description"`
	Score       octypes.NullFloat64   `json:"score"`
	IsActive    octypes.NullBool      `json:"is_active"`
	CreatedAt   octypes.CustomTime    `json:"created_at"`
	UpdatedAt   octypes.CustomTime    `json:"updated_at"`
	Tags        octypes.LocalizedText `json:"tags"`
	Counts      octypes.IntDictionary `json:"counts"`
}

// WriteTo implements binary serialization for TestComplexStruct
func (cs TestComplexStruct) WriteTo(w io.Writer) (n int64, err error) {
	var total int64
	var written int64

	written, err = cs.ID.WriteTo(w)
	total += written
	if err != nil {
		return total, err
	}

	written, err = cs.Name.WriteTo(w)
	total += written
	if err != nil {
		return total, err
	}

	written, err = cs.Description.WriteTo(w)
	total += written
	if err != nil {
		return total, err
	}

	written, err = cs.Score.WriteTo(w)
	total += written
	if err != nil {
		return total, err
	}

	written, err = cs.IsActive.WriteTo(w)
	total += written
	if err != nil {
		return total, err
	}

	written, err = cs.CreatedAt.WriteTo(w)
	total += written
	if err != nil {
		return total, err
	}
	
	written, err = cs.UpdatedAt.WriteTo(w)
	total += written
	if err != nil {
		return total, err
	}

	// Serialize the map fields directly to make it easier 
	var buf []byte
	buf, err = json.Marshal(cs.Tags)
	if err != nil {
		return total, err
	}
	written = int64(len(buf))
	total += written
	_, err = w.Write(buf)
	if err != nil {
		return total, err
	}
	
	buf, err = json.Marshal(cs.Counts)
	if err != nil {
		return total, err
	}
	written = int64(len(buf))
	total += written
	_, err = w.Write(buf)
	
	return total, err
}

// ReadFrom implements binary deserialization for TestComplexStruct
func (cs *TestComplexStruct) ReadFrom(r io.Reader) (n int64, err error) {
	var total int64
	var read int64

	read, err = cs.ID.ReadFrom(r)
	total += read
	if err != nil {
		return total, err
	}

	read, err = cs.Name.ReadFrom(r)
	total += read
	if err != nil {
		return total, err
	}

	read, err = cs.Description.ReadFrom(r)
	total += read
	if err != nil {
		return total, err
	}

	read, err = cs.Score.ReadFrom(r)
	total += read
	if err != nil {
		return total, err
	}

	read, err = cs.IsActive.ReadFrom(r)
	total += read
	if err != nil {
		return total, err
	}

	read, err = cs.CreatedAt.ReadFrom(r)
	total += read
	if err != nil {
		return total, err
	}
	
	read, err = cs.UpdatedAt.ReadFrom(r)
	total += read
	if err != nil {
		return total, err
	}

	// For maps, we'll use a simpler approach since they don't have built-in binary serialization
	var buf [1024]byte
	n, err := r.Read(buf[:])
	if err != nil && err != io.EOF {
		return total, err
	}
	total += int64(n)
	
	err = json.Unmarshal(buf[:n], &cs.Tags)
	if err != nil {
		return total, err
	}
	
	n, err = r.Read(buf[:])
	if err != nil && err != io.EOF {
		return total, err
	}
	total += int64(n)
	
	err = json.Unmarshal(buf[:n], &cs.Counts)
	
	return total, err
}

// Benchmark NullString MarshalJSON
func BenchmarkNullStringMarshalJSON(b *testing.B) {
	ns := octypes.NewNullString("test string")
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(ns)
	}
}

// Benchmark NullString UnmarshalJSON
func BenchmarkNullStringUnmarshalJSON(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var ns octypes.NullString
		_ = json.Unmarshal(stringJSON, &ns)
	}
}

// Benchmark NullString UnmarshalJSON with null
func BenchmarkNullStringUnmarshalNullJSON(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var ns octypes.NullString
		_ = json.Unmarshal(nullStringJSON, &ns)
	}
}

// Benchmark NullInt64 MarshalJSON
func BenchmarkNullInt64MarshalJSON(b *testing.B) {
	ni := octypes.NewNullInt64(12345)
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(ni)
	}
}

// Benchmark NullInt64 UnmarshalJSON
func BenchmarkNullInt64UnmarshalJSON(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var ni octypes.NullInt64
		_ = json.Unmarshal(intJSON, &ni)
	}
}

// Benchmark NullFloat64 MarshalJSON
func BenchmarkNullFloat64MarshalJSON(b *testing.B) {
	nf := octypes.NewNullFloat64(123.456)
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(nf)
	}
}

// Benchmark NullFloat64 UnmarshalJSON
func BenchmarkNullFloat64UnmarshalJSON(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var nf octypes.NullFloat64
		_ = json.Unmarshal(floatJSON, &nf)
	}
}

// Benchmark NullBool MarshalJSON
func BenchmarkNullBoolMarshalJSON(b *testing.B) {
	nb := octypes.NewNullBool(true)
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(nb)
	}
}

// Benchmark NullBool UnmarshalJSON
func BenchmarkNullBoolUnmarshalJSON(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var nb octypes.NullBool
		_ = json.Unmarshal(boolJSON, &nb)
	}
}

// Benchmark CustomTime MarshalJSON
func BenchmarkCustomTimeMarshalJSON(b *testing.B) {
	ct := octypes.NewCustomTime(time.Now())
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(ct)
	}
}

// Benchmark LocalizedText UnmarshalJSON
func BenchmarkLocalizedTextUnmarshalJSON(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var lt octypes.LocalizedText
		_ = json.Unmarshal(complexJSON, &lt)
	}
}

// Benchmark IntDictionary UnmarshalJSON
func BenchmarkIntDictionaryUnmarshalJSON(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var id octypes.IntDictionary
		_ = json.Unmarshal(complexJSON, &id)
	}
}

// Benchmark complex struct
func BenchmarkComplexStructMarshalJSON(b *testing.B) {
	complexStruct := TestComplexStruct{
		ID:          *octypes.NewNullInt64(12345),
		Name:        *octypes.NewNullString("Test Name"),
		Description: *octypes.NewNullString("This is a test description with some lengthy content to test performance with longer strings"),
		Score:       *octypes.NewNullFloat64(98.76),
		IsActive:    *octypes.NewNullBool(true),
		CreatedAt:   *octypes.NewCustomTime(time.Now().Add(-24 * time.Hour)),
		UpdatedAt:   *octypes.NewCustomTime(time.Now()),
		Tags: octypes.LocalizedText{
			"en": "English",
			"fr": "French",
			"de": "German",
			"es": "Spanish",
			"it": "Italian",
		},
		Counts: octypes.IntDictionary{
			"apples": 5,
			"oranges": 10,
			"bananas": 7,
			"grapes": 20,
		},
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(complexStruct)
	}
}

// Benchmark binary serialization for complex struct
func BenchmarkComplexStructBinary(b *testing.B) {
	complexStruct := TestComplexStruct{
		ID:          *octypes.NewNullInt64(12345),
		Name:        *octypes.NewNullString("Test Name"),
		Description: *octypes.NewNullString("This is a test description with some lengthy content to test performance with longer strings"),
		Score:       *octypes.NewNullFloat64(98.76),
		IsActive:    *octypes.NewNullBool(true),
		CreatedAt:   *octypes.NewCustomTime(time.Now().Add(-24 * time.Hour)),
		UpdatedAt:   *octypes.NewCustomTime(time.Now()),
		Tags: octypes.LocalizedText{
			"en": "English",
			"fr": "French",
			"de": "German",
			"es": "Spanish",
			"it": "Italian",
		},
		Counts: octypes.IntDictionary{
			"apples": 5,
			"oranges": 10,
			"bananas": 7,
			"grapes": 20,
		},
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		_, _ = complexStruct.WriteTo(&buf)
	}
}

// Benchmark for memory layout impact (large array allocation)
func BenchmarkMemoryLayoutImpact(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		items := make([]TestComplexStruct, 1000)
		for j := 0; j < 1000; j++ {
			items[j] = TestComplexStruct{
				ID:          *octypes.NewNullInt64(int64(j)),
				Name:        *octypes.NewNullString("Name " + string(j)),
				Description: *octypes.NewNullString("Description " + string(j)),
				Score:       *octypes.NewNullFloat64(float64(j) * 1.5),
				IsActive:    *octypes.NewNullBool(j%2 == 0),
				CreatedAt:   *octypes.NewCustomTime(time.Now()),
				UpdatedAt:   *octypes.NewCustomTime(time.Now()),
			}
		}
		_ = items
	}
}

// Benchmark standard struct JSON operations in batch
func BenchmarkBatchJSONOperations(b *testing.B) {
	type User struct {
		ID       octypes.NullInt64   `json:"id"`
		Name     octypes.NullString  `json:"name"`
		Age      octypes.NullInt64   `json:"age"`
		Active   octypes.NullBool    `json:"active"`
		Score    octypes.NullFloat64 `json:"score"`
		JoinDate octypes.CustomTime  `json:"join_date"`
	}
	
	// Create a batch of users
	users := make([]User, 100)
	for i := 0; i < 100; i++ {
		users[i] = User{
			ID:       *octypes.NewNullInt64(int64(i + 1000)),
			Name:     *octypes.NewNullString("User " + string(i)),
			Age:      *octypes.NewNullInt64(20 + int64(i%30)),
			Active:   *octypes.NewNullBool(i%3 == 0),
			Score:    *octypes.NewNullFloat64(75.0 + float64(i)/10.0),
			JoinDate: *octypes.NewCustomTime(time.Now().AddDate(0, -(i % 12), 0)),
		}
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// Marshal each user
		for _, user := range users {
			jsonData, _ := json.Marshal(user)
			
			// Then unmarshal back
			var u2 User
			_ = json.Unmarshal(jsonData, &u2)
		}
	}
}

// Benchmark for string interning in maps
func BenchmarkStringInterning(b *testing.B) {
	// Create a map with lots of strings
	m := octypes.LocalizedText{}
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = json.Unmarshal(complexJSON, &m)
		
		// Access operations
		for j := 0; j < 1000; j++ {
			_ = m["en"]
			_ = m["fr"]
			_ = m["de"]
			_ = m["es"]
			_ = m["it"]
		}
	}
}

// Benchmark for checking all NULL types
func BenchmarkNullTypesHandling(b *testing.B) {
	nullData := []byte(`{
		"id": null,
		"name": null,
		"description": null,
		"score": null,
		"is_active": null,
		"created_at": null,
		"updated_at": null,
		"tags": null,
		"counts": null
	}`)
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var cs TestComplexStruct
		_ = json.Unmarshal(nullData, &cs)
	}
}