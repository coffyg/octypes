// types_test.go
package octypes

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"strconv"
	"testing"
	"time"
)

// Original implementations for benchmark comparison
type OriginalNullString struct {
	sql.NullString
}

func (ns OriginalNullString) MarshalJSON() ([]byte, error) {
	if ns.Valid {
		return json.Marshal(ns.String)
	}
	return json.Marshal(nil)
}

type OriginalNullInt64 struct {
	sql.NullInt64
}

func (ni OriginalNullInt64) MarshalJSON() ([]byte, error) {
	if ni.Valid {
		return json.Marshal(ni.Int64)
	}
	return json.Marshal(nil)
}

type OriginalNullBool struct {
	sql.NullBool
}

func (nb OriginalNullBool) MarshalJSON() ([]byte, error) {
	if nb.Valid {
		return json.Marshal(nb.Bool)
	}
	return json.Marshal(nil)
}

type OriginalNullFloat64 struct {
	sql.NullFloat64
}

func (nf OriginalNullFloat64) MarshalJSON() ([]byte, error) {
	if nf.Valid {
		return json.Marshal(nf.Float64)
	}
	return json.Marshal(nil)
}

// Regular tests begin here
func TestNullString(t *testing.T) {
	// Test constructor with non-empty string
	ns := NewNullString("hello")
	if !ns.Valid || ns.String != "hello" {
		t.Errorf("Expected Valid true and String 'hello', got Valid %v and String '%s'", ns.Valid, ns.String)
	}

	// Test constructor with empty string
	ns = NewNullString("")
	if ns.Valid {
		t.Errorf("Expected Valid false for empty string")
	}

	// Test JSON marshalling
	jsonData, err := json.Marshal(ns)
	if err != nil {
		t.Errorf("Error marshalling NullString: %v", err)
	}
	if string(jsonData) != "null" {
		t.Errorf("Expected JSON 'null', got %s", jsonData)
	}

	// Test JSON unmarshalling
	err = json.Unmarshal([]byte(`"world"`), ns)
	if err != nil {
		t.Errorf("Error unmarshalling NullString: %v", err)
	}
	if !ns.Valid || ns.String != "world" {
		t.Errorf("Expected Valid true and String 'world', got Valid %v and String '%s'", ns.Valid, ns.String)
	}

	// Test Scan
	err = ns.Scan("scan test")
	if err != nil {
		t.Errorf("Error scanning NullString: %v", err)
	}
	if !ns.Valid || ns.String != "scan test" {
		t.Errorf("Expected Valid true and String 'scan test', got Valid %v and String '%s'", ns.Valid, ns.String)
	}

	// Test Value
	val, err := ns.Value()
	if err != nil {
		t.Errorf("Error getting Value from NullString: %v", err)
	}
	if val != "scan test" {
		t.Errorf("Expected Value 'scan test', got '%v'", val)
	}
}

func TestNullInt64(t *testing.T) {
	// Test constructor with int64
	ni := NewNullInt64(42)
	if !ni.Valid || ni.Int64 != 42 {
		t.Errorf("Expected Valid true and Int64 42, got Valid %v and Int64 %d", ni.Valid, ni.Int64)
	}

	// Test constructor with empty string
	ni = NewNullInt64FromString("")
	if ni.Valid {
		t.Errorf("Expected Valid false for empty string")
	}

	// Test constructor with valid string
	ni = NewNullInt64FromString("100")
	if !ni.Valid || ni.Int64 != 100 {
		t.Errorf("Expected Valid true and Int64 100, got Valid %v and Int64 %d", ni.Valid, ni.Int64)
	}

	// Test JSON marshalling
	jsonData, err := json.Marshal(ni)
	if err != nil {
		t.Errorf("Error marshalling NullInt64: %v", err)
	}
	if string(jsonData) != "100" {
		t.Errorf("Expected JSON '100', got %s", jsonData)
	}

	// Test JSON unmarshalling
	err = json.Unmarshal([]byte(`200`), ni)
	if err != nil {
		t.Errorf("Error unmarshalling NullInt64: %v", err)
	}
	if !ni.Valid || ni.Int64 != 200 {
		t.Errorf("Expected Valid true and Int64 200, got Valid %v and Int64 %d", ni.Valid, ni.Int64)
	}

	// Test Scan
	err = ni.Scan(int64(300))
	if err != nil {
		t.Errorf("Error scanning NullInt64: %v", err)
	}
	if !ni.Valid || ni.Int64 != 300 {
		t.Errorf("Expected Valid true and Int64 300, got Valid %v and Int64 %d", ni.Valid, ni.Int64)
	}

	// Test Value
	val, err := ni.Value()
	if err != nil {
		t.Errorf("Error getting Value from NullInt64: %v", err)
	}
	if val != int64(300) {
		t.Errorf("Expected Value 300, got '%v'", val)
	}
}

