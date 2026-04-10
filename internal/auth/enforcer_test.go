package auth_test

import (
	"testing"
)

func TestEnforcer_AdminFullAccess(t *testing.T) {
	conn := openTestDB(t)
	svc := newTestService(t, conn)

	cases := []struct{ path, method string }{
		{"/api/v1/workspaces", "GET"},
		{"/api/v1/workspaces", "POST"},
		{"/api/v1/workspaces/123", "DELETE"},
		{"/api/v1/users", "GET"},
	}
	for _, c := range cases {
		ok, err := svc.Enforcer.Allow("admin", c.path, c.method)
		if err != nil {
			t.Errorf("Allow admin %s %s: %v", c.method, c.path, err)
		}
		if !ok {
			t.Errorf("admin should be allowed: %s %s", c.method, c.path)
		}
	}
}

func TestEnforcer_ViewerReadOnly(t *testing.T) {
	conn := openTestDB(t)
	svc := newTestService(t, conn)

	allowed := []struct{ path, method string }{
		{"/api/v1/workspaces", "GET"},
		{"/api/v1/workspaces/abc", "GET"},
	}
	for _, c := range allowed {
		ok, err := svc.Enforcer.Allow("viewer", c.path, c.method)
		if err != nil {
			t.Errorf("Allow viewer %s %s: %v", c.method, c.path, err)
		}
		if !ok {
			t.Errorf("viewer should be allowed: %s %s", c.method, c.path)
		}
	}

	denied := []struct{ path, method string }{
		{"/api/v1/workspaces", "POST"},
		{"/api/v1/workspaces/abc", "DELETE"},
		{"/api/v1/workspaces/abc", "PUT"},
		{"/api/v1/users", "GET"},
	}
	for _, c := range denied {
		ok, err := svc.Enforcer.Allow("viewer", c.path, c.method)
		if err != nil {
			t.Errorf("Allow viewer %s %s: %v", c.method, c.path, err)
		}
		if ok {
			t.Errorf("viewer should be denied: %s %s", c.method, c.path)
		}
	}
}

func TestEnforcer_ArchitectWriteWorkspace(t *testing.T) {
	conn := openTestDB(t)
	svc := newTestService(t, conn)

	allowed := []struct{ path, method string }{
		{"/api/v1/workspaces", "GET"},
		{"/api/v1/workspaces/abc", "GET"},
		{"/api/v1/workspaces/abc", "POST"},
		{"/api/v1/workspaces/abc", "PUT"},
		{"/api/v1/workspaces/abc", "DELETE"},
	}
	for _, c := range allowed {
		ok, err := svc.Enforcer.Allow("architect", c.path, c.method)
		if err != nil {
			t.Errorf("Allow architect %s %s: %v", c.method, c.path, err)
		}
		if !ok {
			t.Errorf("architect should be allowed: %s %s", c.method, c.path)
		}
	}

	// architect cannot manage users (admin only)
	ok, _ := svc.Enforcer.Allow("architect", "/api/v1/users", "GET")
	if ok {
		t.Error("architect should not access /api/v1/users")
	}
}

func TestEnforcer_RoleInheritance(t *testing.T) {
	conn := openTestDB(t)
	svc := newTestService(t, conn)

	// admin inherits architect which inherits viewer — admin can read workspaces
	ok, err := svc.Enforcer.Allow("admin", "/api/v1/workspaces/xyz", "GET")
	if err != nil || !ok {
		t.Errorf("admin should inherit viewer GET permission: ok=%v err=%v", ok, err)
	}
}
