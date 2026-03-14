package importer

import "gopkg.in/yaml.v3"

// yamlUnmarshal is a thin wrapper around yaml.v3 Unmarshal.
func yamlUnmarshal(data []byte, out interface{}) error {
	return yaml.Unmarshal(data, out)
}
