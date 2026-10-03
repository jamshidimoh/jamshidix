package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// OCIProvisioner handles automatic OCI VM creation.
type OCIProvisioner struct {
	terraformDir string
	dataDir      string
}

type OCIConfig struct {
	TenancyOCID     string `json:"tenancy_ocid"`
	UserOCID        string `json:"user_ocid"`
	Fingerprint     string `json:"fingerprint"`
	PrivateKeyPath  string `json:"private_key_path"`
	Region          string `json:"region"`
	CompartmentOCID string `json:"compartment_ocid"`
}

type VMDetails struct {
	InstanceID           string `json:"instance_id"`
	PublicIP             string `json:"public_ip"`
	PrivateIP            string `json:"private_ip"`
	AvailabilityDomain   string `json:"availability_domain"`
}

func NewOCIProvisioner(terraformDir, dataDir string) *OCIProvisioner {
	return &OCIProvisioner{terraformDir: terraformDir, dataDir: dataDir}
}

func (p *OCIProvisioner) CheckTerraformInstalled() error {
	_, err := exec.LookPath("terraform")
	if err != nil {
		return fmt.Errorf("terraform not found in PATH")
	}
	return nil
}

func (p *OCIProvisioner) LoadOCICredentials() (*OCIConfig, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	ociConfigPath := filepath.Join(homeDir, ".oci", "config")
	if _, err := os.Stat(ociConfigPath); err == nil {
		return &OCIConfig{Region: "us-phoenix-1"}, nil
	}
	return nil, fmt.Errorf("OCI credentials not found. Configure ~/.oci/config or set env vars")
}

func (p *OCIProvisioner) ProvisionVM(cfg *OCIConfig, instanceName string) (*VMDetails, error) {
	if cfg == nil {
		return nil, fmt.Errorf("OCI config is nil")
	}
	if err := os.MkdirAll(p.terraformDir, 0755); err != nil {
		return nil, err
	}
	if err := os.WriteFile(filepath.Join(p.terraformDir, "terraform.tfvars"), []byte(fmt.Sprintf("tenancy_ocid = \"%s\"\nuser_ocid = \"%s\"\nregion = \"%s\"\ncompartment_ocid = \"%s\"\ninstance_name = \"%s\"\n", cfg.TenancyOCID, cfg.UserOCID, cfg.Region, cfg.CompartmentOCID, instanceName)), 0600); err != nil {
		return nil, err
	}
	cmd := exec.Command("terraform", "init")
	cmd.Dir = p.terraformDir
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("terraform init failed: %w: %s", err, string(out))
	}
	cmd = exec.Command("terraform", "apply", "-auto-approve")
	cmd.Dir = p.terraformDir
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("terraform apply failed: %w: %s", err, string(out))
	}
	return p.ExtractVMDetails()
}

func (p *OCIProvisioner) ExtractVMDetails() (*VMDetails, error) {
	cmd := exec.Command("terraform", "output", "-json")
	cmd.Dir = p.terraformDir
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("terraform output failed: %w", err)
	}
	var result map[string]map[string]interface{}
	if err := json.Unmarshal(out, &result); err != nil {
		return nil, err
	}
	details := &VMDetails{}
	if v, ok := result["public_ip"]; ok {
		details.PublicIP = fmt.Sprintf("%v", v["value"])
	}
	if v, ok := result["instance_id"]; ok {
		details.InstanceID = fmt.Sprintf("%v", v["value"])
	}
	bytes, _ := json.MarshalIndent(details, "", "  ")
	if err := os.WriteFile(filepath.Join(p.dataDir, "server.json"), bytes, 0600); err != nil {
		return nil, err
	}
	return details, nil
}

func (p *OCIProvisioner) DownloadServerSetupScripts() error {
	scripts := []struct {
		name string
		url  string
	}{
		{name: "install_ubuntu.sh", url: "https://raw.githubusercontent.com/jamshidimoh/jamshidix/main/server/install_ubuntu.sh"},
		{name: "generate_server_config.sh", url: "https://raw.githubusercontent.com/jamshidimoh/jamshidix/main/server/generate_server_config.sh"},
	}
	for _, s := range scripts {
		resp, err := http.Get(s.url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		path := filepath.Join(p.dataDir, s.name)
		file, err := os.Create(path)
		if err != nil {
			return err
		}
		if _, err := io.Copy(file, resp.Body); err != nil {
			file.Close()
			return err
		}
		file.Close()
		if strings.HasSuffix(s.name, ".sh") {
			os.Chmod(path, 0755)
		}
	}
	return nil
}

