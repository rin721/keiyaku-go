package mysql

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	derrors "github.com/rin721/keiyaku-go/internal/domain/errors"
	domainplugin "github.com/rin721/keiyaku-go/internal/domain/plugin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PluginServiceModel struct {
	ID                  int64      `gorm:"column:id;primaryKey;autoIncrement"`
	PluginKey           string     `gorm:"column:plugin_key"`
	Name                string     `gorm:"column:name"`
	Protocol            string     `gorm:"column:protocol"`
	CurrentManifestHash string     `gorm:"column:current_manifest_hash"`
	Status              string     `gorm:"column:status"`
	MetadataJSON        string     `gorm:"column:metadata_json"`
	CreatedAt           time.Time  `gorm:"column:created_at"`
	UpdatedAt           time.Time  `gorm:"column:updated_at"`
	DisabledAt          *time.Time `gorm:"column:disabled_at"`
}

func (PluginServiceModel) TableName() string {
	return "plugin_services"
}

type PluginInstanceModel struct {
	ID                  int64      `gorm:"column:id;primaryKey;autoIncrement"`
	PluginKey           string     `gorm:"column:plugin_key"`
	InstanceID          string     `gorm:"column:instance_id"`
	Version             string     `gorm:"column:version"`
	BaseURL             string     `gorm:"column:base_url"`
	HealthPath          string     `gorm:"column:health_path"`
	ManifestHash        string     `gorm:"column:manifest_hash"`
	Status              string     `gorm:"column:status"`
	HealthStatus        string     `gorm:"column:health_status"`
	LastSeenAt          time.Time  `gorm:"column:last_seen_at"`
	LeaseExpiresAt      time.Time  `gorm:"column:lease_expires_at"`
	LastCheckedAt       *time.Time `gorm:"column:last_checked_at"`
	ConsecutiveFailures int        `gorm:"column:consecutive_failures"`
	LastError           string     `gorm:"column:last_error"`
	LastErrorAt         *time.Time `gorm:"column:last_error_at"`
	CreatedAt           time.Time  `gorm:"column:created_at"`
	UpdatedAt           time.Time  `gorm:"column:updated_at"`
}

func (PluginInstanceModel) TableName() string {
	return "plugin_instances"
}

type PluginRouteModel struct {
	ID                int64     `gorm:"column:id;primaryKey;autoIncrement"`
	PluginKey         string    `gorm:"column:plugin_key"`
	ManifestHash      string    `gorm:"column:manifest_hash"`
	RouteID           string    `gorm:"column:route_id"`
	Method            string    `gorm:"column:method"`
	MatchType         string    `gorm:"column:match_type"`
	GatewayPath       string    `gorm:"column:gateway_path"`
	UpstreamPath      string    `gorm:"column:upstream_path"`
	AuthPolicy        string    `gorm:"column:auth_policy"`
	Timeout           string    `gorm:"column:timeout"`
	ForwardAuthHeader bool      `gorm:"column:forward_auth_header"`
	Enabled           bool      `gorm:"column:enabled"`
	MetadataJSON      string    `gorm:"column:metadata_json"`
	CreatedAt         time.Time `gorm:"column:created_at"`
	UpdatedAt         time.Time `gorm:"column:updated_at"`
}

func (PluginRouteModel) TableName() string {
	return "plugin_routes"
}

