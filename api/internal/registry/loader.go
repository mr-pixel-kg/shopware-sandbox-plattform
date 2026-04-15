package registry

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func Load(path string) (*ImageRegistry, error) {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("read registry file %s: %w", path, err)
	}

	var reg ImageRegistry
	if err := yaml.Unmarshal(data, &reg); err != nil {
		return nil, fmt.Errorf("parse registry file %s: %w", path, err)
	}

	if err := validateRegistry(&reg); err != nil {
		return nil, fmt.Errorf("validate registry file %s: %w", path, err)
	}

	return &reg, nil
}
