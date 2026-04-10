package auth

import "fmt"

// Bootstrap ensures the first admin user exists when the DB is empty,
// and upserts the demo viewer account when demo mode is enabled.
func Bootstrap(svc *Service) error {
	if err := bootstrapAdmin(svc); err != nil {
		return err
	}
	return bootstrapDemo(svc)
}

func bootstrapAdmin(svc *Service) error {
	if svc.Cfg.BootstrapEmail == "" || svc.Cfg.BootstrapPassword == "" {
		return nil
	}

	exists, err := svc.Users.Exists()
	if err != nil {
		return fmt.Errorf("bootstrap check: %w", err)
	}
	if exists {
		return nil
	}

	hash, err := HashPassword(svc.Cfg.BootstrapPassword)
	if err != nil {
		return fmt.Errorf("bootstrap hash password: %w", err)
	}

	_, err = svc.Users.Create(svc.Cfg.BootstrapEmail, hash, "admin")
	if err != nil {
		return fmt.Errorf("bootstrap create admin: %w", err)
	}

	fmt.Printf("auth: bootstrapped admin user %q\n", svc.Cfg.BootstrapEmail)
	return nil
}

// bootstrapDemo creates or resets the demo viewer account on every startup
// when DEMO_MODE=true, so the password is always in sync with DEMO_PASSWORD.
func bootstrapDemo(svc *Service) error {
	if !svc.Cfg.DemoMode {
		return nil
	}

	hash, err := HashPassword(svc.Cfg.DemoPassword)
	if err != nil {
		return fmt.Errorf("demo bootstrap hash: %w", err)
	}

	existing, err := svc.Users.GetByEmail(svc.Cfg.DemoEmail)
	if err == ErrNotFound {
		_, err = svc.Users.Create(svc.Cfg.DemoEmail, hash, "viewer")
		if err != nil {
			return fmt.Errorf("demo bootstrap create: %w", err)
		}
		fmt.Printf("auth: created demo user %q\n", svc.Cfg.DemoEmail)
		return nil
	}
	if err != nil {
		return fmt.Errorf("demo bootstrap lookup: %w", err)
	}

	// Update hash in case DEMO_PASSWORD changed.
	if err := svc.Users.UpdatePasswordHash(existing.ID.String(), hash); err != nil {
		return fmt.Errorf("demo bootstrap update: %w", err)
	}
	return nil
}
