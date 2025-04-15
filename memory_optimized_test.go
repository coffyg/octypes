package octypes

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
	"time"
	"unsafe"
)

// TestOptimizedTypesSizes tests the sizes of the internal optimized types.
func TestOptimizedTypesSizes(t *testing.T) {
	// Get actual sizes
	nullStringSize := unsafe.Sizeof(NullString{})
	optimizedNullStringSize := unsafe.Sizeof(OptimizedNullString{})
	nullInt64Size := unsafe.Sizeof(NullInt64{})
	optimizedNullInt64Size := unsafe.Sizeof(OptimizedNullInt64{})
	nullBoolSize := unsafe.Sizeof(NullBool{})
	optimizedNullBoolSize := unsafe.Sizeof(OptimizedNullBool{})
	nullFloat64Size := unsafe.Sizeof(NullFloat64{})
	optimizedNullFloat64Size := unsafe.Sizeof(OptimizedNullFloat64{})
	customTimeSize := unsafe.Sizeof(CustomTime{})
	optimizedCustomTimeSize := unsafe.Sizeof(OptimizedCustomTime{})
	
	// Log sizes in a tabular format
	t.Logf("Memory Layout Comparison Table:")
	t.Logf("%-22s | %-14s | %-14s | %s", "Type", "Original Size", "Optimized Size", "Difference")
	t.Logf("%-22s | %-14s | %-14s | %s", "--------------------", "--------------", "--------------", "----------")
	t.Logf("%-22s | %-14d | %-14d | %d bytes", "NullString", nullStringSize, optimizedNullStringSize, int(nullStringSize)-int(optimizedNullStringSize))
	t.Logf("%-22s | %-14d | %-14d | %d bytes", "NullInt64", nullInt64Size, optimizedNullInt64Size, int(nullInt64Size)-int(optimizedNullInt64Size))
	t.Logf("%-22s | %-14d | %-14d | %d bytes", "NullBool", nullBoolSize, optimizedNullBoolSize, int(nullBoolSize)-int(optimizedNullBoolSize))
	t.Logf("%-22s | %-14d | %-14d | %d bytes", "NullFloat64", nullFloat64Size, optimizedNullFloat64Size, int(nullFloat64Size)-int(optimizedNullFloat64Size))
	t.Logf("%-22s | %-14d | %-14d | %d bytes", "CustomTime", customTimeSize, optimizedCustomTimeSize, int(customTimeSize)-int(optimizedCustomTimeSize))
	
	// Create total for all types
	totalOriginal := nullStringSize + nullInt64Size + nullBoolSize + nullFloat64Size + customTimeSize
	totalOptimized := optimizedNullStringSize + optimizedNullInt64Size + optimizedNullBoolSize + optimizedNullFloat64Size + optimizedCustomTimeSize
	t.Logf("%-22s | %-14d | %-14d | %d bytes", "TOTAL", totalOriginal, totalOptimized, int(totalOriginal)-int(totalOptimized))
	
	// Check that optimized types are not larger than original types
	if optimizedNullStringSize > nullStringSize {
		t.Errorf("OptimizedNullString size %d is larger than NullString size %d", 
			optimizedNullStringSize, nullStringSize)
	}
	
	if optimizedNullInt64Size > nullInt64Size {
		t.Errorf("OptimizedNullInt64 size %d is larger than NullInt64 size %d", 
			optimizedNullInt64Size, nullInt64Size)
	}
	
	if optimizedNullBoolSize > nullBoolSize {
		t.Errorf("OptimizedNullBool size %d is larger than NullBool size %d", 
			optimizedNullBoolSize, nullBoolSize)
	}
	
	if optimizedNullFloat64Size > nullFloat64Size {
		t.Errorf("OptimizedNullFloat64 size %d is larger than NullFloat64 size %d", 
			optimizedNullFloat64Size, nullFloat64Size)
	}
	
	if optimizedCustomTimeSize > customTimeSize {
		t.Errorf("OptimizedCustomTime size %d is larger than CustomTime size %d", 
			optimizedCustomTimeSize, customTimeSize)
	}
}

