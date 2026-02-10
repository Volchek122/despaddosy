package policy

import (
	"net/http"
	"testing"

	"despaddosy/pkg/config"
)

func TestMatchRoute(t *testing.T) {
	routes := []config.RoutePolicy{
		{ID: "root", PathPrefix: "/"},
		{ID: "api", PathPrefix: "/api"},
	}
	matcher := NewMatcher(routes)
	req, _ := http.NewRequest(http.MethodGet, "http://example.com/api/posts", nil)
	route, ok := matcher.Match(req)
	if !ok || route.ID != "api" {
		t.Fatalf("expected api route match, got %+v", route)
	}
}

func TestMethodAllowed(t *testing.T) {
	route := config.RoutePolicy{AllowedMethods: []string{"GET"}}
	if !MethodAllowed(route, http.MethodGet) {
		t.Fatal("expected GET allowed")
	}
	if MethodAllowed(route, http.MethodPost) {
		t.Fatal("expected POST denied")
	}
}
