package plugin

import (
	"context"
	"strconv"

	"github.com/rin721/keiyaku-go/internal/application/apperror"
	"github.com/rin721/keiyaku-go/internal/application/port"
	domainplugin "github.com/rin721/keiyaku-go/internal/domain/plugin"
)

func (s *Service) CheckHealth(ctx context.Context, probe port.PluginHealthProbe) error {
	if s == nil || s.repo == nil {
		return apperror.New(apperror.CodeInternal, apperror.MessagePluginServiceNotReady)
	}
	if probe == nil || !s.config.Enabled || s.config.HealthCheckInterval <= 0 {
		return nil
	}
	targets, err := s.repo.ListHealthCheckTargets(ctx, s.now())
	if err != nil {
		return mapPluginRepoError(err)
	}
	for _, target := range targets {
		if target == nil {
			continue
		}
		s.checkInstanceHealth(ctx, probe, *target)
	}
	return nil
}

func (s *Service) checkInstanceHealth(ctx context.Context, probe port.PluginHealthProbe, instance domainplugin.Instance) {
	previous := instance.HealthStatus.Normalize()
	err := s.validatePluginURL(instance.PluginKey, instance.BaseURL, false)
	if err == nil {
		err = probe.Probe(ctx, instance)
	}
	now := s.now()
	nextStatus := previous
	failures := instance.ConsecutiveFailures
	lastError := ""
	if err != nil {
		failures++
		lastError = err.Error()
		if failures >= s.unhealthyThreshold() {
			nextStatus = domainplugin.HealthStatusUnhealthy
		}
	} else {
		failures = 0
		nextStatus = domainplugin.HealthStatusHealthy
	}
	updated, updateErr := s.repo.UpdateInstanceHealth(ctx, instance.PluginKey, instance.InstanceID, nextStatus, failures, lastError, now)
	if updateErr != nil || updated == nil {
		return
	}
	current := updated.HealthStatus.Normalize()
	if previous == current {
		return
	}
	s.invalidateRouteCache(instance.PluginKey)
	s.recordAudit(ctx, domainplugin.AuditEvent{
		PluginKey:  instance.PluginKey,
		InstanceID: instance.InstanceID,
		Action:     domainplugin.AuditActionHealthChange,
		Message:    "plugin instance health changed",
		Metadata: map[string]string{
			"from":                 string(previous),
			"to":                   string(current),
			"consecutive_failures": strconv.Itoa(updated.ConsecutiveFailures),
		},
	})
	if s.metrics != nil {
		s.metrics.RecordPluginHealth(ctx, domainplugin.HealthMetric{
			PluginKey:           instance.PluginKey,
			InstanceID:          instance.InstanceID,
			PreviousStatus:      previous,
			CurrentStatus:       current,
			ConsecutiveFailures: updated.ConsecutiveFailures,
		})
	}
}

func (s *Service) unhealthyThreshold() int {
	if s == nil || s.config.UnhealthyThreshold <= 0 {
		return defaultUnhealthyThreshold
	}
	return s.config.UnhealthyThreshold
}
