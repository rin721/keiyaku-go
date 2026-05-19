# pkg/plugin

`pkg/plugin` provides the v2 remote HTTP plugin contract, HMAC helpers, registry client, and heartbeat runner. It does not import any `internal` package.

## What It Does

- Defines manifest v2 and route declarations.
- Validates manifests and computes stable hashes.
- Signs and verifies control-plane and gateway HMAC requests.
- Reads bounded request bodies and verifies signed HTTP requests while restoring `req.Body`.
- Registers, heartbeats, and unregisters plugin instances.
- Provides a heartbeat runner for long-running plugin processes.

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

## Gateway Signature Verification

```go
parts := plugin.SignatureFromHeader(req.Header)
err := plugin.Verify(
	req.Method,
	req.URL.EscapedPath(),
	body,
	parts,
	os.Getenv("KEIYAKU_PLUGIN_GATEWAY_SECRET"),
	time.Now().UTC(),
	plugin.DefaultSignatureSkew,
)
```

For HTTP handlers, prefer the bounded helper:

```go
body, parts, err := plugin.VerifySignedRequest(
	req,
	os.Getenv("KEIYAKU_PLUGIN_GATEWAY_SECRET"),
	10<<20,
	time.Now().UTC(),
	plugin.DefaultSignatureSkew,
)
```

## Heartbeat

```go
runner := plugin.HeartbeatRunner{
	Client:     client,
	PluginKey:  manifest.PluginKey,
	InstanceID: manifest.InstanceID,
	Interval:   10 * time.Second,
}
_ = runner.Run(ctx)
```

The host treats expired leases as unavailable for routing, so plugin services should keep heartbeating while they are able to serve traffic.
