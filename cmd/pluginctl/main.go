package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	cmdcli "github.com/rin721/keiyaku-go/pkg/cli"
	pluginsdk "github.com/rin721/keiyaku-go/pkg/plugin"
)

const (
	appName cmdcli.AppName = "keiyaku-pluginctl"

	commandInit       cmdcli.CommandName = "init"
	commandManifest   cmdcli.CommandName = "manifest"
	commandValidate   cmdcli.CommandName = "validate"
	commandRegister   cmdcli.CommandName = "register"
	commandHeartbeat  cmdcli.CommandName = "heartbeat"
	commandUnregister cmdcli.CommandName = "unregister"

	flagDir        cmdcli.FlagName = "dir"
	flagModule     cmdcli.FlagName = "module"
	flagPluginKey  cmdcli.FlagName = "plugin-key"
	flagName       cmdcli.FlagName = "name"
	flagManifest   cmdcli.FlagName = "manifest"
	flagHost       cmdcli.FlagName = "host"
	flagSecret     cmdcli.FlagName = "registration-secret"
	flagInstanceID cmdcli.FlagName = "instance-id"

	envRegistrationSecret cmdcli.EnvName = "KEIYAKU_PLUGIN_REGISTRATION_SECRET"
)

func main() {
	cmdcli.RunAndExit(context.Background(), newAppSpec(), os.Args)
}

func newAppSpec() cmdcli.AppSpec {
	commonClientFlags := []cmdcli.Flag{
		cmdcli.StringFlag(cmdcli.StringFlagSpec{Name: flagHost, Usage: "Keiyaku-Go host base URL", Default: "http://127.0.0.1:8080"}),
		cmdcli.StringFlag(cmdcli.StringFlagSpec{Name: flagSecret, Usage: "Plugin registration secret", EnvVars: []cmdcli.EnvName{envRegistrationSecret}}),
	}
	return cmdcli.AppSpec{
		Name:                   appName,
		Usage:                  "Manage Keiyaku-Go remote plugins",
		UsageText:              "keiyaku-pluginctl <command> [options]",
		UseShortOptionHandling: true,
		Commands: []cmdcli.CommandSpec{
			{
				Name:      commandInit,
				Usage:     "Create a minimal HTTP plugin service scaffold",
				UsageText: "keiyaku-pluginctl init --plugin-key demo --name Demo --module example.com/demo-plugin [--dir ./demo-plugin]",
				Flags: []cmdcli.Flag{
					cmdcli.StringFlag(cmdcli.StringFlagSpec{Name: flagDir, Usage: "Output directory", Default: "plugin-demo"}),
					cmdcli.StringFlag(cmdcli.StringFlagSpec{Name: flagModule, Usage: "Generated Go module path"}),
					cmdcli.StringFlag(cmdcli.StringFlagSpec{Name: flagPluginKey, Usage: "Plugin key"}),
					cmdcli.StringFlag(cmdcli.StringFlagSpec{Name: flagName, Usage: "Plugin display name"}),
				},
				Action: runInit,
			},
			{
				Name:  commandManifest,
				Usage: "Inspect plugin manifests",
				Commands: []cmdcli.CommandSpec{
					{
						Name:      commandValidate,
						Usage:     "Validate a plugin manifest",
						UsageText: "keiyaku-pluginctl manifest validate --manifest manifest.json",
						Flags: []cmdcli.Flag{
							cmdcli.StringFlag(cmdcli.StringFlagSpec{Name: flagManifest, Usage: "Manifest JSON path", Default: "manifest.json"}),
						},
						Action: runValidate,
					},
				},
			},
			{
				Name:      commandRegister,
				Usage:     "Register a plugin manifest with the host",
				UsageText: "keiyaku-pluginctl register --manifest manifest.json --host http://127.0.0.1:8080",
				Flags: append([]cmdcli.Flag{
					cmdcli.StringFlag(cmdcli.StringFlagSpec{Name: flagManifest, Usage: "Manifest JSON path", Default: "manifest.json"}),
				}, commonClientFlags...),
				Action: runRegister,
			},
			{
				Name:      commandHeartbeat,
				Usage:     "Send one heartbeat for a plugin instance",
				UsageText: "keiyaku-pluginctl heartbeat --plugin-key demo --instance-id demo-1",
				Flags: append([]cmdcli.Flag{
					cmdcli.StringFlag(cmdcli.StringFlagSpec{Name: flagPluginKey, Usage: "Plugin key"}),
					cmdcli.StringFlag(cmdcli.StringFlagSpec{Name: flagInstanceID, Usage: "Instance ID"}),
				}, commonClientFlags...),
				Action: runHeartbeat,
			},
			{
				Name:      commandUnregister,
				Usage:     "Disable a plugin instance registration",
				UsageText: "keiyaku-pluginctl unregister --plugin-key demo --instance-id demo-1",
				Flags: append([]cmdcli.Flag{
					cmdcli.StringFlag(cmdcli.StringFlagSpec{Name: flagPluginKey, Usage: "Plugin key"}),
					cmdcli.StringFlag(cmdcli.StringFlagSpec{Name: flagInstanceID, Usage: "Instance ID"}),
				}, commonClientFlags...),
				Action: runUnregister,
			},
		},
	}
}

