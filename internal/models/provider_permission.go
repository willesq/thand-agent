package models

type ProviderPermissionsResponse struct {
	Version     string               `json:"version"`
	Provider    string               `json:"provider"`
	Permissions []ProviderPermission `json:"permissions"`
}

type ProviderPermission struct {
	Name        string `json:"name"`
	Title       string `json:"title"`
	Description string `json:"description"`

	// Store the underlying provider-specific permission object if needed
	Permission any `json:"-"`
}