func TestNullBool(t *testing.T) {
	// Test constructor with bool
	nb := NewNullBool(true)
	if !nb.Valid || nb.Bool != true {
		t.Errorf("Expected Valid true and Bool true, got Valid %v and Bool %v", nb.Valid, nb.Bool)
	}

	// Test constructor with empty string
	nb = NewNullBoolFromString("")
	if nb.Valid {
		t.Errorf("Expected Valid false for empty string")
	}

	// Test constructor with valid string
	nb = NewNullBoolFromString("true")
	if !nb.Valid || nb.Bool != true {
		t.Errorf("Expected Valid true and Bool true, got Valid %v and Bool %v", nb.Valid, nb.Bool)
	}

	// Test JSON marshalling
	jsonData, err := json.Marshal(nb)
	if err != nil {
		t.Errorf("Error marshalling NullBool: %v", err)
	}
	if string(jsonData) != "true" {
		t.Errorf("Expected JSON 'true', got %s", jsonData)
	}

	// Test JSON unmarshalling
	err = json.Unmarshal([]byte(`false`), nb)
	if err != nil {
		t.Errorf("Error unmarshalling NullBool: %v", err)
	}
	if !nb.Valid || nb.Bool != false {
		t.Errorf("Expected Valid true and Bool false, got Valid %v and Bool %v", nb.Valid, nb.Bool)
	}

	// Test Scan
	err = nb.Scan(true)
	if err != nil {
		t.Errorf("Error scanning NullBool: %v", err)
	}
	if !nb.Valid || nb.Bool != true {
		t.Errorf("Expected Valid true and Bool true, got Valid %v and Bool %v", nb.Valid, nb.Bool)
	}

	// Test Value
	val, err := nb.Value()
	if err != nil {
		t.Errorf("Error getting Value from NullBool: %v", err)
	}
	if val != true {
		t.Errorf("Expected Value true, got '%v'", val)
	}
}

func TestNullFloat64(t *testing.T) {
	// Test constructor with float64
	nf := NewNullFloat64(3.14)
	if !nf.Valid || nf.Float64 != 3.14 {
		t.Errorf("Expected Valid true and Float64 3.14, got Valid %v and Float64 %f", nf.Valid, nf.Float64)
	}

	// Test constructor with empty string
	nf = NewNullFloat64FromString("")
	if nf.Valid {
		t.Errorf("Expected Valid false for empty string")
	}

	// Test constructor with valid string
	nf = NewNullFloat64FromString("2.718")
	if !nf.Valid || nf.Float64 != 2.718 {
		t.Errorf("Expected Valid true and Float64 2.718, got Valid %v and Float64 %f", nf.Valid, nf.Float64)
	}

	// Test JSON marshalling
	jsonData, err := json.Marshal(nf)
	if err != nil {
		t.Errorf("Error marshalling NullFloat64: %v", err)
	}
	if string(jsonData) != "2.718" {
		t.Errorf("Expected JSON '2.718', got %s", jsonData)
	}

	// Test JSON unmarshalling
	err = json.Unmarshal([]byte(`1.618`), nf)
	if err != nil {
		t.Errorf("Error unmarshalling NullFloat64: %v", err)
	}
	if !nf.Valid || nf.Float64 != 1.618 {
		t.Errorf("Expected Valid true and Float64 1.618, got Valid %v and Float64 %f", nf.Valid, nf.Float64)
	}

	// Test Scan
	err = nf.Scan(1.414)
	if err != nil {
		t.Errorf("Error scanning NullFloat64: %v", err)
	}
	if !nf.Valid || nf.Float64 != 1.414 {
		t.Errorf("Expected Valid true and Float64 1.414, got Valid %v and Float64 %f", nf.Valid, nf.Float64)
	}

	// Test Value
	val, err := nf.Value()
	if err != nil {
		t.Errorf("Error getting Value from NullFloat64: %v", err)
	}
	if val != 1.414 {
		t.Errorf("Expected Value 1.414, got '%v'", val)
	}
}

func TestCustomTime(t *testing.T) {
	// Test constructor with time.Time
	now := time.Now()
	ct := NewCustomTime(now)
	if !ct.Valid || !ct.Time.Equal(now) {
		t.Errorf("Expected Valid true and Time %v, got Valid %v and Time %v", now, ct.Valid, ct.Time)
	}

	// Test constructor with int64 (milliseconds)
	millis := now.UnixMilli()
	ct = NewCustomTimeInt64(millis)
	if !ct.Valid || ct.Time.UnixMilli() != millis {
		t.Errorf("Expected Valid true and Time with millis %d, got Valid %v and Time with millis %d", millis, ct.Valid, ct.Time.UnixMilli())
	}

	// Test JSON marshalling
	jsonData, err := json.Marshal(ct)
	if err != nil {
		t.Errorf("Error marshalling CustomTime: %v", err)
	}

	var tr TimeResponse
	err = json.Unmarshal(jsonData, &tr)
	if err != nil {
		t.Errorf("Error unmarshalling TimeResponse: %v", err)
	}
	if tr.UnixMS != millis {
		t.Errorf("Expected UnixMS %d, got %d", millis, tr.UnixMS)
	}

	// Test JSON unmarshalling
	ct = &CustomTime{}
	err = json.Unmarshal(jsonData, ct)
	if err != nil {
		t.Errorf("Error unmarshalling CustomTime: %v", err)
	}
	if !ct.Valid || ct.Time.UnixMilli() != millis {
		t.Errorf("Expected Valid true and Time with millis %d, got Valid %v and Time with millis %d", millis, ct.Valid, ct.Time.UnixMilli())
	}

	// Test Scan with time.Time
	err = ct.Scan(now)
	if err != nil {
		t.Errorf("Error scanning CustomTime: %v", err)
	}
	if !ct.Valid || !ct.Time.Equal(now) {
		t.Errorf("Expected Valid true and Time %v, got Valid %v and Time %v", now, ct.Valid, ct.Time)
	}

	// Test Value
	val, err := ct.Value()
	if err != nil {
		t.Errorf("Error getting Value from CustomTime: %v", err)
	}
	if val.(time.Time) != now {
		t.Errorf("Expected Value %v, got '%v'", now, val)
	}
}