func runInit(ctx context.Context, cliCtx *cmdcli.Context) error {
	_ = ctx
	dir := strings.TrimSpace(cliCtx.String(flagDir))
	module := strings.TrimSpace(cliCtx.String(flagModule))
	key := strings.TrimSpace(cliCtx.String(flagPluginKey))
	name := strings.TrimSpace(cliCtx.String(flagName))
	if dir == "" || module == "" || key == "" || name == "" {
		return cmdcli.UsageError("--dir、--module、--plugin-key 和 --name 都是必填项")
	}
	replacePath, err := localModuleReplacePath(dir)
	if err != nil {
		return cmdcli.WrapRuntimeError(cmdcli.OperationAction, "resolve local module path", err)
	}
	files := map[string]string{
		"go.mod":        scaffoldGoMod(module, replacePath),
		"main.go":       scaffoldMain(module, key, name),
		"manifest.json": scaffoldManifest(key, name),
	}
	for path, content := range files {
		target := filepath.Join(dir, path)
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return cmdcli.WrapRuntimeError(cmdcli.OperationAction, "create plugin scaffold directory", err)
		}
		if _, err := os.Stat(target); err == nil {
			return cmdcli.UsageError("%s 已存在", target)
		}
		if err := os.WriteFile(target, []byte(content), 0o644); err != nil {
			return cmdcli.WrapRuntimeError(cmdcli.OperationAction, "write plugin scaffold", err)
		}
	}
	cliCtx.UI().Successf("Plugin scaffold created: %s", dir)
	return nil
}

func runValidate(ctx context.Context, cliCtx *cmdcli.Context) error {
	_ = ctx
	manifest, err := readManifest(cliCtx.String(flagManifest))
	if err != nil {
		return err
	}
	if err := pluginsdk.ValidateManifest(manifest); err != nil {
		return cmdcli.WrapRuntimeError(cmdcli.OperationAction, "validate plugin manifest", err)
	}
	hash, err := pluginsdk.ManifestHash(manifest)
	if err != nil {
		return cmdcli.WrapRuntimeError(cmdcli.OperationAction, "hash plugin manifest", err)
	}
	cliCtx.UI().Successf("Plugin manifest is valid: %s", hash)
	return nil
}

func runRegister(ctx context.Context, cliCtx *cmdcli.Context) error {
	manifest, err := readManifest(cliCtx.String(flagManifest))
	if err != nil {
		return err
	}
	client, err := newClient(cliCtx, manifest.PluginKey)
	if err != nil {
		return err
	}
	result, err := client.Register(ctx, manifest)
	if err != nil {
		return cmdcli.WrapRuntimeError(cmdcli.OperationAction, "register plugin", err)
	}
	cliCtx.UI().Successf("Plugin registered: %s/%s lease=%s", result.PluginKey, result.InstanceID, result.LeaseUntil.Format(time.RFC3339))
	return nil
}

func runHeartbeat(ctx context.Context, cliCtx *cmdcli.Context) error {
	pluginKey := requireString(cliCtx, flagPluginKey)
	client, err := newClient(cliCtx, pluginKey)
	if err != nil {
		return err
	}
	result, err := client.Heartbeat(ctx, pluginKey, requireString(cliCtx, flagInstanceID))
	if err != nil {
		return cmdcli.WrapRuntimeError(cmdcli.OperationAction, "heartbeat plugin", err)
	}
	cliCtx.UI().Successf("Plugin heartbeat accepted: %s/%s lease=%s", result.PluginKey, result.InstanceID, result.LeaseUntil.Format(time.RFC3339))
	return nil
}

func runUnregister(ctx context.Context, cliCtx *cmdcli.Context) error {
	pluginKey := requireString(cliCtx, flagPluginKey)
	client, err := newClient(cliCtx, pluginKey)
	if err != nil {
		return err
	}
	if err := client.Unregister(ctx, pluginKey, requireString(cliCtx, flagInstanceID)); err != nil {
		return cmdcli.WrapRuntimeError(cmdcli.OperationAction, "unregister plugin", err)
	}
	cliCtx.UI().Success("Plugin instance unregistered")
	return nil
}

func readManifest(path string) (pluginsdk.Manifest, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return pluginsdk.Manifest{}, cmdcli.WrapRuntimeError(cmdcli.OperationAction, "read plugin manifest", err)
	}
	var manifest pluginsdk.Manifest
	decoder := json.NewDecoder(strings.NewReader(string(content)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&manifest); err != nil {
		return pluginsdk.Manifest{}, cmdcli.WrapRuntimeError(cmdcli.OperationAction, "decode plugin manifest", err)
	}
	return pluginsdk.NormalizeManifest(manifest), nil
}

