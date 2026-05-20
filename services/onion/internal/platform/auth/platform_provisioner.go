package auth

import "context"

// PlatformProvisioner creates a platform-specific user account and returns
// claims to be embedded in the JWT token.
type PlatformProvisioner interface {
	// ProvisionUser creates the platform user if it doesn't exist and returns platform claims.
	ProvisionUser(ctx context.Context, userID int64) (*PlatformClaims, error)

	// ResolvePlatformClaims returns platform claims for an existing user (used during login).
	ResolvePlatformClaims(ctx context.Context, userID int64) (*PlatformClaims, error)
}

// PlatformClaims are embedded in the JWT alongside the base user_id.
type PlatformClaims struct {
	Platform       string
	PlatformUserID int64
	PlanID         int64
}

// PlatformProvisionerResolver selects the right provisioner based on product platform.
type PlatformProvisionerResolver interface {
	ResolveProvisioner(platform ProductPlatform) PlatformProvisioner
}

// StaticPlatformProvisionerResolver maps product platforms to provisioners.
type StaticPlatformProvisionerResolver struct {
	provisioners map[ProductPlatform]PlatformProvisioner
}

// NewStaticPlatformProvisionerResolver creates a resolver with a static mapping.
func NewStaticPlatformProvisionerResolver(provisioners map[ProductPlatform]PlatformProvisioner) *StaticPlatformProvisionerResolver {
	if provisioners == nil {
		provisioners = make(map[ProductPlatform]PlatformProvisioner)
	}

	return &StaticPlatformProvisionerResolver{provisioners: provisioners}
}

// ResolveProvisioner returns the provisioner for the given platform, or nil if none registered.
func (r *StaticPlatformProvisionerResolver) ResolveProvisioner(platform ProductPlatform) PlatformProvisioner {
	if r == nil {
		return nil
	}

	return r.provisioners[platform]
}
