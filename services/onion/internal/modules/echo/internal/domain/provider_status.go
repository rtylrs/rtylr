package domain

type ProviderStatus struct {
	Available       bool
	PrimaryProvider string
	Providers       map[string]bool
}