func TestLocalizedText(t *testing.T) {
	// Test creation and assignment
	lt := LocalizedText{
		"en": "Hello",
		"fr": "Bonjour",
	}
	if lt["en"] != "Hello" || lt["fr"] != "Bonjour" {
		t.Errorf("Expected 'Hello' and 'Bonjour', got '%s' and '%s'", lt["en"], lt["fr"])
	}

	// Test JSON marshalling
	jsonData, err := json.Marshal(lt)
	if err != nil {
		t.Errorf("Error marshalling LocalizedText: %v", err)
	}
	expectedJSON := `{"en":"Hello","fr":"Bonjour"}`
	if string(jsonData) != expectedJSON {
		t.Errorf("Expected JSON '%s', got '%s'", expectedJSON, jsonData)
	}

	// Test JSON unmarshalling
	var lt2 LocalizedText
	err = json.Unmarshal(jsonData, &lt2)
	if err != nil {
		t.Errorf("Error unmarshalling LocalizedText: %v", err)
	}
	if lt2["en"] != "Hello" || lt2["fr"] != "Bonjour" {
		t.Errorf("Expected 'Hello' and 'Bonjour', got '%s' and '%s'", lt2["en"], lt2["fr"])
	}

	// Test Scan
	asBytes := []byte(`{"en":"Hi","es":"Hola"}`)
	err = lt.Scan(asBytes)
	if err != nil {
		t.Errorf("Error scanning LocalizedText: %v", err)
	}
	if lt["en"] != "Hi" || lt["es"] != "Hola" {
		t.Errorf("Expected 'Hi' and 'Hola', got '%s' and '%s'", lt["en"], lt["es"])
	}

	// Test Value
	val, err := lt.Value()
	if err != nil {
		t.Errorf("Error getting Value from LocalizedText: %v", err)
	}
	if string(val.([]byte)) != `{"en":"Hi","es":"Hola"}` {
		t.Errorf("Expected Value '%s', got '%s'", `{"en":"Hi","es":"Hola"}`, val)
	}
}

func TestIntDictionary(t *testing.T) {
	// Test creation and assignment
	id := IntDictionary{
		"one": 1,
		"two": 2,
	}
	if id["one"] != 1 || id["two"] != 2 {
		t.Errorf("Expected 1 and 2, got %d and %d", id["one"], id["two"])
	}

	// Test JSON marshalling
	jsonData, err := json.Marshal(id)
	if err != nil {
		t.Errorf("Error marshalling IntDictionary: %v", err)
	}
	expectedJSON := `{"one":1,"two":2}`
	if string(jsonData) != expectedJSON {
		t.Errorf("Expected JSON '%s', got '%s'", expectedJSON, jsonData)
	}

	// Test JSON unmarshalling
	var id2 IntDictionary
	err = json.Unmarshal(jsonData, &id2)
	if err != nil {
		t.Errorf("Error unmarshalling IntDictionary: %v", err)
	}
	if id2["one"] != 1 || id2["two"] != 2 {
		t.Errorf("Expected 1 and 2, got %d and %d", id2["one"], id2["two"])
	}

	// Test Scan
	asBytes := []byte(`{"three":3,"four":4}`)
	err = id.Scan(asBytes)
	if err != nil {
		t.Errorf("Error scanning IntDictionary: %v", err)
	}
	if id["three"] != 3 || id["four"] != 4 {
		t.Errorf("Expected 3 and 4, got %d and %d", id["three"], id["four"])
	}

	// Test Value
	val, err := id.Value()
	if err != nil {
		t.Errorf("Error getting Value from IntDictionary: %v", err)
	}
	if string(val.([]byte)) != `{"four":4,"three":3}` && string(val.([]byte)) != `{"three":3,"four":4}` {
		t.Errorf("Expected Value '%s' or '%s', got '%s'", `{"three":3,"four":4}`, `{"four":4,"three":3}`, val)
	}
}

func TestNullTypesIntegration(t *testing.T) {
	// Test NullString with NullInt64 in a struct
	type TestStruct struct {
		Name   NullString  `json:"name"`
		Age    NullInt64   `json:"age"`
		Score  NullFloat64 `json:"score"`
		Active NullBool    `json:"active"`
	}
	ts := TestStruct{
		Name:   *NewNullString("Alice"),
		Age:    *NewNullInt64(30),
		Score:  *NewNullFloat64(95.5),
		Active: *NewNullBool(true),
	}

	// Test JSON marshalling
	jsonData, err := json.Marshal(ts)
	if err != nil {
		t.Errorf("Error marshalling TestStruct: %v", err)
	}
	expectedJSON := `{"name":"Alice","age":30,"score":95.5,"active":true}`
	if string(jsonData) != expectedJSON {
		t.Errorf("Expected JSON '%s', got '%s'", expectedJSON, jsonData)
	}

	// Test JSON unmarshalling
	var ts2 TestStruct
	err = json.Unmarshal(jsonData, &ts2)
	if err != nil {
		t.Errorf("Error unmarshalling TestStruct: %v", err)
	}
	if ts2.Name.String != "Alice" || ts2.Age.Int64 != 30 || ts2.Score.Float64 != 95.5 || ts2.Active.Bool != true {
		t.Errorf("Expected Name 'Alice', Age 30, Score 95.5, Active true, got Name '%s', Age %d, Score %f, Active %v",
			ts2.Name.String, ts2.Age.Int64, ts2.Score.Float64, ts2.Active.Bool)
	}
}

