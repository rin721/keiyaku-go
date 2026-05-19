package plugin

import (
	"sort"

	domainplugin "github.com/rin721/keiyaku-go/internal/domain/plugin"
)

type RouteMatcher struct{}

func (RouteMatcher) Match(method string, path string, routes []*domainplugin.Route) (*domainplugin.Route, string, bool) {
	type candidate struct {
		route       *domainplugin.Route
		suffix      string
		methodScore int
		matchScore  int
		pathLen     int
	}
	var candidates []candidate
	for _, route := range routes {
		if route == nil {
			continue
		}
		suffix, ok := route.Matches(method, path)
		if !ok {
			continue
		}
		methodScore := 0
		if route.Method != domainplugin.MethodAny {
			methodScore = 1
		}
		matchScore := 0
		if route.MatchType == domainplugin.MatchTypeExact {
			matchScore = 1
		}
		candidates = append(candidates, candidate{
			route:       route,
			suffix:      suffix,
			methodScore: methodScore,
			matchScore:  matchScore,
			pathLen:     len(route.GatewayPath),
		})
	}
	if len(candidates) == 0 {
		return nil, "", false
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].methodScore != candidates[j].methodScore {
			return candidates[i].methodScore > candidates[j].methodScore
		}
		if candidates[i].matchScore != candidates[j].matchScore {
			return candidates[i].matchScore > candidates[j].matchScore
		}
		return candidates[i].pathLen > candidates[j].pathLen
	})
	return candidates[0].route, candidates[0].suffix, true
}

func bestRoute(method string, path string, routes []*domainplugin.Route) (*domainplugin.Route, string, bool) {
	return RouteMatcher{}.Match(method, path, routes)
}