func newClient(cliCtx *cmdcli.Context, pluginKey string) (*pluginsdk.Client, error) {
	secret := strings.TrimSpace(cliCtx.String(flagSecret))
	if secret == "" {
		return nil, cmdcli.UsageError("--%s 或 %s 是必填项", flagSecret, envRegistrationSecret)
	}
	return pluginsdk.NewClient(strings.TrimSpace(cliCtx.String(flagHost)), pluginKey, secret), nil
}

func requireString(cliCtx *cmdcli.Context, flag cmdcli.FlagName) string {
	return strings.TrimSpace(cliCtx.String(flag))
}

func localModuleReplacePath(dir string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(absDir, cwd)
	if err != nil {
		return "", err
	}
	return filepath.ToSlash(rel), nil
}

func scaffoldGoMod(module string, replacePath string) string {
	return fmt.Sprintf("module %s\n\ngo 1.25.4\n\nrequire github.com/rin721/keiyaku-go v0.0.0\n\nreplace github.com/rin721/keiyaku-go => %s\n", module, replacePath)
}

func scaffoldMain(module string, key string, name string) string {
	_ = module
	return fmt.Sprintf(`package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	pluginsdk "github.com/rin721/keiyaku-go/pkg/plugin"
)

func main() {
	addr := env("PLUGIN_ADDR", ":9090")
	host := env("KEIYAKU_HOST", "http://127.0.0.1:8080")
	registrationSecret := os.Getenv("KEIYAKU_PLUGIN_REGISTRATION_SECRET")
	gatewaySecret := os.Getenv("KEIYAKU_PLUGIN_GATEWAY_SECRET")
	baseURL := env("PLUGIN_BASE_URL", "http://127.0.0.1"+addr)
	if strings.TrimSpace(registrationSecret) == "" {
		log.Fatal("KEIYAKU_PLUGIN_REGISTRATION_SECRET is required")
	}
	if strings.TrimSpace(gatewaySecret) == "" {
		log.Fatal("KEIYAKU_PLUGIN_GATEWAY_SECRET is required")
	}

	manifest := manifest(baseURL)
	client := pluginsdk.NewClient(host, manifest.PluginKey, registrationSecret)
	ctx := context.Background()
	if _, err := client.Register(ctx, manifest); err != nil {
		log.Printf("register plugin: %%v", err)
	}
	go func() {
		runner := pluginsdk.HeartbeatRunner{
			Client: client,
			PluginKey: manifest.PluginKey,
			InstanceID: manifest.InstanceID,
			Interval: 10 * time.Second,
			OnError: func(err error) { log.Printf("plugin heartbeat: %%v", err) },
		}
		_ = runner.Run(ctx)
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		if !verifyGateway(w, r, gatewaySecret) {
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{
			"plugin": r.Header.Get("X-Keiyaku-Plugin-Key"),
			"trace_id": r.Header.Get("X-Trace-ID"),
			"message": "hello from %s",
		})
	})
	log.Printf("plugin listening on %%s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func verifyGateway(w http.ResponseWriter, r *http.Request, secret string) bool {
	if _, _, err := pluginsdk.VerifySignedRequest(r, secret, 10<<20, time.Now().UTC(), pluginsdk.DefaultSignatureSkew); err != nil {
		status := http.StatusUnauthorized
		msg := "invalid gateway signature"
		if errors.Is(err, pluginsdk.ErrBodyTooLarge) {
			status = http.StatusRequestEntityTooLarge
			msg = "request body too large"
		}
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
		return false
	}
	return true
}

func manifest(baseURL string) pluginsdk.Manifest {
	return pluginsdk.Manifest{
		SchemaVersion: pluginsdk.DefaultSchemaVersion,
		PluginKey: "%s",
		Name: "%s",
		Version: "0.1.0",
		InstanceID: "%s-local",
		Protocol: pluginsdk.ProtocolHTTP,
		BaseURL: baseURL,
		HealthPath: "/healthz",
		Routes: []pluginsdk.Route{
			{
				RouteID: "hello",
				Method: pluginsdk.MethodGet,
				MatchType: pluginsdk.MatchTypeExact,
				GatewayPath: "/api/v1/extensions/%s/hello",
				UpstreamPath: "/hello",
				AuthPolicy: pluginsdk.AuthPolicyAuthenticated,
				Timeout: "5s",
			},
		},
	}
}

func env(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
`, name, key, name, key, key)
}

func scaffoldManifest(key string, name string) string {
	return fmt.Sprintf(`{
  "schema_version": "v2",
  "plugin_key": "%s",
  "name": "%s",
  "version": "0.1.0",
  "instance_id": "%s-local",
  "protocol": "http",
  "base_url": "http://127.0.0.1:9090",
  "health_path": "/healthz",
  "routes": [
    {
      "route_id": "hello",
      "method": "GET",
      "match_type": "exact",
      "gateway_path": "/api/v1/extensions/%s/hello",
      "upstream_path": "/hello",
      "auth_policy": "authenticated",
      "timeout": "5s",
      "forward_auth_header": false
    }
  ]
}
`, key, name, key, key)
}
