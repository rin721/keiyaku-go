package plugin

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/rin721/keiyaku-go/internal/application/apperror"
	derrors "github.com/rin721/keiyaku-go/internal/domain/errors"
	domainplugin "github.com/rin721/keiyaku-go/internal/domain/plugin"
	pkgplugin "github.com/rin721/keiyaku-go/pkg/plugin"
)

func (s *Service) validateControlSignature(ctx context.Context, pluginKey string, signature SignatureCommand) error {
	if !safeIDPattern.MatchString(pluginKey) || signature.PluginKey != pluginKey {
		return apperror.New(apperror.CodeUnauthorized, apperror.MessageInvalidPluginSignature)
	}
	trust, ok := s.trusted[pluginKey]
	if !ok || strings.TrimSpace(trust.registrationSecret) == "" {
		return apperror.New(apperror.CodeForbidden, apperror.MessagePluginKeyNotTrusted)
	}
	parts := pkgplugin.SignatureParts{
		PluginKey: signature.PluginKey,
		Timestamp: signature.Timestamp,
		Nonce:     signature.Nonce,
		Signature: signature.Signature,
	}
	if err := pkgplugin.Verify(signature.Method, signature.Path, signature.RawQuery, signature.Body, parts, trust.registrationSecret, s.now(), s.config.SignatureSkew); err != nil {
		return apperror.Wrap(apperror.CodeUnauthorized, apperror.MessageInvalidPluginSignature, err)
	}
	timestamp, err := pkgplugin.ParseSignatureTimestamp(signature.Timestamp)
	if err != nil {
		return apperror.Wrap(apperror.CodeUnauthorized, apperror.MessageInvalidPluginSignature, err)
	}
	expiresAt := timestamp.Add(s.config.SignatureSkew)
	if expiresAt.Before(s.now()) {
		expiresAt = s.now().Add(time.Minute)
	}
	if err := s.repo.UseSignatureNonce(ctx, pluginKey, signature.Nonce, expiresAt, s.now()); err != nil {
		if errors.Is(err, derrors.ErrConflict) {
			return apperror.Wrap(apperror.CodeUnauthorized, apperror.MessagePluginNonceReused, err)
		}
		return apperror.Wrap(apperror.CodeDependency, apperror.MessageDependency, err)
	}
	return nil
}

func (s *Service) validateManifestGatewayPaths(manifest pkgplugin.Manifest) error {
	trust, ok := s.trusted[manifest.PluginKey]
	if !ok {
		return fmt.Errorf("plugin key %q is not trusted", manifest.PluginKey)
	}
	for _, route := range manifest.Routes {
		route = pkgplugin.NormalizeRoute(route)
		if err := pkgplugin.ValidateGatewayPath(route.GatewayPath, s.config.PublicPrefix); err != nil {
			return err
		}
		if !gatewayPathAllowed(trust.allowedGatewayPrefixes, route.GatewayPath) {
			return fmt.Errorf("gateway_path %q is outside trusted prefixes for %s", route.GatewayPath, manifest.PluginKey)
		}
		if _, ok := trust.allowedAuthPolicies[domainplugin.AuthPolicy(route.AuthPolicy)]; !ok {
			return fmt.Errorf("auth_policy %q is not trusted for %s", route.AuthPolicy, manifest.PluginKey)
		}
		if _, ok := trust.allowedMethods[domainplugin.Method(route.Method)]; !ok {
			return fmt.Errorf("method %q is not trusted for %s", route.Method, manifest.PluginKey)
		}
	}
	return nil
}

func (s *Service) ValidateOutboundURL(pluginKey string, raw string) error {
	if err := s.validatePluginURL(pluginKey, raw, true); err != nil {
		return apperror.Wrap(apperror.CodeBadGateway, apperror.MessagePluginUpstreamFailed, err)
	}
	return nil
}

func (s *Service) ValidateResolvedOutboundIP(pluginKey string, ip net.IP) error {
	if err := s.validateResolvedOutboundIP(pluginKey, ip); err != nil {
		return apperror.Wrap(apperror.CodeBadGateway, apperror.MessagePluginUpstreamFailed, err)
	}
	return nil
}