// Compare memory layout of regular vs optimized structs
func TestStructMemoryLayout(t *testing.T) {
	// Regular struct type
	type RegularStruct struct {
		Name   NullString  `json:"name"`
		Age    NullInt64   `json:"age"`
		Score  NullFloat64 `json:"score"`
		Active NullBool    `json:"active"`
	}

	// Internal optimized struct type with better field ordering
	type OptStruct struct {
		Age    OptimizedNullInt64   `json:"age"`    // 8-byte aligned
		Score  OptimizedNullFloat64 `json:"score"`  // 8-byte aligned
		Name   OptimizedNullString  `json:"name"`   // 16 bytes + bool
		Active OptimizedNullBool    `json:"active"` // bool at the end
	}

	// Check sizes
	regularSize := unsafe.Sizeof(RegularStruct{})
	optimizedSize := unsafe.Sizeof(OptStruct{})

	// Print table for comparison
	t.Logf("\nStruct Memory Layout Comparison:")
	t.Logf("%-22s | %-14s | %-14s | %s", "Struct", "Regular", "Optimized", "Difference")
	t.Logf("%-22s | %-14s | %-14s | %s", "--------------------", "--------------", "--------------", "----------")
	t.Logf("%-22s | %-14d | %-14d | %d bytes", "Test Struct", regularSize, optimizedSize, int(regularSize)-int(optimizedSize))
	
	// Create a slice with typical data to check memory usage more realistically
	regularSlice := make([]RegularStruct, 1000)
	optimizedSlice := make([]OptStruct, 1000)
	
	regularSliceSize := unsafe.Sizeof(regularSlice) + uintptr(len(regularSlice))*regularSize
	optimizedSliceSize := unsafe.Sizeof(optimizedSlice) + uintptr(len(optimizedSlice))*optimizedSize
	
	t.Logf("%-22s | %-14d | %-14d | %d bytes", "Slice (1000 items)", regularSliceSize, optimizedSliceSize, int(regularSliceSize)-int(optimizedSliceSize))
}

// TestOptimizedComplexStructSize tests the size of the internal optimized complex struct.
func TestOptimizedComplexStructSize(t *testing.T) {
	// Standard complex struct from benchmarks
	type ComplexStruct struct {
		ID          NullInt64   `json:"id"`
		Name        NullString  `json:"name"`
		Description NullString  `json:"description"`
		Score       NullFloat64 `json:"score"`
		IsActive    NullBool    `json:"is_active"`
		CreatedAt   CustomTime  `json:"created_at"`
		UpdatedAt   CustomTime  `json:"updated_at"`
	}

	// Get sizes
	stdSize := unsafe.Sizeof(ComplexStruct{})
	optSize := unsafe.Sizeof(OptimizedComplexStruct{})

	// Print comparison table
	t.Logf("\nComplex Struct Memory Layout Comparison:")
	t.Logf("%-22s | %-14s | %-14s | %s", "Struct", "Regular", "Optimized", "Difference")
	t.Logf("%-22s | %-14s | %-14s | %s", "--------------------", "--------------", "--------------", "----------")
	t.Logf("%-22s | %-14d | %-14d | %d bytes", "ComplexStruct", stdSize, optSize, int(stdSize)-int(optSize))
	
	// Calculate actual memory usage in a realistic scenario (1000 item slice)
	stdSliceSize := unsafe.Sizeof(make([]ComplexStruct, 0)) + 1000*stdSize
	optSliceSize := unsafe.Sizeof(make([]OptimizedComplexStruct, 0)) + 1000*optSize
	
	bytesPerThousand := int(stdSliceSize - optSliceSize)
	t.Logf("%-22s | %-14d | %-14d | %d bytes", "1000 items", stdSliceSize, optSliceSize, bytesPerThousand)
	t.Logf("Memory savings for 1000 items: %d bytes (%0.2f KB, %0.2f MB)", 
		bytesPerThousand, float64(bytesPerThousand)/1024, float64(bytesPerThousand)/(1024*1024))
	
	// Check field alignment and padding
	t.Logf("\nComplexStruct Field Layout:")
	t.Logf("%-15s | %-8s | %s", "Field", "Size", "Offset")
	t.Logf("%-15s | %-8s | %s", "---------------", "--------", "--------")
	
	var cs ComplexStruct
	csType := reflect.TypeOf(cs)
	for i := 0; i < csType.NumField(); i++ {
		field := csType.Field(i)
		t.Logf("%-15s | %-8d | %d", field.Name, field.Type.Size(), field.Offset)
	}
	
	t.Logf("\nOptimizedComplexStruct Field Layout:")
	t.Logf("%-15s | %-8s | %s", "Field", "Size", "Offset")
	t.Logf("%-15s | %-8s | %s", "---------------", "--------", "--------")
	
	var ocs OptimizedComplexStruct
	ocsType := reflect.TypeOf(ocs)
	for i := 0; i < ocsType.NumField(); i++ {
		field := ocsType.Field(i)
		t.Logf("%-15s | %-8d | %d", field.Name, field.Type.Size(), field.Offset)
	}

	// Check if optimized is not larger
	if optSize > stdSize {
		t.Errorf("OptimizedComplexStruct size %d is larger than ComplexStruct size %d", 
			optSize, stdSize)
	}
}

