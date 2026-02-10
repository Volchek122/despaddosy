package features

import (
	"crypto/sha256"
	"encoding/hex"
	"math"
	"net/http"
	"strconv"
	"strings"
)

type Vector struct {
	Method               string `json:"method"`
	RouteID              string `json:"route_id"`
	PathLen              int    `json:"path_len"`
	QueryLen             int    `json:"query_len"`
	BodyLen              int64  `json:"body_len"`
	ContentType          string `json:"content_type"`
	UserAgentLen         int    `json:"user_agent_len"`
	UserAgentFamily      string `json:"user_agent_family"`
	AcceptLangPresent    bool   `json:"accept_lang_present"`
	HasAuth              bool   `json:"has_auth"`
	UserIDPresent        bool   `json:"user_id_present"`
	IPHash               string `json:"ip_hash"`
	RequestRateSnapshot  int    `json:"request_rate_snapshot"`
	SuspiciousTokenFlag  bool   `json:"suspicious_tokens_flag"`
	SuspiciousTokenClass string `json:"token_class"`
	HighEntropyQueryFlag bool   `json:"high_entropy_query_flag"`
	TLSVersion           string `json:"tls_version"`
	TLSCipher            string `json:"tls_cipher"`
}

func Extract(r *http.Request, routeID string, bodyLen int64, rateSnapshot int, hasAuth bool, userID string, ipSalt string) Vector {
	ua := r.UserAgent()
	query := r.URL.RawQuery
	ip := clientIP(r.RemoteAddr)
	vector := Vector{
		Method:              r.Method,
		RouteID:             routeID,
		PathLen:             len(r.URL.Path),
		QueryLen:            len(query),
		BodyLen:             bodyLen,
		ContentType:         normalizedContentType(r.Header.Get("Content-Type")),
		UserAgentLen:        len(ua),
		UserAgentFamily:     classifyUA(ua),
		AcceptLangPresent:   r.Header.Get("Accept-Language") != "",
		HasAuth:             hasAuth,
		UserIDPresent:       userID != "",
		IPHash:              hashIP(ip, ipSalt),
		RequestRateSnapshot: rateSnapshot,
	}
	vector.SuspiciousTokenFlag, vector.SuspiciousTokenClass = suspiciousTokens(query)
	vector.HighEntropyQueryFlag = entropy(query) > 4.0
	if r.TLS != nil {
		vector.TLSVersion = tlsVersion(r.TLS.Version)
		vector.TLSCipher = tlsCipher(r.TLS.CipherSuite)
	}
	return vector
}

func normalizedContentType(contentType string) string {
	if contentType == "" {
		return ""
	}
	parts := strings.Split(contentType, ";")
	return strings.TrimSpace(strings.ToLower(parts[0]))
}

func classifyUA(ua string) string {
	ua = strings.ToLower(ua)
	switch {
	case strings.Contains(ua, "chrome"):
		return "chrome"
	case strings.Contains(ua, "firefox"):
		return "firefox"
	case strings.Contains(ua, "safari"):
		return "safari"
	case strings.Contains(ua, "curl"):
		return "cli"
	case ua == "":
		return "empty"
	default:
		return "other"
	}
}

var suspiciousTokenMap = map[string]string{
	"select":    "sql",
	"union":     "sql",
	"<script":   "xss",
	"../":       "path",
	"%3cscript": "xss",
}

func suspiciousTokens(query string) (bool, string) {
	low := strings.ToLower(query)
	for token, class := range suspiciousTokenMap {
		if strings.Contains(low, token) {
			return true, class
		}
	}
	return false, ""
}

func entropy(s string) float64 {
	if s == "" {
		return 0
	}
	counts := make(map[rune]float64)
	for _, ch := range s {
		counts[ch]++
	}
	var ent float64
	length := float64(len(s))
	for _, count := range counts {
		p := count / length
		ent -= p * math.Log2(p)
	}
	return ent
}

func clientIP(remoteAddr string) string {
	parts := strings.Split(remoteAddr, ":")
	if len(parts) == 0 {
		return remoteAddr
	}
	return parts[0]
}

func hashIP(ip, salt string) string {
	h := sha256.Sum256([]byte(ip + salt))
	return hex.EncodeToString(h[:])
}

func tlsVersion(version uint16) string {
	switch version {
	case 0x0304:
		return "TLS1.3"
	case 0x0303:
		return "TLS1.2"
	default:
		return "unknown"
	}
}

func tlsCipher(cipher uint16) string {
	return "cipher-" + strconv.Itoa(int(cipher))
}
