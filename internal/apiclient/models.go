package apiclient

// UpdateInfo matches the JSON structure for update information.
type UpdateInfo struct {
	VersionCode int    `json:"versionCode"`
	FileURL     string `json:"fileUrl"`
}

// UpdateErr matches the JSON structure for API error messages.
type UpdateErr struct {
	Message string `json:"message"`
}

// StatusReportPayload matches the JSON structure for reporting status.
type StatusReportPayload struct {
	VersionCode   int    `json:"versionCode"`
	StatusMessage string `json:"statusMessage"`
}