// TestOptimizedNullString tests the optimized NullString type.
func TestOptimizedNullString(t *testing.T) {
	// Test constructor with non-empty string
	ns := NewOptimizedNullString("hello")
	if !ns.Valid || ns.String != "hello" {
		t.Errorf("Expected Valid true and String 'hello', got Valid %v and String '%s'", ns.Valid, ns.String)
	}

	// Test constructor with empty string
	ns = NewOptimizedNullString("")
	if ns.Valid {
		t.Errorf("Expected Valid false for empty string")
	}

	// Test JSON marshalling
	jsonData, err := json.Marshal(ns)
	if err != nil {
		t.Errorf("Error marshalling OptimizedNullString: %v", err)
	}
	if string(jsonData) != "null" {
		t.Errorf("Expected JSON 'null', got %s", jsonData)
	}

	// Test JSON unmarshalling
	err = json.Unmarshal([]byte(`"world"`), ns)
	if err != nil {
		t.Errorf("Error unmarshalling OptimizedNullString: %v", err)
	}
	if !ns.Valid || ns.String != "world" {
		t.Errorf("Expected Valid true and String 'world', got Valid %v and String '%s'", ns.Valid, ns.String)
	}

	// Test Value
	val, err := ns.Value()
	if err != nil {
		t.Errorf("Error getting Value from OptimizedNullString: %v", err)
	}
	if val != "world" {
		t.Errorf("Expected Value 'world', got '%v'", val)
	}
}

// TestOptimizedNullInt64 tests the optimized NullInt64 type.
func TestOptimizedNullInt64(t *testing.T) {
	// Test constructor with int64
	ni := NewOptimizedNullInt64(42)
	if !ni.Valid || ni.Int64 != 42 {
		t.Errorf("Expected Valid true and Int64 42, got Valid %v and Int64 %d", ni.Valid, ni.Int64)
	}

	// Test JSON marshalling
	jsonData, err := json.Marshal(ni)
	if err != nil {
		t.Errorf("Error marshalling OptimizedNullInt64: %v", err)
	}
	if string(jsonData) != "42" {
		t.Errorf("Expected JSON '42', got %s", jsonData)
	}

	// Test JSON unmarshalling
	err = json.Unmarshal([]byte(`100`), ni)
	if err != nil {
		t.Errorf("Error unmarshalling OptimizedNullInt64: %v", err)
	}
	if !ni.Valid || ni.Int64 != 100 {
		t.Errorf("Expected Valid true and Int64 100, got Valid %v and Int64 %d", ni.Valid, ni.Int64)
	}

	// Test Value
	val, err := ni.Value()
	if err != nil {
		t.Errorf("Error getting Value from OptimizedNullInt64: %v", err)
	}
	if val != int64(100) {
		t.Errorf("Expected Value 100, got '%v'", val)
	}
}

