package casbin

import (
	"fmt"

	casbinv3 "github.com/casbin/casbin/v3"
	"github.com/casbin/casbin/v3/model"
)

const rbacModel = `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && keyMatch2(r.obj, p.obj) && regexMatch(r.act, p.act)
`

func NewEnforcer() (*casbinv3.Enforcer, error) {
	m, err := model.NewModelFromString(rbacModel)
	if err != nil {
		return nil, fmt.Errorf("build casbin model: %w", err)
	}
	enforcer, err := casbinv3.NewEnforcer(m)
	if err != nil {
		return nil, fmt.Errorf("build casbin enforcer: %w", err)
	}
	if err := seedPolicies(enforcer); err != nil {
		return nil, err
	}
	return enforcer, nil
}

func seedPolicies(enforcer *casbinv3.Enforcer) error {
	if enforcer == nil {
		return fmt.Errorf("casbin enforcer is nil")
	}
	policies := [][]string{
		{"admin", "/api/v1/*", "(GET|POST|PUT|PATCH|DELETE)"},
		{"author", "/api/v1/users/me", "GET"},
		{"author", "/api/v1/articles", "POST"},
	}
	for _, policy := range policies {
		if _, err := enforcer.AddPolicy(policy[0], policy[1], policy[2]); err != nil {
			return fmt.Errorf("add casbin policy: %w", err)
		}
	}
	roles := [][]string{
		{"admin", "author"},
	}
	for _, role := range roles {
		if _, err := enforcer.AddGroupingPolicy(role[0], role[1]); err != nil {
			return fmt.Errorf("add casbin role: %w", err)
		}
	}
	return nil
}
