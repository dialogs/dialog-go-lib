package info

// A Info of the service
type Info struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	GoVersion string `json:"goVersion"`
	BuildDate string `json:"buildDate"`
}
