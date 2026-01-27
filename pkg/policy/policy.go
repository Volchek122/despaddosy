package policy

import (
	"net/http"
	"strings"

	"despaddosy/pkg/config"
)

type Matcher struct {
	routes []config.RoutePolicy
}

func NewMatcher(routes []config.RoutePolicy) *Matcher {
	return &Matcher{routes: routes}
}

func (m *Matcher) Match(r *http.Request) (config.RoutePolicy, bool) {
	path := r.URL.Path
	for _, route := range m.routes {
		if route.Host != "" && !strings.EqualFold(route.Host, r.Host) {
			continue
		}
		if strings.HasPrefix(path, route.PathPrefix) {
			return route, true
		}
	}
	return config.RoutePolicy{}, false
}

func MethodAllowed(route config.RoutePolicy, method string) bool {
	if len(route.AllowedMethods) == 0 {
		return true
	}
	for _, allowed := range route.AllowedMethods {
		if strings.EqualFold(allowed, method) {
			return true
		}
	}
	return false
}
