package casbin

import (
	"github.com/casbin/casbin/v2"
	"log"
)

func InitCasbin() *casbin.Enforcer {
	e, err := casbin.NewEnforcer("pkg/casbin/rbac_model.conf", "pkg/casbin/rbac_policy.csv")
	if err != nil {
		log.Fatalf("Failed to create Casbin enforcer: %v", err)
	}

	err = e.LoadPolicy() // Load policies from file
	if err != nil {
		return nil
	}
	return e
}
