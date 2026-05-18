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
	ID             int64     `gorm:"column:id;primaryKey;autoIncrement"`
	PluginKey      string    `gorm:"column:plugin_key"`
	InstanceID     string    `gorm:"column:instance_id"`
	Version        string    `gorm:"column:version"`
	BaseURL        string    `gorm:"column:base_url"`
	HealthPath     string    `gorm:"column:health_path"`
	ManifestHash   string    `gorm:"column:manifest_hash"`
	Status         string    `gorm:"column:status"`
	LastSeenAt     time.Time `gorm:"column:last_seen_at"`
	LeaseExpiresAt time.Time `gorm:"column:lease_expires_at"`
	LastError      string    `gorm:"column:last_error"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at"`
}

func (PluginInstanceModel) TableName() string {
	return "plugin_instances"
}

type PluginRouteModel struct {
	ID                int64     `gorm:"column:id;primaryKey;autoIncrement"`
	PluginKey         string    `gorm:"column:plugin_key"`
	ManifestHash      string    `gorm:"column:manifest_hash"`
	Method            string    `gorm:"column:method"`
	MatchType         string    `gorm:"column:match_type"`
	Path              string    `gorm:"column:path"`
	UpstreamPath      string    `gorm:"column:upstream_path"`
	AuthPolicy        string    `gorm:"column:auth_policy"`
	TimeoutMS         int       `gorm:"column:timeout_ms"`
	ForwardAuthHeader bool      `gorm:"column:forward_auth_header"`
	Enabled           bool      `gorm:"column:enabled"`
	MetadataJSON      string    `gorm:"column:metadata_json"`
	CreatedAt         time.Time `gorm:"column:created_at"`
	UpdatedAt         time.Time `gorm:"column:updated_at"`
}

func (PluginRouteModel) TableName() string {
	return "plugin_routes"
}

type PluginRegistryRepository struct {
	db *gorm.DB
}

func NewPluginRegistryRepository(db *gorm.DB) *PluginRegistryRepository {
	return &PluginRegistryRepository{db: db}
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

func (r *PluginRegistryRepository) FindRoutable(ctx context.Context, pluginKey string, now time.Time) (*domainplugin.Service, []*domainplugin.Instance, []*domainplugin.Route, error) {
	if r == nil || r.db == nil {
		return nil, nil, nil, fmt.Errorf("plugin registry repository is not ready")
	}
	service, err := r.findService(ctx, pluginKey)
	if err != nil {
		return nil, nil, nil, err
	}
	if service.Status != domainplugin.ServiceStatusActive {
		return service, nil, nil, nil
	}
	var instanceModels []PluginInstanceModel
	if err := r.db.WithContext(ctx).
		Where("plugin_key = ? AND manifest_hash = ? AND status = ? AND lease_expires_at >= ?", pluginKey, service.CurrentManifestHash, string(domainplugin.InstanceStatusActive), now).
		Order("instance_id ASC").
		Find(&instanceModels).Error; err != nil {
		return nil, nil, nil, fmt.Errorf("find routable plugin instances: %w", err)
	}
	instances := make([]*domainplugin.Instance, 0, len(instanceModels))
	for i := range instanceModels {
		instances = append(instances, pluginInstanceFromModel(&instanceModels[i]))
	}
	routes, err := r.findRoutes(ctx, pluginKey, service.CurrentManifestHash)
	if err != nil {
		return nil, nil, nil, err
	}
	return service, instances, routes, nil
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
		Order("path DESC").
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
		ID:             instance.ID,
		PluginKey:      instance.PluginKey,
		InstanceID:     instance.InstanceID,
		Version:        instance.Version,
		BaseURL:        instance.BaseURL,
		HealthPath:     instance.HealthPath,
		ManifestHash:   instance.ManifestHash,
		Status:         string(instance.Status),
		LastSeenAt:     instance.LastSeenAt,
		LeaseExpiresAt: instance.LeaseExpiresAt,
		LastError:      instance.LastError,
		CreatedAt:      instance.CreatedAt,
		UpdatedAt:      instance.UpdatedAt,
	}
}

func pluginInstanceFromModel(model *PluginInstanceModel) *domainplugin.Instance {
	return &domainplugin.Instance{
		ID:             model.ID,
		PluginKey:      model.PluginKey,
		InstanceID:     model.InstanceID,
		Version:        model.Version,
		BaseURL:        model.BaseURL,
		HealthPath:     model.HealthPath,
		ManifestHash:   model.ManifestHash,
		Status:         domainplugin.InstanceStatus(model.Status),
		LastSeenAt:     model.LastSeenAt,
		LeaseExpiresAt: model.LeaseExpiresAt,
		LastError:      model.LastError,
		CreatedAt:      model.CreatedAt,
		UpdatedAt:      model.UpdatedAt,
	}
}

func pluginRouteToModel(route *domainplugin.Route) *PluginRouteModel {
	return &PluginRouteModel{
		ID:                route.ID,
		PluginKey:         route.PluginKey,
		ManifestHash:      route.ManifestHash,
		Method:            string(route.Method),
		MatchType:         string(route.MatchType),
		Path:              route.Path,
		UpstreamPath:      route.UpstreamPath,
		AuthPolicy:        string(route.AuthPolicy),
		TimeoutMS:         int(route.Timeout / time.Millisecond),
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
		Method:            domainplugin.Method(model.Method),
		MatchType:         domainplugin.MatchType(model.MatchType),
		Path:              model.Path,
		UpstreamPath:      model.UpstreamPath,
		AuthPolicy:        domainplugin.AuthPolicy(model.AuthPolicy),
		Timeout:           time.Duration(model.TimeoutMS) * time.Millisecond,
		ForwardAuthHeader: model.ForwardAuthHeader,
		Enabled:           model.Enabled,
		Metadata:          metadata,
		CreatedAt:         model.CreatedAt,
		UpdatedAt:         model.UpdatedAt,
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
