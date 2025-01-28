package buildkit

type BuildKitSecretStore struct {
	secrets map[string]string
}

func NewBuildKitSecretStore() *BuildKitSecretStore {
	return &BuildKitSecretStore{
		secrets: make(map[string]string),
	}
}

func (s *BuildKitSecretStore) GetAllSecrets() map[string][]byte {
	secrets := make(map[string][]byte)
	for k, v := range s.secrets {
		secrets[k] = []byte(v)
	}
	return secrets
}

func (s *BuildKitSecretStore) SetSecret(key, value string) {
	s.secrets[key] = value
}

func (s *BuildKitSecretStore) GetSecret(key string) (string, bool) {
	value, ok := s.secrets[key]
	return value, ok
}