type PluginSignatureNonceModel struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement"`
	PluginKey string    `gorm:"column:plugin_key"`
	Nonce     string    `gorm:"column:nonce"`
	ExpiresAt time.Time `gorm:"column:expires_at"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (PluginSignatureNonceModel) TableName() string {
	return "plugin_signature_nonces"
}

type PluginAuditEventModel struct {
	ID           int64     `gorm:"column:id;primaryKey;autoIncrement"`
	PluginKey    string    `gorm:"column:plugin_key"`
	InstanceID   string    `gorm:"column:instance_id"`
	Action       string    `gorm:"column:action"`
	Message      string    `gorm:"column:message"`
	MetadataJSON string    `gorm:"column:metadata_json"`
	CreatedAt    time.Time `gorm:"column:created_at"`
}

func (PluginAuditEventModel) TableName() string {
	return "plugin_audit_events"
}

type PluginRegistryRepository struct {
	db *gorm.DB
}

func NewPluginRegistryRepository(db *gorm.DB) *PluginRegistryRepository {
	return &PluginRegistryRepository{db: db}
}

func (r *PluginRegistryRepository) RecordPluginAudit(ctx context.Context, event domainplugin.AuditEvent) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("plugin audit repository is not ready")
	}
	model := pluginAuditEventToModel(&event)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return fmt.Errorf("record plugin audit event: %w", err)
	}
	return nil
}

func (r *PluginRegistryRepository) ListPluginAuditEvents(ctx context.Context, pluginKey string, limit int) ([]*domainplugin.AuditEvent, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("plugin audit repository is not ready")
	}
	if limit <= 0 {
		limit = 100
	}
	var models []PluginAuditEventModel
	if err := r.db.WithContext(ctx).
		Where("plugin_key = ?", pluginKey).
		Order("created_at DESC, id DESC").
		Limit(limit).
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("list plugin audit events: %w", err)
	}
	events := make([]*domainplugin.AuditEvent, 0, len(models))
	for i := range models {
		event, err := pluginAuditEventFromModel(&models[i])
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, nil
}

func (r *PluginRegistryRepository) UpsertRegistration(ctx context.Context, registration domainplugin.Registration) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("plugin registry repository is not ready")
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := registration.Service.UpdatedAt
		var existing PluginServiceModel
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("plugin_key = ?", registration.Service.PluginKey).
			First(&existing).Error
		switch {
		case err == nil:
			existing.Name = registration.Service.Name
			existing.Protocol = string(registration.Service.Protocol)
			existing.CurrentManifestHash = registration.Service.CurrentManifestHash
			existing.Status = string(domainplugin.ServiceStatusActive)
			existing.MetadataJSON = marshalStringMap(registration.Service.Metadata)
			existing.UpdatedAt = now
			existing.DisabledAt = nil
			if err := tx.Save(&existing).Error; err != nil {
				return fmt.Errorf("update plugin service: %w", err)
			}
		case IsNotFound(err):
			model := pluginServiceToModel(&registration.Service)
			if err := tx.Create(model).Error; err != nil {
				return fmt.Errorf("create plugin service: %w", err)
			}
		default:
			return fmt.Errorf("load plugin service: %w", err)
		}

		instance := pluginInstanceToModel(&registration.Instance)
		var existingInstance PluginInstanceModel
		err = tx.Where("plugin_key = ? AND instance_id = ?", instance.PluginKey, instance.InstanceID).First(&existingInstance).Error
		switch {
		case err == nil:
			instance.ID = existingInstance.ID
			instance.CreatedAt = existingInstance.CreatedAt
			if err := tx.Save(instance).Error; err != nil {
				return fmt.Errorf("update plugin instance: %w", err)
			}
		case IsNotFound(err):
			if err := tx.Create(instance).Error; err != nil {
				return fmt.Errorf("create plugin instance: %w", err)
			}
		default:
			return fmt.Errorf("load plugin instance: %w", err)
		}

		if err := tx.Model(&PluginInstanceModel{}).
			Where("plugin_key = ? AND manifest_hash <> ? AND status = ?", registration.Service.PluginKey, registration.Service.CurrentManifestHash, string(domainplugin.InstanceStatusActive)).
			Updates(map[string]interface{}{"status": string(domainplugin.InstanceStatusIncompatible), "updated_at": now}).Error; err != nil {
			return fmt.Errorf("mark incompatible plugin instances: %w", err)
		}
		if err := tx.Where("plugin_key = ?", registration.Service.PluginKey).Delete(&PluginRouteModel{}).Error; err != nil {
			return fmt.Errorf("delete plugin routes: %w", err)
		}
		for i := range registration.Routes {
			model := pluginRouteToModel(&registration.Routes[i])
			if err := tx.Create(model).Error; err != nil {
				return fmt.Errorf("create plugin route: %w", err)
			}
		}
		return nil
	})
}

func (r *PluginRegistryRepository) FindRouteConflict(ctx context.Context, pluginKey string, routes []domainplugin.Route) (*domainplugin.RouteConflict, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("plugin registry repository is not ready")
	}
	for i := range routes {
		route := routes[i]
		var model PluginRouteModel
		err := r.db.WithContext(ctx).
			Table("plugin_routes AS r").
			Select("r.*").
			Joins("JOIN plugin_services AS s ON s.plugin_key = r.plugin_key AND s.current_manifest_hash = r.manifest_hash").
			Where("s.status = ? AND r.enabled = ? AND r.plugin_key <> ? AND r.method = ? AND r.match_type = ? AND r.gateway_path = ?",
				string(domainplugin.ServiceStatusActive), true, pluginKey, string(route.Method), string(route.MatchType), route.GatewayPath).
			Order("r.plugin_key ASC, r.route_id ASC").
			First(&model).Error
		switch {
		case err == nil:
			return &domainplugin.RouteConflict{
				PluginKey:   model.PluginKey,
				RouteID:     model.RouteID,
				Method:      domainplugin.Method(model.Method),
				MatchType:   domainplugin.MatchType(model.MatchType),
				GatewayPath: model.GatewayPath,
			}, nil
		case IsNotFound(err):
			continue
		default:
			return nil, fmt.Errorf("find plugin route conflict: %w", err)
		}
	}
	return nil, nil
}

func (r *PluginRegistryRepository) TouchInstance(ctx context.Context, pluginKey string, instanceID string, leaseExpiresAt time.Time, now time.Time) (*domainplugin.Instance, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("plugin registry repository is not ready")
	}
	var model PluginInstanceModel
	err := r.db.WithContext(ctx).Where("plugin_key = ? AND instance_id = ?", pluginKey, instanceID).First(&model).Error
	if err != nil {
		if IsNotFound(err) {
			return nil, derrors.ErrNotFound
		}
		return nil, fmt.Errorf("load plugin instance: %w", err)
	}
	if model.Status == string(domainplugin.InstanceStatusDisabled) {
		return nil, derrors.ErrNotFound
	}
	model.Status = string(domainplugin.InstanceStatusActive)
	model.LastSeenAt = now
	model.LeaseExpiresAt = leaseExpiresAt
	model.UpdatedAt = now
	if err := r.db.WithContext(ctx).Save(&model).Error; err != nil {
		return nil, fmt.Errorf("touch plugin instance: %w", err)
	}
	return pluginInstanceFromModel(&model), nil
}

func (r *PluginRegistryRepository) DisableInstance(ctx context.Context, pluginKey string, instanceID string, now time.Time) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("plugin registry repository is not ready")
	}
	result := r.db.WithContext(ctx).Model(&PluginInstanceModel{}).
		Where("plugin_key = ? AND instance_id = ?", pluginKey, instanceID).
		Updates(map[string]interface{}{"status": string(domainplugin.InstanceStatusDisabled), "updated_at": now})
	if result.Error != nil {
		return fmt.Errorf("disable plugin instance: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return derrors.ErrNotFound
	}
	return nil
}

func (r *PluginRegistryRepository) SetServiceStatus(ctx context.Context, pluginKey string, status domainplugin.ServiceStatus, now time.Time) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("plugin registry repository is not ready")
	}
	updates := map[string]interface{}{
		"status":     string(status),
		"updated_at": now,
	}
	if status == domainplugin.ServiceStatusDisabled {
		updates["disabled_at"] = now
	} else {
		updates["disabled_at"] = nil
	}
	result := r.db.WithContext(ctx).Model(&PluginServiceModel{}).
		Where("plugin_key = ?", pluginKey).
		Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("set plugin service status: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return derrors.ErrNotFound
	}
	return nil
}

func (r *PluginRegistryRepository) SetInstanceStatus(ctx context.Context, pluginKey string, instanceID string, status domainplugin.InstanceStatus, now time.Time) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("plugin registry repository is not ready")
	}
	result := r.db.WithContext(ctx).Model(&PluginInstanceModel{}).
		Where("plugin_key = ? AND instance_id = ?", pluginKey, instanceID).
		Updates(map[string]interface{}{"status": string(status), "updated_at": now})
	if result.Error != nil {
		return fmt.Errorf("set plugin instance status: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return derrors.ErrNotFound
	}
	return nil
}

func (r *PluginRegistryRepository) ListPluginServices(ctx context.Context) ([]*domainplugin.Service, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("plugin registry repository is not ready")
	}
	var models []PluginServiceModel
	if err := r.db.WithContext(ctx).Order("plugin_key ASC").Find(&models).Error; err != nil {
		return nil, fmt.Errorf("list plugin services: %w", err)
	}
	items := make([]*domainplugin.Service, 0, len(models))
	for i := range models {
		service, err := pluginServiceFromModel(&models[i])
		if err != nil {
			return nil, err
		}
		items = append(items, service)
	}
	return items, nil
}

func (r *PluginRegistryRepository) ListPluginInstances(ctx context.Context, pluginKey string) ([]*domainplugin.Instance, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("plugin registry repository is not ready")
	}
	return r.findInstances(ctx, pluginKey, "")
}

func (r *PluginRegistryRepository) ListHealthCheckTargets(ctx context.Context, now time.Time) ([]*domainplugin.Instance, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("plugin registry repository is not ready")
	}
	var models []PluginInstanceModel
	if err := r.db.WithContext(ctx).
		Table("plugin_instances AS i").
		Select("i.*").
		Joins("JOIN plugin_services AS s ON s.plugin_key = i.plugin_key AND s.current_manifest_hash = i.manifest_hash").
		Where("s.status = ? AND i.status = ? AND i.lease_expires_at >= ?", string(domainplugin.ServiceStatusActive), string(domainplugin.InstanceStatusActive), now).
		Order("i.plugin_key ASC, i.instance_id ASC").
		Scan(&models).Error; err != nil {
		return nil, fmt.Errorf("list plugin health check targets: %w", err)
	}
	instances := make([]*domainplugin.Instance, 0, len(models))
	for i := range models {
		instances = append(instances, pluginInstanceFromModel(&models[i]))
	}
	return instances, nil
}

func (r *PluginRegistryRepository) GetPluginService(ctx context.Context, pluginKey string) (*domainplugin.Service, []*domainplugin.Instance, []*domainplugin.Route, error) {
	if r == nil || r.db == nil {
		return nil, nil, nil, fmt.Errorf("plugin registry repository is not ready")
	}
	service, err := r.findService(ctx, pluginKey)
	if err != nil {
		return nil, nil, nil, err
	}
	instances, err := r.findInstances(ctx, pluginKey, "")
	if err != nil {
		return nil, nil, nil, err
	}
	routes, err := r.findRoutes(ctx, pluginKey, service.CurrentManifestHash)
	if err != nil {
		return nil, nil, nil, err
	}
	return service, instances, routes, nil
}

func (r *PluginRegistryRepository) UpdateInstanceHealth(ctx context.Context, pluginKey string, instanceID string, healthStatus domainplugin.HealthStatus, consecutiveFailures int, lastError string, checkedAt time.Time) (*domainplugin.Instance, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("plugin registry repository is not ready")
	}
	updates := map[string]interface{}{
		"health_status":        string(healthStatus.Normalize()),
		"last_checked_at":      checkedAt,
		"consecutive_failures": consecutiveFailures,
		"last_error":           truncateString(lastError, 512),
		"updated_at":           checkedAt,
	}
	if lastError == "" {
		updates["last_error_at"] = nil
	} else {
		updates["last_error_at"] = checkedAt
	}
	result := r.db.WithContext(ctx).Model(&PluginInstanceModel{}).
		Where("plugin_key = ? AND instance_id = ?", pluginKey, instanceID).
		Updates(updates)
	if result.Error != nil {
		return nil, fmt.Errorf("update plugin instance health: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, derrors.ErrNotFound
	}
	var model PluginInstanceModel
	if err := r.db.WithContext(ctx).Where("plugin_key = ? AND instance_id = ?", pluginKey, instanceID).First(&model).Error; err != nil {
		return nil, fmt.Errorf("reload plugin instance health: %w", err)
	}
	return pluginInstanceFromModel(&model), nil
}

func (r *PluginRegistryRepository) UseSignatureNonce(ctx context.Context, pluginKey string, nonce string, expiresAt time.Time, now time.Time) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("plugin registry repository is not ready")
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("expires_at < ?", now).Delete(&PluginSignatureNonceModel{}).Error; err != nil {
			return fmt.Errorf("prune plugin signature nonces: %w", err)
		}
		model := &PluginSignatureNonceModel{PluginKey: pluginKey, Nonce: nonce, ExpiresAt: expiresAt, CreatedAt: now}
		if err := tx.Create(model).Error; err != nil {
			if isDuplicate(err) {
				return derrors.ErrConflict
			}
			return fmt.Errorf("record plugin signature nonce: %w", err)
		}
		return nil
	})
}

func (r *PluginRegistryRepository) PruneSignatureNonces(ctx context.Context, now time.Time) (int64, error) {
	if r == nil || r.db == nil {
		return 0, fmt.Errorf("plugin registry repository is not ready")
	}
	result := r.db.WithContext(ctx).Where("expires_at < ?", now).Delete(&PluginSignatureNonceModel{})
	if result.Error != nil {
		return 0, fmt.Errorf("prune plugin signature nonces: %w", result.Error)
	}
	return result.RowsAffected, nil
}

func (r *PluginRegistryRepository) PrunePluginAuditEvents(ctx context.Context, before time.Time) (int64, error) {
	if r == nil || r.db == nil {
		return 0, fmt.Errorf("plugin audit repository is not ready")
	}
	result := r.db.WithContext(ctx).Where("created_at < ?", before).Delete(&PluginAuditEventModel{})
	if result.Error != nil {
		return 0, fmt.Errorf("prune plugin audit events: %w", result.Error)
	}
	return result.RowsAffected, nil
}

func (r *PluginRegistryRepository) DisableStalePluginInstances(ctx context.Context, staleBefore time.Time, now time.Time) (int64, error) {
	if r == nil || r.db == nil {
		return 0, fmt.Errorf("plugin registry repository is not ready")
	}
	result := r.db.WithContext(ctx).Model(&PluginInstanceModel{}).
		Where("status = ? AND lease_expires_at < ?", string(domainplugin.InstanceStatusActive), staleBefore).
		Updates(map[string]interface{}{"status": string(domainplugin.InstanceStatusDisabled), "updated_at": now})
	if result.Error != nil {
		return 0, fmt.Errorf("disable stale plugin instances: %w", result.Error)
	}
	return result.RowsAffected, nil
}

func (r *PluginRegistryRepository) FindRoutable(ctx context.Context, now time.Time) ([]*domainplugin.Service, []*domainplugin.Instance, []*domainplugin.Route, error) {
	if r == nil || r.db == nil {
		return nil, nil, nil, fmt.Errorf("plugin registry repository is not ready")
	}
	var serviceModels []PluginServiceModel
	if err := r.db.WithContext(ctx).
		Where("status = ?", string(domainplugin.ServiceStatusActive)).
		Order("plugin_key ASC").
		Find(&serviceModels).Error; err != nil {
		return nil, nil, nil, fmt.Errorf("find routable plugin services: %w", err)
	}
	var instanceModels []PluginInstanceModel
	if err := r.db.WithContext(ctx).
		Table("plugin_instances AS i").
		Select("i.*").
		Joins("JOIN plugin_services AS s ON s.plugin_key = i.plugin_key AND s.current_manifest_hash = i.manifest_hash").
		Where("s.status = ? AND i.status = ? AND i.health_status IN ? AND i.lease_expires_at >= ?",
			string(domainplugin.ServiceStatusActive), string(domainplugin.InstanceStatusActive), []string{string(domainplugin.HealthStatusUnknown), string(domainplugin.HealthStatusHealthy)}, now).
		Order("i.plugin_key ASC, i.instance_id ASC").
		Find(&instanceModels).Error; err != nil {
		return nil, nil, nil, fmt.Errorf("find routable plugin instances: %w", err)
	}
	var routeModels []PluginRouteModel
	if err := r.db.WithContext(ctx).
		Table("plugin_routes AS r").
		Select("r.*").
		Joins("JOIN plugin_services AS s ON s.plugin_key = r.plugin_key AND s.current_manifest_hash = r.manifest_hash").
		Where("s.status = ? AND r.enabled = ?", string(domainplugin.ServiceStatusActive), true).
		Order("r.gateway_path DESC").
		Find(&routeModels).Error; err != nil {
		return nil, nil, nil, fmt.Errorf("find routable plugin routes: %w", err)
	}
	services := make([]*domainplugin.Service, 0, len(serviceModels))
	for i := range serviceModels {
		service, err := pluginServiceFromModel(&serviceModels[i])
		if err != nil {
			return nil, nil, nil, err
		}
		services = append(services, service)
	}
	instances := make([]*domainplugin.Instance, 0, len(instanceModels))
	for i := range instanceModels {
		instances = append(instances, pluginInstanceFromModel(&instanceModels[i]))
	}
	routes := make([]*domainplugin.Route, 0, len(routeModels))
	for i := range routeModels {
		route, err := pluginRouteFromModel(&routeModels[i])
		if err != nil {
			return nil, nil, nil, err
		}
		routes = append(routes, route)
	}
	return services, instances, routes, nil
}

func (r *PluginRegistryRepository) findService(ctx context.Context, pluginKey string) (*domainplugin.Service, error) {
	var model PluginServiceModel
	if err := r.db.WithContext(ctx).Where("plugin_key = ?", pluginKey).First(&model).Error; err != nil {
		if IsNotFound(err) {
			return nil, derrors.ErrNotFound
		}
		return nil, fmt.Errorf("load plugin service: %w", err)
	}
	return pluginServiceFromModel(&model)
}

func (r *PluginRegistryRepository) findInstances(ctx context.Context, pluginKey string, manifestHash string) ([]*domainplugin.Instance, error) {
	query := r.db.WithContext(ctx).Where("plugin_key = ?", pluginKey)
	if manifestHash != "" {
		query = query.Where("manifest_hash = ?", manifestHash)
	}
	var models []PluginInstanceModel
	if err := query.Order("instance_id ASC").Find(&models).Error; err != nil {
		return nil, fmt.Errorf("find plugin instances: %w", err)
	}
	instances := make([]*domainplugin.Instance, 0, len(models))
	for i := range models {
		instances = append(instances, pluginInstanceFromModel(&models[i]))
	}
	return instances, nil
}

func (r *PluginRegistryRepository) findRoutes(ctx context.Context, pluginKey string, manifestHash string) ([]*domainplugin.Route, error) {
	var models []PluginRouteModel
	if err := r.db.WithContext(ctx).
		Where("plugin_key = ? AND manifest_hash = ? AND enabled = ?", pluginKey, manifestHash, true).
		Order("gateway_path DESC").
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("find plugin routes: %w", err)
	}
	routes := make([]*domainplugin.Route, 0, len(models))
	for i := range models {
		route, err := pluginRouteFromModel(&models[i])
		if err != nil {
			return nil, err
		}
		routes = append(routes, route)
	}
	return routes, nil
}

func pluginServiceToModel(service *domainplugin.Service) *PluginServiceModel {
	return &PluginServiceModel{
		ID:                  service.ID,
		PluginKey:           service.PluginKey,
		Name:                service.Name,
		Protocol:            string(service.Protocol),
		CurrentManifestHash: service.CurrentManifestHash,
		Status:              string(service.Status),
		MetadataJSON:        marshalStringMap(service.Metadata),
		CreatedAt:           service.CreatedAt,
		UpdatedAt:           service.UpdatedAt,
		DisabledAt:          service.DisabledAt,
	}
}

func pluginServiceFromModel(model *PluginServiceModel) (*domainplugin.Service, error) {
	metadata, err := unmarshalStringMap(model.MetadataJSON)
	if err != nil {
		return nil, fmt.Errorf("decode plugin service metadata: %w", err)
	}
	return &domainplugin.Service{
		ID:                  model.ID,
		PluginKey:           model.PluginKey,
		Name:                model.Name,
		Protocol:            domainplugin.Protocol(model.Protocol),
		CurrentManifestHash: model.CurrentManifestHash,
		Status:              domainplugin.ServiceStatus(model.Status),
		Metadata:            metadata,
		CreatedAt:           model.CreatedAt,
		UpdatedAt:           model.UpdatedAt,
		DisabledAt:          model.DisabledAt,
	}, nil
}

func pluginInstanceToModel(instance *domainplugin.Instance) *PluginInstanceModel {
	return &PluginInstanceModel{
		ID:                  instance.ID,
		PluginKey:           instance.PluginKey,
		InstanceID:          instance.InstanceID,
		Version:             instance.Version,
		BaseURL:             instance.BaseURL,
		HealthPath:          instance.HealthPath,
		ManifestHash:        instance.ManifestHash,
		Status:              string(instance.Status),
		HealthStatus:        string(instance.HealthStatus.Normalize()),
		LastSeenAt:          instance.LastSeenAt,
		LeaseExpiresAt:      instance.LeaseExpiresAt,
		LastCheckedAt:       instance.LastCheckedAt,
		ConsecutiveFailures: instance.ConsecutiveFailures,
		LastError:           instance.LastError,
		LastErrorAt:         instance.LastErrorAt,
		CreatedAt:           instance.CreatedAt,
		UpdatedAt:           instance.UpdatedAt,
	}
}

func pluginInstanceFromModel(model *PluginInstanceModel) *domainplugin.Instance {
	return &domainplugin.Instance{
		ID:                  model.ID,
		PluginKey:           model.PluginKey,
		InstanceID:          model.InstanceID,
		Version:             model.Version,
		BaseURL:             model.BaseURL,
		HealthPath:          model.HealthPath,
		ManifestHash:        model.ManifestHash,
		Status:              domainplugin.InstanceStatus(model.Status),
		HealthStatus:        domainplugin.HealthStatus(model.HealthStatus).Normalize(),
		LastSeenAt:          model.LastSeenAt,
		LeaseExpiresAt:      model.LeaseExpiresAt,
		LastCheckedAt:       model.LastCheckedAt,
		ConsecutiveFailures: model.ConsecutiveFailures,
		LastError:           model.LastError,
		LastErrorAt:         model.LastErrorAt,
		CreatedAt:           model.CreatedAt,
		UpdatedAt:           model.UpdatedAt,
	}
}

func pluginRouteToModel(route *domainplugin.Route) *PluginRouteModel {
	return &PluginRouteModel{
		ID:                route.ID,
		PluginKey:         route.PluginKey,
		ManifestHash:      route.ManifestHash,
		RouteID:           route.RouteID,
		Method:            string(route.Method),
		MatchType:         string(route.MatchType),
		GatewayPath:       route.GatewayPath,
		UpstreamPath:      route.UpstreamPath,
		AuthPolicy:        string(route.AuthPolicy),
		Timeout:           route.Timeout.String(),
		ForwardAuthHeader: route.ForwardAuthHeader,
		Enabled:           route.Enabled,
		MetadataJSON:      marshalStringMap(route.Metadata),
		CreatedAt:         route.CreatedAt,
		UpdatedAt:         route.UpdatedAt,
	}
}

func pluginRouteFromModel(model *PluginRouteModel) (*domainplugin.Route, error) {
	metadata, err := unmarshalStringMap(model.MetadataJSON)
	if err != nil {
		return nil, fmt.Errorf("decode plugin route metadata: %w", err)
	}
	return &domainplugin.Route{
		ID:                model.ID,
		PluginKey:         model.PluginKey,
		ManifestHash:      model.ManifestHash,
		RouteID:           model.RouteID,
		Method:            domainplugin.Method(model.Method),
		MatchType:         domainplugin.MatchType(model.MatchType),
		GatewayPath:       model.GatewayPath,
		UpstreamPath:      model.UpstreamPath,
		AuthPolicy:        domainplugin.AuthPolicy(model.AuthPolicy),
		Timeout:           parsePluginRouteTimeout(model.Timeout),
		ForwardAuthHeader: model.ForwardAuthHeader,
		Enabled:           model.Enabled,
		Metadata:          metadata,
		CreatedAt:         model.CreatedAt,
		UpdatedAt:         model.UpdatedAt,
	}, nil
}

func pluginAuditEventToModel(event *domainplugin.AuditEvent) *PluginAuditEventModel {
	return &PluginAuditEventModel{
		ID:           event.ID,
		PluginKey:    event.PluginKey,
		InstanceID:   event.InstanceID,
		Action:       string(event.Action),
		Message:      truncateString(event.Message, 255),
		MetadataJSON: marshalStringMap(event.Metadata),
		CreatedAt:    event.CreatedAt,
	}
}

func pluginAuditEventFromModel(model *PluginAuditEventModel) (*domainplugin.AuditEvent, error) {
	metadata, err := unmarshalStringMap(model.MetadataJSON)
	if err != nil {
		return nil, fmt.Errorf("decode plugin audit event metadata: %w", err)
	}
	return &domainplugin.AuditEvent{
		ID:         model.ID,
		PluginKey:  model.PluginKey,
		InstanceID: model.InstanceID,
		Action:     domainplugin.AuditAction(model.Action),
		Message:    model.Message,
		Metadata:   metadata,
		CreatedAt:  model.CreatedAt,
	}, nil
}

func marshalStringMap(value map[string]string) string {
	if len(value) == 0 {
		return "{}"
	}
	content, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return string(content)
}

func unmarshalStringMap(raw string) (map[string]string, error) {
	if raw == "" {
		return map[string]string{}, nil
	}
	var value map[string]string
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		return nil, err
	}
	if value == nil {
		value = map[string]string{}
	}
	return value, nil
}

func truncateString(value string, limit int) string {
	if limit <= 0 || len(value) <= limit {
		return value
	}
	return value[:limit]
}

func parsePluginRouteTimeout(raw string) time.Duration {
	timeout, err := time.ParseDuration(raw)
	if err != nil || timeout <= 0 {
		return 5 * time.Second
	}
	return timeout
}
