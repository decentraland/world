package profile

type ProfileV1 struct {
	SchemaVersion int    `json:schemaVersion`
	BodyShape     string `json:bodyShape`
	Nose          string `json:nose`
	Torso         string `json:torso`
}