func TestNullTypesWithNullValues(t *testing.T) {
	// Test NullString with null value
	ns := NewNullString("")
	jsonData, err := json.Marshal(ns)
	if err != nil {
		t.Errorf("Error marshalling NullString: %v", err)
	}
	if string(jsonData) != "null" {
		t.Errorf("Expected JSON 'null', got '%s'", jsonData)
	}

	// Test NullInt64 with null value
	ni := NewNullInt64FromString("")
	jsonData, err = json.Marshal(ni)
	if err != nil {
		t.Errorf("Error marshalling NullInt64: %v", err)
	}
	if string(jsonData) != "null" {
		t.Errorf("Expected JSON 'null', got '%s'", jsonData)
	}

	// Test NullFloat64 with null value
	nf := NewNullFloat64FromString("")
	jsonData, err = json.Marshal(nf)
	if err != nil {
		t.Errorf("Error marshalling NullFloat64: %v", err)
	}
	if string(jsonData) != "null" {
		t.Errorf("Expected JSON 'null', got '%s'", jsonData)
	}

	// Test NullBool with null value
	nb := NewNullBoolFromString("")
	jsonData, err = json.Marshal(nb)
	if err != nil {
		t.Errorf("Error marshalling NullBool: %v", err)
	}
	if string(jsonData) != "null" {
		t.Errorf("Expected JSON 'null', got '%s'", jsonData)
	}
}

func TestCustomTimeUnmarshalInvalidFormat(t *testing.T) {
	ct := &CustomTime{}
	err := json.Unmarshal([]byte(`"invalid time format"`), ct)
	if err == nil {
		t.Errorf("Expected error when unmarshalling invalid time format, got nil")
	}
}

func TestNullInt64UnmarshalInvalidFormat(t *testing.T) {
	ni := &NullInt64{}
	err := json.Unmarshal([]byte(`"not an int"`), ni)
	if err == nil {
		t.Errorf("Expected error when unmarshalling invalid int format, got nil")
	}
}

func TestNullBoolUnmarshalInvalidFormat(t *testing.T) {
	nb := &NullBool{}
	err := json.Unmarshal([]byte(`"not a bool"`), nb)
	if err == nil {
		t.Errorf("Expected error when unmarshalling invalid bool format, got nil")
	}
}

func TestNullFloat64UnmarshalInvalidFormat(t *testing.T) {
	nf := &NullFloat64{}
	err := json.Unmarshal([]byte(`"not a float"`), nf)
	if err == nil {
		t.Errorf("Expected error when unmarshalling invalid float format, got nil")
	}
}

func TestLocalizedTextUnmarshalInvalidFormat(t *testing.T) {
	lt := &LocalizedText{}
	err := json.Unmarshal([]byte(`"not a map"`), lt)
	if err == nil {
		t.Errorf("Expected error when unmarshalling invalid LocalizedText format, got nil")
	}
}

func TestIntDictionaryUnmarshalInvalidFormat(t *testing.T) {
	id := &IntDictionary{}
	err := json.Unmarshal([]byte(`"not a map"`), id)
	if err == nil {
		t.Errorf("Expected error when unmarshalling invalid IntDictionary format, got nil")
	}
}

func TestCustomTimeScanInvalidType(t *testing.T) {
	ct := &CustomTime{}
	err := ct.Scan([]byte("not a time"))
	if err == nil {
		t.Errorf("Expected error when scanning invalid type into CustomTime, got nil")
	}
}

func TestNullStringScanInvalidType(t *testing.T) {
	ns := &NullString{}
	err := ns.Scan(123)
	if err != nil {
		t.Errorf("Expected no error when scanning int into NullString, got %v", err)
	}
	if ns.Valid && ns.String != "123" {
		t.Errorf("Expected Valid true and String '123', got Valid %v and String '%s'", ns.Valid, ns.String)
	}
}

func TestNullInt64ScanInvalidType(t *testing.T) {
	ni := &NullInt64{}
	err := ni.Scan("not an int")
	if err == nil {
		t.Errorf("Expected error when scanning invalid type into NullInt64, got nil")
	}
}

func TestNullBoolScanInvalidType(t *testing.T) {
	nb := &NullBool{}
	err := nb.Scan("not a bool")
	if err == nil {
		t.Errorf("Expected error when scanning invalid type into NullBool, got nil")
	}
}

func TestNullFloat64ScanInvalidType(t *testing.T) {
	nf := &NullFloat64{}
	err := nf.Scan("not a float")
	if err == nil {
		t.Errorf("Expected error when scanning invalid type into NullFloat64, got nil")
	}
}

