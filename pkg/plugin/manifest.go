package plugin

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"
)

var keyPattern = regexp.MustCompile(`^[a-z][a-z0-9-]{2,63}$`)

func NormalizeManifest(manifest Manifest) Manifest {
	if manifest.SchemaVersion == "" {
		manifest.SchemaVersion = DefaultSchemaVersion
	}
	if manifest.Protocol == "" {
		manifest.Protocol = DefaultProtocol
	}
	if manifest.HealthPath == "" {
		manifest.HealthPath = "/healthz"
	}
	for i := range manifest.Routes {
		manifest.Routes[i] = NormalizeRoute(manifest.Routes[i])
	}
	return manifest
}

func NormalizeRoute(route Route) Route {
	if route.Method == "" {
		route.Method = MethodAny
	}
	route.Method = Method(strings.ToUpper(string(route.Method)))
	if route.MatchType == "" {
		route.MatchType = MatchTypePrefix
	}
	if route.AuthPolicy == "" {
		route.AuthPolicy = AuthPolicyInherit
	}
	if route.UpstreamPath == "" {
		route.UpstreamPath = route.Path
	}
	if route.TimeoutMS < 0 {
		route.TimeoutMS = 0
	}
	return route
}

func ValidateManifest(manifest Manifest) error {
	manifest = NormalizeManifest(manifest)
	if manifest.SchemaVersion != DefaultSchemaVersion {
		return validationError("unsupported schema_version", ErrInvalidManifest)
	}
	if !keyPattern.MatchString(manifest.PluginKey) {
		return validationError("plugin_key must match ^[a-z][a-z0-9-]{2,63}$", ErrInvalidManifest)
	}
	if strings.TrimSpace(manifest.Name) == "" {
		return validationError("name is required", ErrInvalidManifest)
	}
	if strings.TrimSpace(manifest.Version) == "" {
		return validationError("version is required", ErrInvalidManifest)
	}
	if !keyPattern.MatchString(manifest.InstanceID) {
		return validationError("instance_id must match ^[a-z][a-z0-9-]{2,63}$", ErrInvalidManifest)
	}
	if manifest.Protocol != ProtocolHTTP {
		return validationError("protocol must be http", ErrInvalidManifest)
	}
	if _, err := parseServiceURL(manifest.BaseURL); err != nil {
		return validationError("base_url is invalid", err)
	}
	if !validPath(manifest.HealthPath) {
		return validationError("health_path must start with /", ErrInvalidManifest)
	}
	if len(manifest.Routes) == 0 {
		return validationError("routes must not be empty", ErrInvalidManifest)
	}
	seen := map[string]struct{}{}
	for _, route := range manifest.Routes {
		route = NormalizeRoute(route)
		if err := ValidateRoute(route); err != nil {
			return err
		}
		key := fmt.Sprintf("%s:%s:%s", route.Method, route.MatchType, route.Path)
		if _, ok := seen[key]; ok {
			return validationError("duplicate route declaration", ErrInvalidManifest)
		}
		seen[key] = struct{}{}
	}
	return nil
}

func ValidateRoute(route Route) error {
	switch route.Method {
	case MethodAny, MethodGet, MethodPost, MethodPut, MethodPatch, MethodDelete:
	default:
		return validationError("unsupported route method", ErrInvalidManifest)
	}
	switch route.MatchType {
	case MatchTypeExact, MatchTypePrefix:
	default:
		return validationError("unsupported match_type", ErrInvalidManifest)
	}
	switch route.AuthPolicy {
	case AuthPolicyInherit, AuthPolicyAuthenticated, AuthPolicyRBAC, AuthPolicyAdmin, AuthPolicyPublic:
	default:
		return validationError("unsupported auth_policy", ErrInvalidManifest)
	}
	if !validPath(route.Path) {
		return validationError("route path must start with /", ErrInvalidManifest)
	}
	if !validPath(route.UpstreamPath) {
		return validationError("route upstream_path must start with /", ErrInvalidManifest)
	}
	return nil
}

func ManifestHash(manifest Manifest) (string, error) {
	manifest = NormalizeManifest(manifest)
	type hashRoute struct {
		Method            Method            `json:"method"`
		MatchType         MatchType         `json:"match_type"`
		Path              string            `json:"path"`
		UpstreamPath      string            `json:"upstream_path"`
		AuthPolicy        AuthPolicy        `json:"auth_policy"`
		TimeoutMS         int               `json:"timeout_ms"`
		ForwardAuthHeader bool              `json:"forward_auth_header"`
		Metadata          map[string]string `json:"metadata,omitempty"`
	}
	input := struct {
		SchemaVersion string            `json:"schema_version"`
		PluginKey     string            `json:"plugin_key"`
		Name          string            `json:"name"`
		Version       string            `json:"version"`
		Protocol      Protocol          `json:"protocol"`
		HealthPath    string            `json:"health_path"`
		OpenAPIURL    string            `json:"openapi_url,omitempty"`
		Routes        []hashRoute       `json:"routes"`
		Metadata      map[string]string `json:"metadata,omitempty"`
	}{
		SchemaVersion: manifest.SchemaVersion,
		PluginKey:     manifest.PluginKey,
		Name:          manifest.Name,
		Version:       manifest.Version,
		Protocol:      manifest.Protocol,
		HealthPath:    manifest.HealthPath,
		OpenAPIURL:    manifest.OpenAPIURL,
		Metadata:      manifest.Metadata,
	}
	for _, route := range manifest.Routes {
		route = NormalizeRoute(route)
		input.Routes = append(input.Routes, hashRoute(route))
	}
	sort.Slice(input.Routes, func(i, j int) bool {
		left := fmt.Sprintf("%s:%s:%s", input.Routes[i].Method, input.Routes[i].MatchType, input.Routes[i].Path)
		right := fmt.Sprintf("%s:%s:%s", input.Routes[j].Method, input.Routes[j].MatchType, input.Routes[j].Path)
		return left < right
	})
	content, err := json.Marshal(input)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:]), nil
}

func parseServiceURL(raw string) (*url.URL, error) {
	parsed, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, fmt.Errorf("unsupported scheme %q", parsed.Scheme)
	}
	if parsed.Hostname() == "" {
		return nil, fmt.Errorf("host is required")
	}
	if parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return nil, fmt.Errorf("base URL must not include userinfo, query, or fragment")
	}
	return parsed, nil
}

func validPath(path string) bool {
	return strings.HasPrefix(path, "/") && !strings.Contains(path, "\x00")
}
