package generate

type Metadata struct {
	Properties map[string]string `json:"properties"`
}

func NewMetadata() *Metadata {
	return &Metadata{
		Properties: make(map[string]string),
	}
}

func (m *Metadata) Set(key string, value string) {
	m.Properties[key] = value
}

func (m *Metadata) Get(key string) string {
	return m.Properties[key]
}