func TestLocalizedTextScanInvalidType(t *testing.T) {
	lt := &LocalizedText{}
	err := lt.Scan(123)
	if err == nil {
		t.Errorf("Expected error when scanning invalid type into LocalizedText, got nil")
	}
}

func TestIntDictionaryScanInvalidType(t *testing.T) {
	id := &IntDictionary{}
	err := id.Scan(123)
	if err == nil {
		t.Errorf("Expected error when scanning invalid type into IntDictionary, got nil")
	}
}

func TestNullStringValue(t *testing.T) {
	ns := NewNullString("test")
	val, err := ns.Value()
	if err != nil {
		t.Errorf("Expected no error from NullString.Value(), got %v", err)
	}
	if val != "test" {
		t.Errorf("Expected value 'test', got '%v'", val)
	}
}

func TestNullInt64Value(t *testing.T) {
	ni := NewNullInt64(123)
	val, err := ni.Value()
	if err != nil {
		t.Errorf("Expected no error from NullInt64.Value(), got %v", err)
	}
	if val != int64(123) {
		t.Errorf("Expected value 123, got '%v'", val)
	}
}

func TestNullBoolValue(t *testing.T) {
	nb := NewNullBool(true)
	val, err := nb.Value()
	if err != nil {
		t.Errorf("Expected no error from NullBool.Value(), got %v", err)
	}
	if val != true {
		t.Errorf("Expected value true, got '%v'", val)
	}
}

func TestNullFloat64Value(t *testing.T) {
	nf := NewNullFloat64(3.14)
	val, err := nf.Value()
	if err != nil {
		t.Errorf("Expected no error from NullFloat64.Value(), got %v", err)
	}
	if val != 3.14 {
		t.Errorf("Expected value 3.14, got '%v'", val)
	}
}

func TestLocalizedTextValue(t *testing.T) {
	lt := LocalizedText{"en": "Hello"}
	val, err := lt.Value()
	if err != nil {
		t.Errorf("Expected no error from LocalizedText.Value(), got %v", err)
	}
	expectedJSON := `{"en":"Hello"}`
	if string(val.([]byte)) != expectedJSON {
		t.Errorf("Expected value '%s', got '%s'", expectedJSON, val)
	}
}

func TestIntDictionaryValue(t *testing.T) {
	id := IntDictionary{"one": 1}
	val, err := id.Value()
	if err != nil {
		t.Errorf("Expected no error from IntDictionary.Value(), got %v", err)
	}
	expectedJSON := `{"one":1}`
	if string(val.([]byte)) != expectedJSON {
		t.Errorf("Expected value '%s', got '%s'", expectedJSON, val)
	}
}

func TestNullStringValueNil(t *testing.T) {
	ns := NewNullString("")
	val, err := ns.Value()
	if err != nil {
		t.Errorf("Expected no error from NullString.Value(), got %v", err)
	}
	if val != nil {
		t.Errorf("Expected value nil, got '%v'", val)
	}
}

func TestNullInt64ValueNil(t *testing.T) {
	ni := NewNullInt64FromString("")
	val, err := ni.Value()
	if err != nil {
		t.Errorf("Expected no error from NullInt64.Value(), got %v", err)
	}
	if val != nil {
		t.Errorf("Expected value nil, got '%v'", val)
	}
}

func TestNullBoolValueNil(t *testing.T) {
	nb := NewNullBoolFromString("")
	val, err := nb.Value()
	if err != nil {
		t.Errorf("Expected no error from NullBool.Value(), got %v", err)
	}
	if val != nil {
		t.Errorf("Expected value nil, got '%v'", val)
	}
}

func TestNullFloat64ValueNil(t *testing.T) {
	nf := NewNullFloat64FromString("")
	val, err := nf.Value()
	if err != nil {
		t.Errorf("Expected no error from NullFloat64.Value(), got %v", err)
	}
	if val != nil {
		t.Errorf("Expected value nil, got '%v'", val)
	}
}

func TestCustomTimeValueNil(t *testing.T) {
	ct := NewCustomTimeNull()
	val, err := ct.Value()
	if err != nil {
		t.Errorf("Expected no error from CustomTime.Value(), got %v", err)
	}
	if val != nil {
		t.Errorf("Expected value nil, got '%v'", val)
	}
}

func TestLocalizedTextValueNil(t *testing.T) {
	var lt LocalizedText
	val, err := lt.Value()
	if err != nil {
		t.Errorf("Expected no error from LocalizedText.Value(), got %v", err)
	}
	if val != nil {
		t.Errorf("Expected value nil, got '%v'", val)
	}
}

func TestIntDictionaryValueNil(t *testing.T) {
	var id IntDictionary
	val, err := id.Value()
	if err != nil {
		t.Errorf("Expected no error from IntDictionary.Value(), got %v", err)
	}
	if val != nil {
		t.Errorf("Expected value nil, got '%v'", val)
	}
}

func TestCustomTimeMarshalNull(t *testing.T) {
	ct := NewCustomTimeNull()
	jsonData, err := json.Marshal(ct)
	if err != nil {
		t.Errorf("Error marshalling CustomTime: %v", err)
	}
	if string(jsonData) != "null" {
		t.Errorf("Expected JSON 'null', got '%s'", jsonData)
	}
}

