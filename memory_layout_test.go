package octypes

import (
	"database/sql"
	"fmt"
	"reflect"
	"testing"
	"unsafe"
)

// MemoryLayout returns information about the memory layout of a struct.
func MemoryLayout(t reflect.Type) string {
	if t.Kind() != reflect.Struct {
		return "Not a struct type"
	}

	result := fmt.Sprintf("Struct %s: size=%d, align=%d\n", t.Name(), t.Size(), t.Align())
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		result += fmt.Sprintf("  Field: %s, offset=%d, size=%d, align=%d\n",
			f.Name, f.Offset, f.Type.Size(), f.Type.Align())
	}
	return result
}

// TestMemoryLayout tests the memory layouts of our structs.
func TestMemoryLayout(t *testing.T) {
	// Temporarily enable this test to see memory layouts
	// t.Skip("Skip memory layout test normally")

	// Test current struct memory layouts
	fmt.Println("\n=== NullString ===")
	fmt.Println(MemoryLayout(reflect.TypeOf(NullString{})))

	fmt.Println("\n=== NullInt64 ===")
	fmt.Println(MemoryLayout(reflect.TypeOf(NullInt64{})))

	fmt.Println("\n=== NullBool ===")
	fmt.Println(MemoryLayout(reflect.TypeOf(NullBool{})))

	fmt.Println("\n=== NullFloat64 ===")
	fmt.Println(MemoryLayout(reflect.TypeOf(NullFloat64{})))

	fmt.Println("\n=== CustomTime ===")
	fmt.Println(MemoryLayout(reflect.TypeOf(CustomTime{})))

	// Test sql.Null* memory layouts for comparison
	fmt.Println("\n=== sql.NullString ===")
	fmt.Println(MemoryLayout(reflect.TypeOf(struct{ sql.NullString }{})))

	fmt.Println("\n=== sql.NullInt64 ===")
	fmt.Println(MemoryLayout(reflect.TypeOf(struct{ sql.NullInt64 }{})))

	fmt.Println("\n=== sql.NullBool ===")
	fmt.Println(MemoryLayout(reflect.TypeOf(struct{ sql.NullBool }{})))

	fmt.Println("\n=== sql.NullFloat64 ===")
	fmt.Println(MemoryLayout(reflect.TypeOf(struct{ sql.NullFloat64 }{})))

	// Test type used in benchmark
	type TestStruct struct {
		Name   NullString  `json:"name"`
		Age    NullInt64   `json:"age"`
		Score  NullFloat64 `json:"score"`
		Active NullBool    `json:"active"`
	}

	fmt.Println("\n=== TestStruct ===")
	fmt.Println(MemoryLayout(reflect.TypeOf(TestStruct{})))

	// Try an optimized layout
	type OptimizedTestStruct struct {
		Age    NullInt64   `json:"age"`       // 8-byte aligned
		Score  NullFloat64 `json:"score"`     // 8-byte aligned
		Name   NullString  `json:"name"`      // string (8-byte pointer) + bool
		Active NullBool    `json:"active"`    // bool at the end
	}

	fmt.Println("\n=== OptimizedTestStruct ===")
	fmt.Println(MemoryLayout(reflect.TypeOf(OptimizedTestStruct{})))

	// Define our optimized wrappers
	type OptNullBool struct {
		Valid bool
		Bool  bool
		// 6 bytes padding here
	}

	type OptNullInt64 struct {
		Int64 int64
		Valid bool
		// 7 bytes padding here
	}

	type OptNullFloat64 struct {
		Float64 float64
		Valid   bool
		// 7 bytes padding here
	}

	type OptNullString struct {
		String string // 16 bytes (pointer + len)
		Valid  bool   // 1 byte
		// 7 bytes padding here
	}

	// Calculate total size
	totalSize := int(unsafe.Sizeof(OptNullBool{})) +
		int(unsafe.Sizeof(OptNullInt64{})) +
		int(unsafe.Sizeof(OptNullFloat64{})) +
		int(unsafe.Sizeof(OptNullString{}))

	fmt.Printf("\nTotal size of optimized types: %d bytes\n", totalSize)

	// Check size of a struct that uses these
	type FullyOptimizedStruct struct {
		Int64  OptNullInt64
		Float  OptNullFloat64
		String OptNullString
		Bool   OptNullBool
	}

	fmt.Println("\n=== FullyOptimizedStruct ===")
	fmt.Println(MemoryLayout(reflect.TypeOf(FullyOptimizedStruct{})))
}