package features

import "testing"

func BenchmarkSuspiciousTokens_Safe(b *testing.B) {
	query := "q=some+innocent+query+that+is+long+enough"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		suspiciousTokens(query)
	}
}

func BenchmarkSuspiciousTokens_Suspicious(b *testing.B) {
	query := "q=select+*+from+users"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		suspiciousTokens(query)
	}
}
