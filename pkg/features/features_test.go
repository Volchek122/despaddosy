package features

import (
	"net/http"
	"testing"
)

func TestExtractSuspicious(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "http://example.com/search?q=select+1", nil)
	vector := Extract(req, "route", 0, 0, false, "", "salt")
	if !vector.SuspiciousTokenFlag {
		t.Fatal("expected suspicious flag")
	}
	if vector.IPHash == "" {
		t.Fatal("expected ip hash")
	}
}