// TestCompatibility tests that optimized types are compatible with the original types.
func TestCompatibility(t *testing.T) {
	// Test NullString and OptimizedNullString
	{
		original := NewNullString("test")
		optimized := NewOptimizedNullString("test")
		
		originalJSON, _ := json.Marshal(original)
		optimizedJSON, _ := json.Marshal(optimized)
		
		if string(originalJSON) != string(optimizedJSON) {
			t.Errorf("JSON mismatch: NullString %s vs OptimizedNullString %s", 
				originalJSON, optimizedJSON)
		}
		
		// Test null values
		originalNull := NewNullStringNull()
		optimizedNull := NewOptimizedNullStringNull()
		
		originalNullJSON, _ := json.Marshal(originalNull)
		optimizedNullJSON, _ := json.Marshal(optimizedNull)
		
		if string(originalNullJSON) != string(optimizedNullJSON) {
			t.Errorf("JSON null mismatch: NullString %s vs OptimizedNullString %s", 
				originalNullJSON, optimizedNullJSON)
		}
	}
	
	// Test NullInt64 and OptimizedNullInt64
	{
		original := NewNullInt64(42)
		optimized := NewOptimizedNullInt64(42)
		
		originalJSON, _ := json.Marshal(original)
		optimizedJSON, _ := json.Marshal(optimized)
		
		if string(originalJSON) != string(optimizedJSON) {
			t.Errorf("JSON mismatch: NullInt64 %s vs OptimizedNullInt64 %s", 
				originalJSON, optimizedJSON)
		}
	}
	
	// Test NullBool and OptimizedNullBool
	{
		original := NewNullBool(true)
		optimized := NewOptimizedNullBool(true)
		
		originalJSON, _ := json.Marshal(original)
		optimizedJSON, _ := json.Marshal(optimized)
		
		if string(originalJSON) != string(optimizedJSON) {
			t.Errorf("JSON mismatch: NullBool %s vs OptimizedNullBool %s", 
				originalJSON, optimizedJSON)
		}
	}
	
	// Test NullFloat64 and OptimizedNullFloat64
	{
		original := NewNullFloat64(3.14)
		optimized := NewOptimizedNullFloat64(3.14)
		
		originalJSON, _ := json.Marshal(original)
		optimizedJSON, _ := json.Marshal(optimized)
		
		if string(originalJSON) != string(optimizedJSON) {
			t.Errorf("JSON mismatch: NullFloat64 %s vs OptimizedNullFloat64 %s", 
				originalJSON, optimizedJSON)
		}
	}
	
	// Test CustomTime and OptimizedCustomTime
	{
		now := time.Now()
		original := NewCustomTime(now)
		optimized := NewOptimizedCustomTime(now)
		
		originalJSON, _ := json.Marshal(original)
		optimizedJSON, _ := json.Marshal(optimized)
		
		// Parse both JSON to verify they contain the same data
		var origMap map[string]interface{}
		var optMap map[string]interface{}
		
		json.Unmarshal(originalJSON, &origMap)
		json.Unmarshal(optimizedJSON, &optMap)
		
		// Check ISO time field
		if origMap["iso"] != optMap["iso"] {
			t.Errorf("ISO time mismatch: %v vs %v", origMap["iso"], optMap["iso"])
		}
	}
}

// TestBinarySerialization tests binary serialization and deserialization.
func TestBinarySerialization(t *testing.T) {
	// Test NullString and OptimizedNullString
	{
		original := NewNullString("test")
		optimized := NewOptimizedNullString("test")
		
		// Serialize
		var origBuf, optBuf bytes.Buffer
		original.WriteTo(&origBuf)
		optimized.WriteTo(&optBuf)
		
		// Deserialize
		var origResult NullString
		var optResult OptimizedNullString
		origResult.ReadFrom(bytes.NewReader(origBuf.Bytes()))
		optResult.ReadFrom(bytes.NewReader(optBuf.Bytes()))
		
		// Verify
		if original.Valid != origResult.Valid || original.String != origResult.String {
			t.Errorf("NullString binary serialization failed")
		}
		
		if optimized.Valid != optResult.Valid || optimized.String != optResult.String {
			t.Errorf("OptimizedNullString binary serialization failed")
		}
	}
	
	// Test NullInt64 and OptimizedNullInt64
	{
		original := NewNullInt64(42)
		optimized := NewOptimizedNullInt64(42)
		
		// Serialize
		var origBuf, optBuf bytes.Buffer
		original.WriteTo(&origBuf)
		optimized.WriteTo(&optBuf)
		
		// Deserialize
		var origResult NullInt64
		var optResult OptimizedNullInt64
		origResult.ReadFrom(bytes.NewReader(origBuf.Bytes()))
		optResult.ReadFrom(bytes.NewReader(optBuf.Bytes()))
		
		// Verify
		if original.Valid != origResult.Valid || original.Int64 != origResult.Int64 {
			t.Errorf("NullInt64 binary serialization failed")
		}
		
		if optimized.Valid != optResult.Valid || optimized.Int64 != optResult.Int64 {
			t.Errorf("OptimizedNullInt64 binary serialization failed")
		}
	}
}