func TestCustomTimeUnmarshalNull(t *testing.T) {
	ct := &CustomTime{}
	err := json.Unmarshal([]byte(`null`), ct)
	if err != nil {
		t.Errorf("Error unmarshalling CustomTime: %v", err)
	}
	if ct.Valid {
		t.Errorf("Expected Valid false, got Valid %v", ct.Valid)
	}
}

func TestCustomTimeUnmarshalUnixMS(t *testing.T) {
	ct := &CustomTime{}
	millis := time.Now().UnixMilli()
	err := json.Unmarshal([]byte(strconv.FormatInt(millis, 10)), ct)
	if err != nil {
		t.Errorf("Error unmarshalling CustomTime: %v", err)
	}
	if !ct.Valid || ct.Time.UnixMilli() != millis {
		t.Errorf("Expected Valid true and Time with millis %d, got Valid %v and Time with millis %d", millis, ct.Valid, ct.Time.UnixMilli())
	}
}

func TestCustomTimeUnmarshalStringDate(t *testing.T) {
	ct := &CustomTime{}
	dateStr := `"2023-01-01"`
	err := json.Unmarshal([]byte(dateStr), ct)
	if err != nil {
		t.Errorf("Error unmarshalling CustomTime: %v", err)
	}
	expectedTime, _ := time.Parse("2006-01-02", "2023-01-01")
	if !ct.Valid || !ct.Time.Equal(expectedTime) {
		t.Errorf("Expected Valid true and Time %v, got Valid %v and Time %v", expectedTime, ct.Valid, ct.Time)
	}
}

func TestNewNullStringValid(t *testing.T) {
	// Test with empty string that should be valid
	ns := NewNullStringValid("")
	if !ns.Valid {
		t.Errorf("Expected Valid true for empty string with NewNullStringValid")
	}

	// Test marshalling
	jsonData, err := json.Marshal(ns)
	if err != nil {
		t.Errorf("Error marshalling NullString: %v", err)
	}
	if string(jsonData) != `""` {
		t.Errorf("Expected JSON '\"\"', got '%s'", jsonData)
	}
}

func TestNewNullInt64Zero(t *testing.T) {
	// Test constructor for zero value
	ni := NewNullInt64Zero()
	if !ni.Valid || ni.Int64 != 0 {
		t.Errorf("Expected Valid true and Int64 0, got Valid %v and Int64 %d", ni.Valid, ni.Int64)
	}

	// Test marshalling
	jsonData, err := json.Marshal(ni)
	if err != nil {
		t.Errorf("Error marshalling NullInt64: %v", err)
	}
	if string(jsonData) != "0" {
		t.Errorf("Expected JSON '0', got '%s'", jsonData)
	}
}

func TestNewNullBoolFalse(t *testing.T) {
	// Test constructor for false value
	nb := NewNullBoolFalse()
	if !nb.Valid || nb.Bool != false {
		t.Errorf("Expected Valid true and Bool false, got Valid %v and Bool %v", nb.Valid, nb.Bool)
	}

	// Test marshalling
	jsonData, err := json.Marshal(nb)
	if err != nil {
		t.Errorf("Error marshalling NullBool: %v", err)
	}
	if string(jsonData) != "false" {
		t.Errorf("Expected JSON 'false', got '%s'", jsonData)
	}
}

func TestNewNullFloat64Zero(t *testing.T) {
	// Test constructor for zero value
	nf := NewNullFloat64Zero()
	if !nf.Valid || nf.Float64 != 0.0 {
		t.Errorf("Expected Valid true and Float64 0.0, got Valid %v and Float64 %f", nf.Valid, nf.Float64)
	}

	// Test marshalling
	jsonData, err := json.Marshal(nf)
	if err != nil {
		t.Errorf("Error marshalling NullFloat64: %v", err)
	}
	if string(jsonData) != "0" {
		t.Errorf("Expected JSON '0', got '%s'", jsonData)
	}
}

// Benchmark tests - New optimized implementation
func BenchmarkNullStringJSON(b *testing.B) {
	ns := NewNullString("test")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(ns)
	}
}

func BenchmarkNullInt64JSON(b *testing.B) {
	ni := NewNullInt64(42)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(ni)
	}
}

func BenchmarkNullBoolJSON(b *testing.B) {
	nb := NewNullBool(true)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(nb)
	}
}

func BenchmarkNullFloat64JSON(b *testing.B) {
	nf := NewNullFloat64(3.14)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(nf)
	}
}

// Test nulls
func BenchmarkNullStringNullJSON(b *testing.B) {
	ns := NewNullStringNull()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(ns)
	}
}

func BenchmarkNullInt64NullJSON(b *testing.B) {
	ni := NewNullInt64Null()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(ni)
	}
}

func BenchmarkNullBoolNullJSON(b *testing.B) {
	nb := NewNullBoolNull()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(nb)
	}
}

func BenchmarkNullFloat64NullJSON(b *testing.B) {
	nf := NewNullFloat64Null()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(nf)
	}
}

func BenchmarkCustomTimeJSON(b *testing.B) {
	now := time.Now()
	ct := NewCustomTime(now)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(ct)
	}
}

