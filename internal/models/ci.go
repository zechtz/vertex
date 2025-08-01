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

// New models for library preview functionality
type LibraryPreview struct {
	HasLibraries   bool                    `json:"hasLibraries"`
	ServiceName    string                  `json:"serviceName"`
	ServiceID      string                  `json:"serviceId"`
	Environments   []EnvironmentLibraries  `json:"environments"`
	TotalLibraries int                     `json:"totalLibraries"`
	GitlabCIExists bool                    `json:"gitlabCIExists"`
	ErrorMessage   string                  `json:"errorMessage,omitempty"`
}

type EnvironmentLibraries struct {
	Environment string                  `json:"environment"`
	JobName     string                  `json:"jobName"`
	Libraries   []LibraryInstallation   `json:"libraries"`
	Branches    []string                `json:"branches"`
}

type LibraryInstallRequest struct {
	Environments []string `json:"environments"`
	Confirmed    bool     `json:"confirmed"`
}

type InstallProgress struct {
	ServiceID     string                    `json:"serviceId"`
	Status        string                    `json:"status"` // "started", "in_progress", "completed", "failed"
	Environments  []EnvironmentProgress     `json:"environments"`
	OverallProgress float64                 `json:"overallProgress"`
	StartTime     string                    `json:"startTime"`
	EndTime       string                    `json:"endTime,omitempty"`
	ErrorMessage  string                    `json:"errorMessage,omitempty"`
}

type EnvironmentProgress struct {
	Name            string   `json:"name"`
	Status          string   `json:"status"` // "pending", "installing", "completed", "failed"
	Total           int      `json:"total"`
	Completed       int      `json:"completed"`
	CurrentLibrary  string   `json:"currentLibrary,omitempty"`
	Errors          []string `json:"errors,omitempty"`
}
