package metrics

import (
	"context"

	domainplugin "github.com/rin721/keiyaku-go/internal/domain/plugin"
)

type NoopPluginMetrics struct{}

func (NoopPluginMetrics) RecordPluginGateway(context.Context, domainplugin.GatewayMetric) {}

func (NoopPluginMetrics) RecordPluginHealth(context.Context, domainplugin.HealthMetric) {}
