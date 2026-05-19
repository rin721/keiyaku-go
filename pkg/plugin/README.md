# pkg/plugin

`pkg/plugin` provides the v3 remote HTTP plugin contract, HMAC helpers, registry client, and lifecycle runner. It does not import any `internal` package.

## What It Does

- Defines manifest v3 and route declarations.
- Validates manifests and computes stable hashes.
- Signs and verifies control-plane and gateway HMAC requests.
- Reads bounded request bodies and verifies signed HTTP requests while restoring `req.Body`.
- Registers, heartbeats, and unregisters plugin instances.
- Provides a lifecycle runner for register, heartbeat, re-register, backoff, and shutdown unregister.

## Minimal Example

```go
manifest := plugin.Manifest{
	SchemaVersion: plugin.DefaultSchemaVersion,
	PluginKey:     "demo-plugin",
	Name:          "Demo Plugin",
	Version:       "0.1.0",
	InstanceID:    "demo-plugin-1",
	Protocol:      plugin.ProtocolHTTP,
	BaseURL:       "http://127.0.0.1:9090",
	HealthPath:    "/healthz",
	Routes: []plugin.Route{
		{
			RouteID:      "hello",
			Method:       plugin.MethodGet,
			MatchType:    plugin.MatchTypeExact,
			GatewayPath:  "/api/v1/extensions/demo-plugin/hello",
			UpstreamPath: "/hello",
			AuthPolicy:   plugin.AuthPolicyAuthenticated,
			Timeout:      "5s",
		},
	},
}

client := plugin.NewClient(
	"http://127.0.0.1:8080",
	manifest.PluginKey,
	os.Getenv("KEIYAKU_PLUGIN_REGISTRATION_SECRET"),
)
result, err := client.Register(context.Background(), manifest)
```

Long-running plugins should normally use the lifecycle runner instead of hand-rolled register and heartbeat loops:

```go
runner := plugin.LifecycleRunner{
	Client:            client,
	Manifest:          manifest,
	HeartbeatInterval: 10 * time.Second,
	RegisterTimeout:   5 * time.Second,
	UnregisterTimeout: 5 * time.Second,
}
_ = runner.Run(ctx)
```

## Gateway Signature Verification

```go
parts := plugin.SignatureFromHeader(req.Header)
err := plugin.Verify(
	req.Method,
	req.URL.EscapedPath(),
	req.URL.RawQuery,
	body,
	parts,
	os.Getenv("KEIYAKU_PLUGIN_GATEWAY_SECRET"),
	time.Now().UTC(),
	plugin.DefaultSignatureSkew,
)
```

For HTTP handlers, prefer the bounded helper:

```go
body, parts, err := plugin.VerifySignedRequest(req, plugin.VerifyRequestOptions{
	Secret:            os.Getenv("KEIYAKU_PLUGIN_GATEWAY_SECRET"),
	MaxBodyBytes:      10 << 20,
	Now:               time.Now().UTC(),
	Skew:              plugin.DefaultSignatureSkew,
	ExpectedPluginKey: "demo-plugin",
	NonceStore:        plugin.NewMemoryNonceStore(),
})
```

## Heartbeat Compatibility

`HeartbeatRunner` remains available as a lower-level primitive. v3 scaffolds and first-party examples use `LifecycleRunner` so 404 and manifest mismatch responses can trigger automatic re-registration.
