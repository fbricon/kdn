/**********************************************************************
 * Copyright (C) 2026 Red Hat, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * SPDX-License-Identifier: Apache-2.0
 **********************************************************************/

package agent

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	// OpencodeConfigPath is the relative path to the OpenCode configuration file.
	OpencodeConfigPath = ".config/opencode/opencode.json"
)

// opencodeAgent is the implementation of Agent for OpenCode.
type opencodeAgent struct{}

// Compile-time check to ensure opencodeAgent implements Agent interface
var _ Agent = (*opencodeAgent)(nil)

// NewOpencode creates a new OpenCode agent implementation.
func NewOpencode() Agent {
	return &opencodeAgent{}
}

// Name returns the agent name.
func (o *opencodeAgent) Name() string {
	return "opencode"
}

// SkillsDir returns empty string as OpenCode does not currently support skills mounting.
func (o *opencodeAgent) SkillsDir() string {
	return ""
}

// SkipOnboarding returns settings unchanged as OpenCode does not have
// an onboarding flow that needs to be skipped.
func (o *opencodeAgent) SkipOnboarding(settings map[string][]byte, _ string) (map[string][]byte, error) {
	if settings == nil {
		settings = make(map[string][]byte)
	}

	return settings, nil
}

// SetModel configures the model ID in OpenCode settings.
// It sets the model field in .config/opencode/opencode.json.
// All other fields in the settings file are preserved.
func (o *opencodeAgent) SetModel(settings map[string][]byte, modelID string) (map[string][]byte, error) {
	if settings == nil {
		settings = make(map[string][]byte)
	}

	var existingContent []byte
	var exists bool
	if existingContent, exists = settings[OpencodeConfigPath]; !exists {
		existingContent = []byte("{}")
	}

	var config map[string]interface{}
	if err := json.Unmarshal(existingContent, &config); err != nil {
		return nil, fmt.Errorf("failed to parse existing %s: %w", OpencodeConfigPath, err)
	}

	config["model"] = modelID

	// Auto-configure provider block based on model prefix (e.g. "ollama/gemma4:26b")
	if provider, modelName, ok := strings.Cut(modelID, "/"); ok {
		ensureProviderConfig(config, provider, modelName)
	}

	modifiedContent, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal modified %s: %w", OpencodeConfigPath, err)
	}

	settings[OpencodeConfigPath] = modifiedContent
	return settings, nil
}

// ollamaProviderConfig holds the default configuration for the Ollama provider.
var ollamaProviderConfig = map[string]interface{}{
	"name": "Ollama",
	"npm":  "@ai-sdk/openai-compatible",
	"options": map[string]interface{}{
		"baseURL": "http://host.containers.internal:11434/v1",
	},
}

// ramalamaProviderConfig holds the default configuration for the RamaLama provider.
var ramalamaProviderConfig = map[string]interface{}{
	"name": "RamaLama",
	"npm":  "@ai-sdk/openai-compatible",
	"options": map[string]interface{}{
		"baseURL": "http://host.containers.internal:8080/v1",
	},
}

// providerDefaults maps known provider prefixes to their default configuration.
var providerDefaults = map[string]map[string]interface{}{
	"ollama":   ollamaProviderConfig,
	"ramalama": ramalamaProviderConfig,
}

// ensureProviderConfig adds a provider configuration block if the provider prefix
// is recognized and not already configured. It also registers the model under
// the provider's models map.
func ensureProviderConfig(config map[string]interface{}, provider, modelName string) {
	defaults, known := providerDefaults[provider]
	if !known {
		return
	}

	// Get or create the top-level "provider" map
	providers, _ := config["provider"].(map[string]interface{})
	if providers == nil {
		providers = make(map[string]interface{})
	}
	config["provider"] = providers

	// Get or create the specific provider entry
	providerEntry, _ := providers[provider].(map[string]interface{})
	if providerEntry == nil {
		providerEntry = make(map[string]interface{})
		// Apply defaults for new provider entries
		for k, v := range defaults {
			providerEntry[k] = v
		}
	}
	providers[provider] = providerEntry

	// Register the model in the provider's models map
	models, _ := providerEntry["models"].(map[string]interface{})
	if models == nil {
		models = make(map[string]interface{})
	}
	if _, exists := models[modelName]; !exists {
		models[modelName] = map[string]interface{}{
			"_launch": true,
			"name":    modelName,
		}
	}
	providerEntry["models"] = models
}