// Benchmarks for OptimizedComplexStruct
func BenchmarkOptimizedComplexStructJSON(b *testing.B) {
	cs := OptimizedComplexStruct{
		Score:       *NewOptimizedNullFloat64(98.76),
		Age:         *NewOptimizedNullInt64(12345),
		CreatedAt:   *NewOptimizedCustomTime(time.Now().Add(-24 * time.Hour)),
		UpdatedAt:   *NewOptimizedCustomTime(time.Now()),
		Name:        *NewOptimizedNullString("Test Name"),
		Description: *NewOptimizedNullString("This is a test description with some more text"),
		IsActive:    *NewOptimizedNullBool(true),
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(cs)
	}
}

func BenchmarkOptimizedComplexStructBinary(b *testing.B) {
	cs := OptimizedComplexStruct{
		Score:       *NewOptimizedNullFloat64(98.76),
		Age:         *NewOptimizedNullInt64(12345),
		CreatedAt:   *NewOptimizedCustomTime(time.Now().Add(-24 * time.Hour)),
		UpdatedAt:   *NewOptimizedCustomTime(time.Now()),
		Name:        *NewOptimizedNullString("Test Name"),
		Description: *NewOptimizedNullString("This is a test description with some more text"),
		IsActive:    *NewOptimizedNullBool(true),
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		_, _ = cs.WriteTo(&buf)
	}
}

func BenchmarkOptimizedComplexStructFromJSON(b *testing.B) {
	cs := OptimizedComplexStruct{
		Score:       *NewOptimizedNullFloat64(98.76),
		Age:         *NewOptimizedNullInt64(12345),
		CreatedAt:   *NewOptimizedCustomTime(time.Now().Add(-24 * time.Hour)),
		UpdatedAt:   *NewOptimizedCustomTime(time.Now()),
		Name:        *NewOptimizedNullString("Test Name"),
		Description: *NewOptimizedNullString("This is a test description with some more text"),
		IsActive:    *NewOptimizedNullBool(true),
	}
	
	jsonData, _ := json.Marshal(cs)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var cs2 OptimizedComplexStruct
		_ = json.Unmarshal(jsonData, &cs2)
	}
}

func BenchmarkOptimizedComplexStructFromBinary(b *testing.B) {
	cs := OptimizedComplexStruct{
		Score:       *NewOptimizedNullFloat64(98.76),
		Age:         *NewOptimizedNullInt64(12345),
		CreatedAt:   *NewOptimizedCustomTime(time.Now().Add(-24 * time.Hour)),
		UpdatedAt:   *NewOptimizedCustomTime(time.Now()),
		Name:        *NewOptimizedNullString("Test Name"),
		Description: *NewOptimizedNullString("This is a test description with some more text"),
		IsActive:    *NewOptimizedNullBool(true),
	}
	
	var buf bytes.Buffer
	_, _ = cs.WriteTo(&buf)
	data := buf.Bytes()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var cs2 OptimizedComplexStruct
		reader := bytes.NewReader(data)
		_, _ = cs2.ReadFrom(reader)
	}
}

// Individual type benchmarks for optimized types
func BenchmarkOptimizedNullStringJSON(b *testing.B) {
	ns := NewOptimizedNullString("test")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(ns)
	}
}

func BenchmarkOptimizedNullInt64JSON(b *testing.B) {
	ni := NewOptimizedNullInt64(42)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(ni)
	}
}

func BenchmarkOptimizedNullBoolJSON(b *testing.B) {
	nb := NewOptimizedNullBool(true)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(nb)
	}
}

func BenchmarkOptimizedNullFloat64JSON(b *testing.B) {
	nf := NewOptimizedNullFloat64(3.14)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(nf)
	}
}

func BenchmarkOptimizedCustomTimeJSON(b *testing.B) {
	now := time.Now()
	ct := NewOptimizedCustomTime(now)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(ct)
	}
}

// Test nulls
func BenchmarkOptimizedNullStringNullJSON(b *testing.B) {
	ns := NewOptimizedNullStringNull()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(ns)
	}
}

func BenchmarkOptimizedNullInt64NullJSON(b *testing.B) {
	ni := NewOptimizedNullInt64Null()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(ni)
	}
}

func BenchmarkOptimizedNullBoolNullJSON(b *testing.B) {
	nb := NewOptimizedNullBoolNull()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(nb)
	}
}

