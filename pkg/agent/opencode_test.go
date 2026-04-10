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
	"testing"
)

func TestOpencode_Name(t *testing.T) {
	t.Parallel()

	agent := NewOpencode()
	if got := agent.Name(); got != "opencode" {
		t.Errorf("Name() = %q, want %q", got, "opencode")
	}
}

func TestOpencode_SkipOnboarding_NoExistingSettings(t *testing.T) {
	t.Parallel()

	agent := NewOpencode()
	settings := make(map[string][]byte)

	result, err := agent.SkipOnboarding(settings, "/workspace/sources")
	if err != nil {
		t.Fatalf("SkipOnboarding() error = %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result map")
	}
}

func TestOpencode_SkipOnboarding_NilSettings(t *testing.T) {
	t.Parallel()

	agent := NewOpencode()

	result, err := agent.SkipOnboarding(nil, "/workspace/sources")
	if err != nil {
		t.Fatalf("SkipOnboarding() error = %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result map")
	}
}

func TestOpencode_SkipOnboarding_PreservesExistingSettings(t *testing.T) {
	t.Parallel()

	agent := NewOpencode()

	existingSettings := map[string]interface{}{
		"model": "anthropic/claude-sonnet-4-6",
		"provider": map[string]interface{}{
			"anthropic": map[string]interface{}{
				"name": "Anthropic",
			},
		},
	}

	existingJSON, err := json.Marshal(existingSettings)
	if err != nil {
		t.Fatalf("Failed to marshal existing settings: %v", err)
	}

	settings := map[string][]byte{
		OpencodeConfigPath: existingJSON,
	}

	result, err := agent.SkipOnboarding(settings, "/workspace/sources")
	if err != nil {
		t.Fatalf("SkipOnboarding() error = %v", err)
	}

	// Verify existing settings are preserved unchanged
	resultJSON, exists := result[OpencodeConfigPath]
	if !exists {
		t.Fatal("Expected existing config to be preserved")
	}

	var config map[string]interface{}
	if err := json.Unmarshal(resultJSON, &config); err != nil {
		t.Fatalf("Failed to parse result JSON: %v", err)
	}

	if model, ok := config["model"].(string); !ok || model != "anthropic/claude-sonnet-4-6" {
		t.Errorf("model = %v, want %q", config["model"], "anthropic/claude-sonnet-4-6")
	}
}

