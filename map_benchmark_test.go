package octypes

import (
	"encoding/json"
	"testing"
)

var complexJSONForStringInternBenchmark = []byte(`{
	"en": "English",
	"fr": "French",
	"de": "German",
	"es": "Spanish",
	"en-US": "American English",
	"fr-FR": "France French",
	"de-DE": "German German",
	"es-ES": "Spain Spanish",
	"created_at": "2023-01-01",
	"updated_at": "2023-01-02",
	"user_id": "12345",
	"description": "This is a long description that exceeds 24 bytes and should be interned"
}`)

// Test for string-interning
func BenchmarkLocalizedTextUnmarshalWithIntern(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var lt LocalizedText
		_ = json.Unmarshal(complexJSONForStringInternBenchmark, &lt)
	}
}

// Test for map operations 
func BenchmarkLocalizedTextMapOperations(b *testing.B) {
	var lt LocalizedText
	_ = json.Unmarshal(complexJSONForStringInternBenchmark, &lt)
	
	b.ResetTimer()
	b.ReportAllocs()
	
	longKey := "description"
	mediumKey := "created_at"
	shortKey := "en"
	
	for i := 0; i < b.N; i++ {
		_ = lt[longKey]
		_ = lt[mediumKey]
		_ = lt[shortKey]
	}
}

// Reference implementation without string interning for comparison
type ReferenceLocalizedText map[string]string

func (rlt *ReferenceLocalizedText) UnmarshalJSON(b []byte) error {
	if isNullJSON(b) {
		*rlt = nil
		return nil
	}
	
	if len(b) <= 2 && b[0] == '{' && b[len(b)-1] == '}' {
		*rlt = make(ReferenceLocalizedText)
		return nil
	}
	
	m := make(map[string]string)
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	
	*rlt = make(ReferenceLocalizedText, len(m))
	for k, v := range m {
		(*rlt)[k] = v // No interning
	}
	
	return nil
}

func BenchmarkReferenceLocalizedTextUnmarshal(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var rlt ReferenceLocalizedText
		_ = json.Unmarshal(complexJSONForStringInternBenchmark, &rlt)
	}
}

func BenchmarkReferenceLocalizedTextMapOperations(b *testing.B) {
	var rlt ReferenceLocalizedText
	_ = json.Unmarshal(complexJSONForStringInternBenchmark, &rlt)
	
	b.ResetTimer()
	b.ReportAllocs()
	
	longKey := "description"
	mediumKey := "created_at"
	shortKey := "en"
	
	for i := 0; i < b.N; i++ {
		_ = rlt[longKey]
		_ = rlt[mediumKey]
		_ = rlt[shortKey]
	}
}