func (s *Service) validatePluginURL(pluginKey string, raw string, allowQuery bool) error {
	trust, ok := s.trusted[pluginKey]
	if !ok {
		return fmt.Errorf("plugin key %q is not trusted", pluginKey)
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return err
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("unsupported scheme %q", parsed.Scheme)
	}
	if parsed.Scheme == "http" && !trust.allowInsecureHTTP {
		return fmt.Errorf("insecure http plugin URL is not allowed")
	}
	if parsed.User != nil || parsed.Fragment != "" || (!allowQuery && parsed.RawQuery != "") {
		return fmt.Errorf("plugin URL must not include userinfo, query, or fragment")
	}
	host := parsed.Hostname()
	if host == "" {
		return fmt.Errorf("base_url host is required")
	}
	ip := net.ParseIP(host)
	if ip != nil {
		if ip.IsLoopback() && !trust.allowLoopback {
			return fmt.Errorf("loopback plugin URL is not allowed")
		}
		if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
			return fmt.Errorf("link-local plugin URL is not allowed")
		}
		for _, cidr := range trust.allowedCIDRs {
			if cidr.Contains(ip) {
				return nil
			}
		}
		return fmt.Errorf("ip %q is not in trusted plugin allowed_cidrs", host)
	}
	if hostAllowed(trust.allowedHosts, host) {
		return nil
	}
	return fmt.Errorf("host %q is not in trusted plugin allowed_hosts", host)
}

func (s *Service) validateResolvedOutboundIP(pluginKey string, ip net.IP) error {
	trust, ok := s.trusted[pluginKey]
	if !ok {
		return fmt.Errorf("plugin key %q is not trusted", pluginKey)
	}
	if ip == nil {
		return fmt.Errorf("resolved plugin IP is required")
	}
	if ip.IsLoopback() && !trust.allowLoopback {
		return fmt.Errorf("loopback plugin IP is not allowed")
	}
	if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return fmt.Errorf("link-local plugin IP is not allowed")
	}
	for _, cidr := range trust.allowedCIDRs {
		if cidr.Contains(ip) {
			return nil
		}
	}
	if ip.IsPrivate() {
		return fmt.Errorf("private plugin IP %q is not allowed", ip.String())
	}
	return nil
}

func hostAllowed(allowedHosts []string, host string) bool {
	host = strings.ToLower(strings.TrimSuffix(host, "."))
	for _, allowed := range allowedHosts {
		allowed = strings.ToLower(strings.TrimSpace(strings.TrimSuffix(allowed, ".")))
		if allowed == "" {
			continue
		}
		if host == allowed {
			return true
		}
		if strings.HasPrefix(allowed, "*.") && strings.HasSuffix(host, strings.TrimPrefix(allowed, "*")) {
			return true
		}
	}
	return false
}

func normalizeAllowedGatewayPrefixes(pluginKey string, publicPrefix string, prefixes []string) ([]string, error) {
	if len(prefixes) == 0 {
		prefix := pkgplugin.NormalizePublicPrefix(publicPrefix)
		if prefix == "/" {
			return []string{"/" + pluginKey}, nil
		}
		return []string{strings.TrimRight(prefix, "/") + "/" + pluginKey}, nil
	}
	normalized := make([]string, 0, len(prefixes))
	for _, prefix := range prefixes {
		prefix = pkgplugin.NormalizePublicPrefix(prefix)
		if err := pkgplugin.ValidateGatewayPath(prefix, publicPrefix); err != nil {
			return nil, err
		}
		normalized = append(normalized, prefix)
	}
	return normalized, nil
}

func normalizeAllowedAuthPolicies(values []string) map[domainplugin.AuthPolicy]struct{} {
	if len(values) == 0 {
		values = []string{
			string(domainplugin.AuthPolicyInherit),
			string(domainplugin.AuthPolicyAuthenticated),
			string(domainplugin.AuthPolicyRBAC),
			string(domainplugin.AuthPolicyAdmin),
			string(domainplugin.AuthPolicyPublic),
		}
	}
	allowed := make(map[domainplugin.AuthPolicy]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		allowed[domainplugin.AuthPolicy(value)] = struct{}{}
	}
	return allowed
}

func normalizeAllowedMethods(values []string) map[domainplugin.Method]struct{} {
	if len(values) == 0 {
		values = []string{
			string(domainplugin.MethodAny),
			string(domainplugin.MethodGet),
			string(domainplugin.MethodPost),
			string(domainplugin.MethodPut),
			string(domainplugin.MethodPatch),
			string(domainplugin.MethodDelete),
		}
	}
	allowed := make(map[domainplugin.Method]struct{}, len(values))
	for _, value := range values {
		value = strings.ToUpper(strings.TrimSpace(value))
		if value == "" {
			continue
		}
		allowed[domainplugin.Method(value)] = struct{}{}
	}
	return allowed
}

func gatewayPathAllowed(prefixes []string, gatewayPath string) bool {
	for _, prefix := range prefixes {
		if segmentPathPrefix(gatewayPath, prefix) {
			return true
		}
	}
	return false
}

func segmentPathPrefix(path string, prefix string) bool {
	if prefix == "/" {
		return strings.HasPrefix(path, "/")
	}
	if path == prefix {
		return true
	}
	return strings.HasPrefix(path, strings.TrimRight(prefix, "/")+"/")
}
