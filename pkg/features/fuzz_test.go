package features

import (
	"net/http"
	"net/url"
	"testing"
)

func FuzzQueryAnalysis(f *testing.F) {
	f.Add("q=hello")
	f.Add("q=%3cscript%3e")
	f.Fuzz(func(t *testing.T, query string) {
		_ = entropy(query)
		_, _ = suspiciousTokens(query)
	})
}

func FuzzExtract(f *testing.F) {
	f.Add("/path", "ua")
	f.Fuzz(func(t *testing.T, path, ua string) {
		u := &url.URL{Path: path, RawQuery: "q=test"}
		req := &http.Request{Method: http.MethodGet, URL: u, Header: http.Header{"User-Agent": {ua}}}
		_ = Extract(req, "route", 0, 0, false, "", "salt")
	})
}
