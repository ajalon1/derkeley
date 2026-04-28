// Copyright 2026 DataRobot, Inc. and its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// PersistableKeys is the allowlist of viper keys that may be written back to
// drconfig.yaml. Any key not in this set is intentionally NOT persisted, even
// if it is currently set in the live viper config (e.g. transient flags such
// as --yes, --verbose, --force-interactive).
//
// To add a new persistable key, add it here. Keys may use viper dotted-path
// notation (e.g. "pulumi.config.passphrase") for nested values, but the
// current callers all use flat top-level keys.
var PersistableKeys = map[string]struct{}{
	DataRobotURL:               {},
	DataRobotAPIKey:            {},
	APIConsumerTrackingEnabled: {},
	"ssl_verify":               {},
	"pulumi_config_passphrase": {},
}

// UpdateConfigFile writes only the allowlisted keys from viper back to the
// drconfig.yaml file on disk, preserving any other fields that already exist
// in the file but are not currently tracked by viper.
//
// This replaces direct calls to viper.WriteConfig(), which would otherwise
// serialize the entire viper.AllSettings() map -- including transient command
// flags such as --yes that should never be persisted.
//
// The keys argument optionally restricts the write to a subset of the
// allowlist. If keys is empty, all allowlisted keys currently set in viper
// are written. Any key passed in that is not in the allowlist is ignored.
func UpdateConfigFile(keys ...string) error {
	if err := CreateConfigFileDirIfNotExists(); err != nil {
		return err
	}

	configFile, err := resolveConfigFilePath()
	if err != nil {
		return err
	}

	existing, err := readYAMLFile(configFile)
	if err != nil {
		return err
	}

	applyAllowedKeys(existing, keys)

	out, err := yaml.Marshal(existing)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configFile, out, 0o600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// resolveConfigFilePath returns the active drconfig.yaml path, falling back
// to the default location if viper has not yet recorded a config file used.
func resolveConfigFilePath() (string, error) {
	if configFile := viper.ConfigFileUsed(); configFile != "" {
		return configFile, nil
	}

	dir, err := GetConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, configFileName), nil
}

// applyAllowedKeys overlays values from viper onto target for each key in
// the allowlist. If keys is empty, all keys in PersistableKeys are
// considered. Keys not in PersistableKeys are silently ignored.
func applyAllowedKeys(target map[string]interface{}, keys []string) {
	candidates := keys
	if len(candidates) == 0 {
		candidates = make([]string, 0, len(PersistableKeys))

		for k := range PersistableKeys {
			candidates = append(candidates, k)
		}
	}

	for _, key := range candidates {
		if _, ok := PersistableKeys[key]; !ok {
			continue
		}

		if !viper.IsSet(key) {
			continue
		}

		setNestedKey(target, key, viper.Get(key))
	}
}

// readYAMLFile reads a YAML file into a generic map. If the file does not
// exist or is empty, an empty map is returned.
func readYAMLFile(path string) (map[string]interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return map[string]interface{}{}, nil
		}

		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if len(data) == 0 {
		return map[string]interface{}{}, nil
	}

	out := map[string]interface{}{}
	if err := yaml.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if out == nil {
		out = map[string]interface{}{}
	}

	return out, nil
}

// setNestedKey sets a value at a dotted-path key in a nested map, creating
// intermediate maps as needed. Keys without dots are set at the top level.
func setNestedKey(m map[string]interface{}, key string, value interface{}) {
	parts := strings.Split(key, ".")

	cur := m

	for i, p := range parts {
		if i == len(parts)-1 {
			cur[p] = value
			return
		}

		next, ok := cur[p].(map[string]interface{})
		if !ok {
			next = map[string]interface{}{}
			cur[p] = next
		}

		cur = next
	}
}