func TestOpencode_SetModel_NoExistingSettings(t *testing.T) {
	t.Parallel()

	agent := NewOpencode()
	settings := make(map[string][]byte)

	result, err := agent.SetModel(settings, "anthropic/claude-sonnet-4-6")
	if err != nil {
		t.Fatalf("SetModel() error = %v", err)
	}

	configJSON, exists := result[OpencodeConfigPath]
	if !exists {
		t.Fatalf("Expected %s to be created", OpencodeConfigPath)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(configJSON, &config); err != nil {
		t.Fatalf("Failed to parse result JSON: %v", err)
	}

	if model, ok := config["model"].(string); !ok || model != "anthropic/claude-sonnet-4-6" {
		t.Errorf("model = %v, want %q", config["model"], "anthropic/claude-sonnet-4-6")
	}
}

func TestOpencode_SetModel_NilSettings(t *testing.T) {
	t.Parallel()

	agent := NewOpencode()

	result, err := agent.SetModel(nil, "anthropic/claude-sonnet-4-6")
	if err != nil {
		t.Fatalf("SetModel() error = %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result map")
	}

	if _, exists := result[OpencodeConfigPath]; !exists {
		t.Errorf("Expected %s to be created", OpencodeConfigPath)
	}
}

func TestOpencode_SetModel_PreservesExistingFields(t *testing.T) {
	t.Parallel()

	agent := NewOpencode()

	existingSettings := map[string]interface{}{
		"$schema": "https://opencode.ai/config.json",
		"provider": map[string]interface{}{
			"ollama": map[string]interface{}{
				"name": "Ollama",
			},
		},
	}

	existingJSON, err := json.Marshal(existingSettings)
	if err != nil {
		t.Fatalf("Failed to marshal existing settings: %v", err)
	}

	settings := map[string][]byte{
		OpencodeConfigPath: existingJSON,
	}

	result, err := agent.SetModel(settings, "ollama/gemma4:26b")
	if err != nil {
		t.Fatalf("SetModel() error = %v", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(result[OpencodeConfigPath], &config); err != nil {
		t.Fatalf("Failed to parse result JSON: %v", err)
	}

	// Verify model was set
	if model, ok := config["model"].(string); !ok || model != "ollama/gemma4:26b" {
		t.Errorf("model = %v, want %q", config["model"], "ollama/gemma4:26b")
	}

	// Verify existing fields are preserved
	if schema, ok := config["$schema"].(string); !ok || schema != "https://opencode.ai/config.json" {
		t.Errorf("$schema = %v, want %q", config["$schema"], "https://opencode.ai/config.json")
	}

	provider, ok := config["provider"].(map[string]interface{})
	if !ok {
		t.Fatal("provider is not preserved")
	}

	ollama, ok := provider["ollama"].(map[string]interface{})
	if !ok {
		t.Fatal("provider.ollama is not preserved")
	}

	if name, ok := ollama["name"].(string); !ok || name != "Ollama" {
		t.Errorf("provider.ollama.name = %v, want %q", ollama["name"], "Ollama")
	}
}

func TestOpencode_SetModel_OllamaProviderAutoConfigured(t *testing.T) {
	t.Parallel()

	agent := NewOpencode()
	settings := make(map[string][]byte)

	result, err := agent.SetModel(settings, "ollama/kimi-k2.5:cloud")
	if err != nil {
		t.Fatalf("SetModel() error = %v", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(result[OpencodeConfigPath], &config); err != nil {
		t.Fatalf("Failed to parse result JSON: %v", err)
	}

	// Verify model was set
	if model, ok := config["model"].(string); !ok || model != "ollama/kimi-k2.5:cloud" {
		t.Errorf("model = %v, want %q", config["model"], "ollama/kimi-k2.5:cloud")
	}

	// Verify provider block was auto-generated
	providers, ok := config["provider"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected provider block to be auto-generated")
	}

	ollama, ok := providers["ollama"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected provider.ollama to exist")
	}

	if name, ok := ollama["name"].(string); !ok || name != "Ollama" {
		t.Errorf("provider.ollama.name = %v, want %q", ollama["name"], "Ollama")
	}

	if npm, ok := ollama["npm"].(string); !ok || npm != "@ai-sdk/openai-compatible" {
		t.Errorf("provider.ollama.npm = %v, want %q", ollama["npm"], "@ai-sdk/openai-compatible")
	}

	options, ok := ollama["options"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected provider.ollama.options to exist")
	}

	if baseURL, ok := options["baseURL"].(string); !ok || baseURL != "http://host.containers.internal:11434/v1" {
		t.Errorf("provider.ollama.options.baseURL = %v, want %q", options["baseURL"], "http://host.containers.internal:11434/v1")
	}

	// Verify model is registered in provider's models map
	models, ok := ollama["models"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected provider.ollama.models to exist")
	}

	modelEntry, ok := models["kimi-k2.5:cloud"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected provider.ollama.models[kimi-k2.5:cloud] to exist")
	}

	if name, ok := modelEntry["name"].(string); !ok || name != "kimi-k2.5:cloud" {
		t.Errorf("model entry name = %v, want %q", modelEntry["name"], "kimi-k2.5:cloud")
	}

	if launch, ok := modelEntry["_launch"].(bool); !ok || !launch {
		t.Errorf("model entry _launch = %v, want true", modelEntry["_launch"])
	}
}

func TestOpencode_SetModel_RamalamaProviderAutoConfigured(t *testing.T) {
	t.Parallel()

	agent := NewOpencode()
	settings := make(map[string][]byte)

	result, err := agent.SetModel(settings, "ramalama/granite3.3:8b")
	if err != nil {
		t.Fatalf("SetModel() error = %v", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(result[OpencodeConfigPath], &config); err != nil {
		t.Fatalf("Failed to parse result JSON: %v", err)
	}

	if model, ok := config["model"].(string); !ok || model != "ramalama/granite3.3:8b" {
		t.Errorf("model = %v, want %q", config["model"], "ramalama/granite3.3:8b")
	}

	providers, ok := config["provider"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected provider block to be auto-generated")
	}

	ramalama, ok := providers["ramalama"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected provider.ramalama to exist")
	}

	if name, ok := ramalama["name"].(string); !ok || name != "RamaLama" {
		t.Errorf("provider.ramalama.name = %v, want %q", ramalama["name"], "RamaLama")
	}

	if npm, ok := ramalama["npm"].(string); !ok || npm != "@ai-sdk/openai-compatible" {
		t.Errorf("provider.ramalama.npm = %v, want %q", ramalama["npm"], "@ai-sdk/openai-compatible")
	}

	options, ok := ramalama["options"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected provider.ramalama.options to exist")
	}

	if baseURL, ok := options["baseURL"].(string); !ok || baseURL != "http://host.containers.internal:8080/v1" {
		t.Errorf("provider.ramalama.options.baseURL = %v, want %q", options["baseURL"], "http://host.containers.internal:8080/v1")
	}

	models, ok := ramalama["models"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected provider.ramalama.models to exist")
	}

	modelEntry, ok := models["granite3.3:8b"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected provider.ramalama.models[granite3.3:8b] to exist")
	}

	if name, ok := modelEntry["name"].(string); !ok || name != "granite3.3:8b" {
		t.Errorf("model entry name = %v, want %q", modelEntry["name"], "granite3.3:8b")
	}

	if launch, ok := modelEntry["_launch"].(bool); !ok || !launch {
		t.Errorf("model entry _launch = %v, want true", modelEntry["_launch"])
	}
}

func TestOpencode_SetModel_UnknownProviderNoAutoConfig(t *testing.T) {
	t.Parallel()

	agent := NewOpencode()
	settings := make(map[string][]byte)

	result, err := agent.SetModel(settings, "anthropic/claude-sonnet-4-6")
	if err != nil {
		t.Fatalf("SetModel() error = %v", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(result[OpencodeConfigPath], &config); err != nil {
		t.Fatalf("Failed to parse result JSON: %v", err)
	}

	// Unknown providers should not get auto-configured
	if _, ok := config["provider"]; ok {
		t.Error("Expected no provider block for unknown provider prefix")
	}
}

func TestOpencode_SetModel_OllamaPreservesExistingProviderConfig(t *testing.T) {
	t.Parallel()

	agent := NewOpencode()

	existingSettings := map[string]interface{}{
		"provider": map[string]interface{}{
			"ollama": map[string]interface{}{
				"name": "Ollama",
				"options": map[string]interface{}{
					"baseURL": "http://my-ollama-server:11434/v1",
				},
			},
		},
	}

	existingJSON, err := json.Marshal(existingSettings)
	if err != nil {
		t.Fatalf("Failed to marshal existing settings: %v", err)
	}

	settings := map[string][]byte{
		OpencodeConfigPath: existingJSON,
	}

	result, err := agent.SetModel(settings, "ollama/kimi-k2.5:cloud")
	if err != nil {
		t.Fatalf("SetModel() error = %v", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(result[OpencodeConfigPath], &config); err != nil {
		t.Fatalf("Failed to parse result JSON: %v", err)
	}

	// Verify existing provider config is preserved (custom baseURL)
	providers := config["provider"].(map[string]interface{})
	ollama := providers["ollama"].(map[string]interface{})
	options := ollama["options"].(map[string]interface{})

	if baseURL := options["baseURL"].(string); baseURL != "http://my-ollama-server:11434/v1" {
		t.Errorf("baseURL = %v, want custom URL preserved", baseURL)
	}
}

func TestOpencode_SetModel_InvalidJSON(t *testing.T) {
	t.Parallel()

	agent := NewOpencode()

	settings := map[string][]byte{
		OpencodeConfigPath: []byte("invalid json {{{"),
	}

	_, err := agent.SetModel(settings, "anthropic/claude-sonnet-4-6")
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestOpencode_SetModel_OverwritesExistingModel(t *testing.T) {
	t.Parallel()

	agent := NewOpencode()

	existingSettings := map[string]interface{}{
		"model":   "ollama/gemma4:26b",
		"$schema": "https://opencode.ai/config.json",
	}

	existingJSON, err := json.Marshal(existingSettings)
	if err != nil {
		t.Fatalf("Failed to marshal existing settings: %v", err)
	}

	settings := map[string][]byte{
		OpencodeConfigPath: existingJSON,
	}

	result, err := agent.SetModel(settings, "anthropic/claude-sonnet-4-6")
	if err != nil {
		t.Fatalf("SetModel() error = %v", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(result[OpencodeConfigPath], &config); err != nil {
		t.Fatalf("Failed to parse result JSON: %v", err)
	}

	// Verify model was overwritten
	if model, ok := config["model"].(string); !ok || model != "anthropic/claude-sonnet-4-6" {
		t.Errorf("model = %v, want %q (should overwrite existing)", config["model"], "anthropic/claude-sonnet-4-6")
	}

	// Verify other fields are preserved
	if schema, ok := config["$schema"].(string); !ok || schema != "https://opencode.ai/config.json" {
		t.Errorf("$schema = %v, want %q", config["$schema"], "https://opencode.ai/config.json")
	}
}