func BenchmarkOptimizedNullFloat64NullJSON(b *testing.B) {
	nf := NewOptimizedNullFloat64Null()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(nf)
	}
}

// Benchmarks for JSON Unmarshaling
func BenchmarkOptimizedNullStringUnmarshalJSON(b *testing.B) {
	data := []byte(`"test string"`)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var ns OptimizedNullString
		_ = json.Unmarshal(data, &ns)
	}
}

func BenchmarkOptimizedNullStringUnmarshalNullJSON(b *testing.B) {
	data := []byte(`null`)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var ns OptimizedNullString
		_ = json.Unmarshal(data, &ns)
	}
}

func BenchmarkOptimizedNullInt64UnmarshalJSON(b *testing.B) {
	data := []byte(`42`)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var ni OptimizedNullInt64
		_ = json.Unmarshal(data, &ni)
	}
}

func BenchmarkOptimizedNullBoolUnmarshalJSON(b *testing.B) {
	data := []byte(`true`)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var nb OptimizedNullBool
		_ = json.Unmarshal(data, &nb)
	}
}

func BenchmarkOptimizedNullFloat64UnmarshalJSON(b *testing.B) {
	data := []byte(`3.14159`)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var nf OptimizedNullFloat64
		_ = json.Unmarshal(data, &nf)
	}
}

// Binary encoding benchmarks
func BenchmarkOptimizedNullStringBinary(b *testing.B) {
	ns := NewOptimizedNullString("test string")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		_, _ = ns.WriteTo(&buf)
	}
}

func BenchmarkOptimizedNullInt64Binary(b *testing.B) {
	ni := NewOptimizedNullInt64(42)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		_, _ = ni.WriteTo(&buf)
	}
}

func BenchmarkOptimizedNullBoolBinary(b *testing.B) {
	nb := NewOptimizedNullBool(true)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		_, _ = nb.WriteTo(&buf)
	}
}

func BenchmarkOptimizedNullFloat64Binary(b *testing.B) {
	nf := NewOptimizedNullFloat64(3.14)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		_, _ = nf.WriteTo(&buf)
	}
}

func BenchmarkOptimizedCustomTimeBinary(b *testing.B) {
	ct := NewOptimizedCustomTime(time.Now())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		_, _ = ct.WriteTo(&buf)
	}
}

// Binary decoding benchmarks
func BenchmarkOptimizedNullStringFromBinary(b *testing.B) {
	ns := NewOptimizedNullString("test string")
	var buf bytes.Buffer
	_, _ = ns.WriteTo(&buf)
	data := buf.Bytes()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var ns2 OptimizedNullString
		reader := bytes.NewReader(data)
		_, _ = ns2.ReadFrom(reader)
	}
}

func BenchmarkOptimizedNullInt64FromBinary(b *testing.B) {
	ni := NewOptimizedNullInt64(42)
	var buf bytes.Buffer
	_, _ = ni.WriteTo(&buf)
	data := buf.Bytes()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var ni2 OptimizedNullInt64
		reader := bytes.NewReader(data)
		_, _ = ni2.ReadFrom(reader)
	}
}

func BenchmarkOptimizedNullBoolFromBinary(b *testing.B) {
	nb := NewOptimizedNullBool(true)
	var buf bytes.Buffer
	_, _ = nb.WriteTo(&buf)
	data := buf.Bytes()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var nb2 OptimizedNullBool
		reader := bytes.NewReader(data)
		_, _ = nb2.ReadFrom(reader)
	}
}

func BenchmarkOptimizedNullFloat64FromBinary(b *testing.B) {
	nf := NewOptimizedNullFloat64(3.14)
	var buf bytes.Buffer
	_, _ = nf.WriteTo(&buf)
	data := buf.Bytes()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var nf2 OptimizedNullFloat64
		reader := bytes.NewReader(data)
		_, _ = nf2.ReadFrom(reader)
	}
}

func BenchmarkOptimizedCustomTimeFromBinary(b *testing.B) {
	ct := NewOptimizedCustomTime(time.Now())
	var buf bytes.Buffer
	_, _ = ct.WriteTo(&buf)
	data := buf.Bytes()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var ct2 OptimizedCustomTime
		reader := bytes.NewReader(data)
		_, _ = ct2.ReadFrom(reader)
	}
}