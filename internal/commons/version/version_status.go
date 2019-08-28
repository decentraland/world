package version

var version = "Not available"

type versionResponse struct {
	Version string `json:"version"`
}

func Version() string {
	return version
}
