package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ConfigEngine struct {
	dataDir      string
	templatePath string
}

type ClientConfig struct {
	Log       map[string]interface{} `json:"log"`
	DNS       map[string]interface{} `json:"dns"`
	Inbounds  []map[string]interface{} `json:"inbounds"`
	Outbounds []map[string]interface{} `json:"outbounds"`
	Route     map[string]interface{} `json:"route"`
}

func NewConfigEngine(dataDir, templatePath string) *ConfigEngine {
	return &ConfigEngine{dataDir: dataDir, templatePath: templatePath}
}

func (c *ConfigEngine) GenerateClientConfig(serverIP, uuid, publicKey, shortID, handshakeHost string) error {
	templateBytes, err := os.ReadFile(c.templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template: %w", err)
	}
	cfg := string(templateBytes)
	cfg = strings.ReplaceAll(cfg, "REPLACE_WITH_SERVER_IP", serverIP)
	cfg = strings.ReplaceAll(cfg, "REPLACE_WITH_UUID", uuid)
	cfg = strings.ReplaceAll(cfg, "REPLACE_WITH_REALITY_PUBLIC_KEY", publicKey)
	cfg = strings.ReplaceAll(cfg, "REPLACE_WITH_SHORT_ID", shortID)
	cfg = strings.ReplaceAll(cfg, "REPLACE_WITH_HANDSHAKE_HOST", handshakeHost)

	var parsed interface{}
	if err := json.Unmarshal([]byte(cfg), &parsed); err != nil {
		return fmt.Errorf("invalid template JSON: %w", err)
	}
	path := filepath.Join(c.dataDir, "client.json")
	if err := os.WriteFile(path, []byte(cfg), 0600); err != nil {
		return fmt.Errorf("failed to write client config: %w", err)
	}
	return nil
}

func (c *ConfigEngine) ValidateConfig(configPath string) error {
	cfgBytes, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}
	var cfg interface{}
	if err := json.Unmarshal(cfgBytes, &cfg); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	return nil
}

