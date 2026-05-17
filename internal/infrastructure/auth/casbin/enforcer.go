package casbin

import (
	"fmt"

	casbinv3 "github.com/casbin/casbin/v3"
	"github.com/rin721/keiyaku-go/internal/infrastructure/config"
)

type Authorizer struct {
	enforcer *casbinv3.Enforcer
}

func NewAuthorizer(cfg config.RBACConfig) (*Authorizer, error) {
	enforcer, err := casbinv3.NewEnforcer(cfg.ModelPath, cfg.PolicyPath)
	if err != nil {
		return nil, fmt.Errorf("build casbin authorizer: %w", err)
	}
	return &Authorizer{enforcer: enforcer}, nil
}

func (a *Authorizer) Allow(role string, object string, action string) (bool, error) {
	if a == nil || a.enforcer == nil {
		return false, fmt.Errorf("casbin authorizer is nil")
	}
	allowed, err := a.enforcer.Enforce(role, object, action)
	if err != nil {
		return false, fmt.Errorf("enforce casbin policy: %w", err)
	}
	return allowed, nil
}
