package models

type LibraryInstallation struct {
	File       string `json:"file"`
	GroupID    string `json:"group_id"`
	ArtifactID string `json:"artifact_id"`
	Version    string `json:"version"`
	Packaging  string `json:"packaging"`
	Command    string `json:"command"`
}

type GitLabCIConfig struct {
	ServiceID    string                `json:"service_id"`
	Libraries    []LibraryInstallation `json:"libraries"`
	HasLibraries bool                  `json:"has_libraries"`
	ErrorMessage string                `json:"error_message,omitempty"`
}
