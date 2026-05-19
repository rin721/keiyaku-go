package plugin

import (
	"context"

	"github.com/rin721/keiyaku-go/internal/application/apperror"
)

func (s *Service) Maintain(ctx context.Context) (*MaintenanceResult, error) {
	if s == nil || s.repo == nil {
		return nil, apperror.New(apperror.CodeInternal, apperror.MessagePluginServiceNotReady)
	}
	if !s.config.Enabled {
		return &MaintenanceResult{}, nil
	}
	now := s.now()
	result := &MaintenanceResult{}
	prunedNonces, err := s.repo.PruneSignatureNonces(ctx, now)
	if err != nil {
		return nil, mapPluginRepoError(err)
	}
	result.PrunedSignatureNonces = prunedNonces
	if s.config.AuditRetentionDays > 0 {
		prunedAudits, err := s.repo.PrunePluginAuditEvents(ctx, now.AddDate(0, 0, -s.config.AuditRetentionDays))
		if err != nil {
			return nil, mapPluginRepoError(err)
		}
		result.PrunedAuditEvents = prunedAudits
	}
	staleBefore := now.Add(-s.config.HeartbeatTTL)
	disabled, err := s.repo.DisableStalePluginInstances(ctx, staleBefore, now)
	if err != nil {
		return nil, mapPluginRepoError(err)
	}
	result.DisabledStaleInstances = disabled
	if disabled > 0 {
		s.invalidateRouteCache("")
	}
	return result, nil
}
