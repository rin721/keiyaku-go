# pkg/plugin

`pkg/plugin` provides the remote plugin service contract and client helpers for Keiyaku-Go. It is intended for independently deployed plugin services and does not import any `internal` package.

## What It Does

- Defines the plugin manifest, route declaration, protocol, match type, and auth policy types.
- Validates manifests before registration.
- Computes a stable manifest hash.
- Registers, heartbeats, and unregisters plugin instances against the host service.
- Provides a heartbeat runner for long-running plugin processes.

## What It Does Not Do

- It does not load Go dynamic plugins.
- It does not provide a runtime DI container.
- It does not implement plugin business logic or persistence.
- It does not bypass host-side route, token, or allow-list validation.

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
			Method:       plugin.MethodGet,
			MatchType:    plugin.MatchTypeExact,
			Path:         "/hello",
			UpstreamPath: "/hello",
			AuthPolicy:   plugin.AuthPolicyAuthenticated,
			TimeoutMS:    5000,
		},
	},
}

client := plugin.NewClient("http://127.0.0.1:8080", os.Getenv("KEIYAKU_PLUGIN_TOKEN"))
result, err := client.Register(context.Background(), manifest)
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
