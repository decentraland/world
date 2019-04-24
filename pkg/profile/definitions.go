package profile

type ProfileV1 struct {
	SchemaVersion int    `json:"schemaVersion"`
	BodyShape     string `json:"bodyShape,omitempty"`
	Nose          string `json:"nose,omitempty"`
	Torso         string `json:"torso,omitempty"`
}
