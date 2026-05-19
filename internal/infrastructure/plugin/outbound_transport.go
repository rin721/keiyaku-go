package plugin

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"
)

type URLValidator func(pluginKey string, raw string) error
type ResolvedIPValidator func(pluginKey string, ip net.IP) error

type outboundGuardContextKey struct{}

type outboundGuardContext struct {
	pluginKey string
}

type OutboundGuardTransport struct {
	transport   *http.Transport
	resolver    *net.Resolver
	dialer      *net.Dialer
	validateURL URLValidator
	validateIP  ResolvedIPValidator
}

func NewOutboundHTTPClient(timeout time.Duration, validateURL URLValidator, validateIP ResolvedIPValidator) *http.Client {
	return &http.Client{
		Timeout:   timeout,
		Transport: NewOutboundGuardTransport(validateURL, validateIP),
	}
}

func NewOutboundGuardTransport(validateURL URLValidator, validateIP ResolvedIPValidator) *OutboundGuardTransport {
	base := http.DefaultTransport.(*http.Transport).Clone()
	guard := &OutboundGuardTransport{
		transport:   base,
		resolver:    net.DefaultResolver,
		dialer:      &net.Dialer{},
		validateURL: validateURL,
		validateIP:  validateIP,
	}
	base.DialContext = guard.dialContext
	return guard
}

func (t *OutboundGuardTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req == nil || req.URL == nil {
		return nil, fmt.Errorf("plugin outbound request is nil")
	}
	pluginKey := req.Header.Get("X-Keiyaku-Plugin-Key")
	if t != nil && t.validateURL != nil {
		if err := t.validateURL(pluginKey, req.URL.String()); err != nil {
			return nil, err
		}
	}
	ctx := context.WithValue(req.Context(), outboundGuardContextKey{}, outboundGuardContext{pluginKey: pluginKey})
	return t.transport.RoundTrip(req.WithContext(ctx))
}

func (t *OutboundGuardTransport) dialContext(ctx context.Context, network string, address string) (net.Conn, error) {
	if t == nil {
		return (&net.Dialer{}).DialContext(ctx, network, address)
	}
	guard, _ := ctx.Value(outboundGuardContextKey{}).(outboundGuardContext)
	if err := t.validateResolvedAddress(ctx, guard.pluginKey, address); err != nil {
		return nil, err
	}
	return t.dialer.DialContext(ctx, network, address)
}

func (t *OutboundGuardTransport) validateResolvedAddress(ctx context.Context, pluginKey string, address string) error {
	if t.validateIP == nil {
		return nil
	}
	host, port, err := net.SplitHostPort(address)
	if err != nil || port == "" {
		host = address
	}
	if ip := net.ParseIP(host); ip != nil {
		return t.validateIP(pluginKey, ip)
	}
	resolver := t.resolver
	if resolver == nil {
		resolver = net.DefaultResolver
	}
	ips, err := resolver.LookupIPAddr(ctx, host)
	if err != nil {
		return err
	}
	for _, resolved := range ips {
		if err := t.validateIP(pluginKey, resolved.IP); err != nil {
			return err
		}
	}
	return nil
}