func BenchmarkLocalizedTextJSON(b *testing.B) {
	lt := LocalizedText{"en": "Hello", "fr": "Bonjour"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(lt)
	}
}

func BenchmarkIntDictionaryJSON(b *testing.B) {
	id := IntDictionary{"one": 1, "two": 2}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(id)
	}
}

func BenchmarkNullTypesStruct(b *testing.B) {
	type TestStruct struct {
		Name   NullString  `json:"name"`
		Age    NullInt64   `json:"age"`
		Score  NullFloat64 `json:"score"`
		Active NullBool    `json:"active"`
	}
	
	ts := TestStruct{
		Name:   *NewNullString("Alice"),
		Age:    *NewNullInt64(30),
		Score:  *NewNullFloat64(95.5),
		Active: *NewNullBool(true),
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(ts)
	}
}

// Benchmarks for JSON Unmarshaling
func BenchmarkNullStringUnmarshalJSON(b *testing.B) {
	data := []byte(`"test string"`)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var ns NullString
		_ = json.Unmarshal(data, &ns)
	}
}

func BenchmarkNullStringUnmarshalNullJSON(b *testing.B) {
	data := []byte(`null`)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var ns NullString
		_ = json.Unmarshal(data, &ns)
	}
}

func BenchmarkNullInt64UnmarshalJSON(b *testing.B) {
	data := []byte(`42`)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var ni NullInt64
		_ = json.Unmarshal(data, &ni)
	}
}

func BenchmarkNullInt64UnmarshalNullJSON(b *testing.B) {
	data := []byte(`null`)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var ni NullInt64
		_ = json.Unmarshal(data, &ni)
	}
}

func BenchmarkNullBoolUnmarshalJSON(b *testing.B) {
	data := []byte(`true`)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var nb NullBool
		_ = json.Unmarshal(data, &nb)
	}
}

func BenchmarkNullFloat64UnmarshalJSON(b *testing.B) {
	data := []byte(`3.14159`)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var nf NullFloat64
		_ = json.Unmarshal(data, &nf)
	}
}

func BenchmarkNullTypesStructUnmarshal(b *testing.B) {
	type TestStruct struct {
		Name   NullString  `json:"name"`
		Age    NullInt64   `json:"age"`
		Score  NullFloat64 `json:"score"`
		Active NullBool    `json:"active"`
	}
	
	data := []byte(`{"name":"Alice","age":30,"score":95.5,"active":true}`)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var ts TestStruct
		_ = json.Unmarshal(data, &ts)
	}
}

// Binary encoding benchmarks
func BenchmarkNullStringBinary(b *testing.B) {
	ns := NewNullString("test string")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		_, _ = ns.WriteTo(&buf)
	}
}

func BenchmarkNullInt64Binary(b *testing.B) {
	ni := NewNullInt64(42)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		_, _ = ni.WriteTo(&buf)
	}
}

func BenchmarkNullBoolBinary(b *testing.B) {
	nb := NewNullBool(true)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		_, _ = nb.WriteTo(&buf)
	}
}

func BenchmarkNullFloat64Binary(b *testing.B) {
	nf := NewNullFloat64(3.14)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		_, _ = nf.WriteTo(&buf)
	}
}

func BenchmarkCustomTimeBinary(b *testing.B) {
	ct := NewCustomTime(time.Now())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		_, _ = ct.WriteTo(&buf)
	}
}

// Binary decoding benchmarks
func BenchmarkNullStringFromBinary(b *testing.B) {
	ns := NewNullString("test string")
	var buf bytes.Buffer
	_, _ = ns.WriteTo(&buf)
	data := buf.Bytes()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var ns2 NullString
		reader := bytes.NewReader(data)
		_, _ = ns2.ReadFrom(reader)
	}
}

func BenchmarkNullInt64FromBinary(b *testing.B) {
	ni := NewNullInt64(42)
	var buf bytes.Buffer
	_, _ = ni.WriteTo(&buf)
	data := buf.Bytes()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var ni2 NullInt64
		reader := bytes.NewReader(data)
		_, _ = ni2.ReadFrom(reader)
	}
}

func BenchmarkNullBoolFromBinary(b *testing.B) {
	nb := NewNullBool(true)
	var buf bytes.Buffer
	_, _ = nb.WriteTo(&buf)
	data := buf.Bytes()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var nb2 NullBool
		reader := bytes.NewReader(data)
		_, _ = nb2.ReadFrom(reader)
	}
}

func BenchmarkNullFloat64FromBinary(b *testing.B) {
	nf := NewNullFloat64(3.14)
	var buf bytes.Buffer
	_, _ = nf.WriteTo(&buf)
	data := buf.Bytes()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var nf2 NullFloat64
		reader := bytes.NewReader(data)
		_, _ = nf2.ReadFrom(reader)
	}
}

func BenchmarkCustomTimeFromBinary(b *testing.B) {
	ct := NewCustomTime(time.Now())
	var buf bytes.Buffer
	_, _ = ct.WriteTo(&buf)
	data := buf.Bytes()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var ct2 CustomTime
		reader := bytes.NewReader(data)
		_, _ = ct2.ReadFrom(reader)
	}
}

