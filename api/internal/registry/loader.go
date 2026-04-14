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

	if err := validate(&reg); err != nil {
		return nil, fmt.Errorf("validate registry file %s: %w", path, err)
	}

	return &reg, nil
}

func validate(reg *ImageRegistry) error {
	if len(reg.Images) == 0 {
		return fmt.Errorf("registry must contain at least one image entry")
	}

	for i, entry := range reg.Images {
		if entry.Match == "" {
			return fmt.Errorf("entry %d: match pattern is required", i)
		}

		for j, cmd := range entry.PostStart {
			if len(cmd.Command) == 0 {
				return fmt.Errorf("entry %d: post_start[%d]: command must not be empty", i, j)
			}
		}

		for j, cmd := range entry.PreStop {
			if len(cmd.Command) == 0 {
				return fmt.Errorf("entry %d: pre_stop[%d]: command must not be empty", i, j)
			}
		}

		seen := make(map[string]bool)
		for j, item := range entry.Metadata {
			if item.Key == "" {
				return fmt.Errorf("entry %d: metadata[%d]: key is required", i, j)
			}
			if item.Type == "" {
				return fmt.Errorf("entry %d: metadata[%d]: type is required", i, j)
			}
			if seen[item.Key] {
				return fmt.Errorf("entry %d: metadata[%d]: duplicate key %q", i, j, item.Key)
			}
			seen[item.Key] = true
		}

		seenLogKeys := make(map[string]bool)
		for j, ls := range entry.Logs {
			if ls.Key == "" {
				return fmt.Errorf("entry %d: logs[%d]: key is required", i, j)
			}
			if ls.Label == "" {
				return fmt.Errorf("entry %d: logs[%d]: label is required", i, j)
			}
			if ls.Type != LogSourceTypeDocker && ls.Type != LogSourceTypeFile && ls.Type != LogSourceTypeLifecycle {
				return fmt.Errorf("entry %d: logs[%d]: type must be %q, %q, or %q", i, j, LogSourceTypeDocker, LogSourceTypeFile, LogSourceTypeLifecycle)
			}
			if ls.Type == LogSourceTypeFile && ls.Path == "" {
				return fmt.Errorf("entry %d: logs[%d]: path is required for file log sources", i, j)
			}
			if (ls.Type == LogSourceTypeDocker || ls.Type == LogSourceTypeLifecycle) && ls.Path != "" {
				return fmt.Errorf("entry %d: logs[%d]: path must not be set for %s log sources", i, j, ls.Type)
			}
			if seenLogKeys[ls.Key] {
				return fmt.Errorf("entry %d: logs[%d]: duplicate key %q", i, j, ls.Key)
			}
			seenLogKeys[ls.Key] = true
		}
	}

	return nil
}