// Complex struct for comparing JSON vs Binary performance
type ComplexStruct struct {
	ID          NullInt64   `json:"id"`
	Name        NullString  `json:"name"`
	Description NullString  `json:"description"`
	Score       NullFloat64 `json:"score"`
	IsActive    NullBool    `json:"is_active"`
	CreatedAt   CustomTime  `json:"created_at"`
	UpdatedAt   CustomTime  `json:"updated_at"`
}

// WriteTo implements binary serialization for ComplexStruct
func (cs ComplexStruct) WriteTo(w io.Writer) (n int64, err error) {
	var n1, n2, n3, n4, n5, n6, n7 int64
	
	n1, err = cs.ID.WriteTo(w)
	if err != nil {
		return n1, err
	}
	
	n2, err = cs.Name.WriteTo(w)
	if err != nil {
		return n1 + n2, err
	}
	
	n3, err = cs.Description.WriteTo(w)
	if err != nil {
		return n1 + n2 + n3, err
	}
	
	n4, err = cs.Score.WriteTo(w)
	if err != nil {
		return n1 + n2 + n3 + n4, err
	}
	
	n5, err = cs.IsActive.WriteTo(w)
	if err != nil {
		return n1 + n2 + n3 + n4 + n5, err
	}
	
	n6, err = cs.CreatedAt.WriteTo(w)
	if err != nil {
		return n1 + n2 + n3 + n4 + n5 + n6, err
	}
	
	n7, err = cs.UpdatedAt.WriteTo(w)
	return n1 + n2 + n3 + n4 + n5 + n6 + n7, err
}

// ReadFrom implements binary deserialization for ComplexStruct
func (cs *ComplexStruct) ReadFrom(r io.Reader) (n int64, err error) {
	var n1, n2, n3, n4, n5, n6, n7 int64
	
	n1, err = cs.ID.ReadFrom(r)
	if err != nil {
		return n1, err
	}
	
	n2, err = cs.Name.ReadFrom(r)
	if err != nil {
		return n1 + n2, err
	}
	
	n3, err = cs.Description.ReadFrom(r)
	if err != nil {
		return n1 + n2 + n3, err
	}
	
	n4, err = cs.Score.ReadFrom(r)
	if err != nil {
		return n1 + n2 + n3 + n4, err
	}
	
	n5, err = cs.IsActive.ReadFrom(r)
	if err != nil {
		return n1 + n2 + n3 + n4 + n5, err
	}
	
	n6, err = cs.CreatedAt.ReadFrom(r)
	if err != nil {
		return n1 + n2 + n3 + n4 + n5 + n6, err
	}
	
	n7, err = cs.UpdatedAt.ReadFrom(r)
	return n1 + n2 + n3 + n4 + n5 + n6 + n7, err
}

// Benchmark comparing JSON vs Binary serialization of complex struct
func BenchmarkComplexStructJSON(b *testing.B) {
	cs := ComplexStruct{
		ID:          *NewNullInt64(12345),
		Name:        *NewNullString("Test Name"),
		Description: *NewNullString("This is a test description with some more text"),
		Score:       *NewNullFloat64(98.76),
		IsActive:    *NewNullBool(true),
		CreatedAt:   *NewCustomTime(time.Now().Add(-24 * time.Hour)),
		UpdatedAt:   *NewCustomTime(time.Now()),
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(cs)
	}
}

func BenchmarkComplexStructBinary(b *testing.B) {
	cs := ComplexStruct{
		ID:          *NewNullInt64(12345),
		Name:        *NewNullString("Test Name"),
		Description: *NewNullString("This is a test description with some more text"),
		Score:       *NewNullFloat64(98.76),
		IsActive:    *NewNullBool(true),
		CreatedAt:   *NewCustomTime(time.Now().Add(-24 * time.Hour)),
		UpdatedAt:   *NewCustomTime(time.Now()),
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		_, _ = cs.WriteTo(&buf)
	}
}

func BenchmarkComplexStructFromJSON(b *testing.B) {
	cs := ComplexStruct{
		ID:          *NewNullInt64(12345),
		Name:        *NewNullString("Test Name"),
		Description: *NewNullString("This is a test description with some more text"),
		Score:       *NewNullFloat64(98.76),
		IsActive:    *NewNullBool(true),
		CreatedAt:   *NewCustomTime(time.Now().Add(-24 * time.Hour)),
		UpdatedAt:   *NewCustomTime(time.Now()),
	}
	
	jsonData, _ := json.Marshal(cs)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var cs2 ComplexStruct
		_ = json.Unmarshal(jsonData, &cs2)
	}
}

func BenchmarkComplexStructFromBinary(b *testing.B) {
	cs := ComplexStruct{
		ID:          *NewNullInt64(12345),
		Name:        *NewNullString("Test Name"),
		Description: *NewNullString("This is a test description with some more text"),
		Score:       *NewNullFloat64(98.76),
		IsActive:    *NewNullBool(true),
		CreatedAt:   *NewCustomTime(time.Now().Add(-24 * time.Hour)),
		UpdatedAt:   *NewCustomTime(time.Now()),
	}
	
	var buf bytes.Buffer
	_, _ = cs.WriteTo(&buf)
	data := buf.Bytes()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var cs2 ComplexStruct
		reader := bytes.NewReader(data)
		_, _ = cs2.ReadFrom(reader)
	}
}