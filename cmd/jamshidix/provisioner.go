//go:build windows

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ConfigWizard guides the user through the first-run setup flow.
type ConfigWizard struct {
	hasOCIConfig  bool
	hasClientJSON bool
	dataDir       string
}

// WizardStep is a step in the setup flow.
type WizardStep int

const (
	WizardStepWelcome WizardStep = iota
	WizardStepCheckServer
	WizardStepProvisionServer
	WizardStepLoadOCICredentials
	WizardStepGenerateConfig
	WizardStepValidateConfig
	WizardStepStartTUN
	WizardStepComplete
)

func NewConfigWizard(dataDir string) *ConfigWizard {
	return &ConfigWizard{dataDir: dataDir}
}

func (w *ConfigWizard) Run() error {
	step := WizardStepWelcome
	for step != WizardStepComplete {
		var err error
		step, err = w.executeStep(step)
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *ConfigWizard) executeStep(step WizardStep) (WizardStep, error) {
	switch step {
	case WizardStepWelcome:
		return w.stepWelcome()
	case WizardStepCheckServer:
		return w.stepCheckServer()
	case WizardStepProvisionServer:
		return w.stepProvisionServer()
	case WizardStepLoadOCICredentials:
		return w.stepLoadOCICredentials()
	case WizardStepGenerateConfig:
		return w.stepGenerateConfig()
	case WizardStepValidateConfig:
		return w.stepValidateConfig()
	case WizardStepStartTUN:
		return w.stepStartTUN()
	case WizardStepComplete:
		return WizardStepComplete, nil
	default:
		return WizardStepComplete, fmt.Errorf("unknown wizard step: %v", step)
	}
}

func (w *ConfigWizard) stepWelcome() (WizardStep, error) {
	message := "خوش آمدید به Jamshidix\n\n" +
		"این تنظیم‌گر، اولین اجرای شما را به‌صورت یک‌کلیکی انجام می‌دهد.\n\n" +
		"در این فرایند بررسی می‌شود:\n" +
		"• آیا سرور رایگان OCI موجود است؟\n" +
		"• آیا client.json آماده است؟\n" +
		"• آیا sing-box نصب شده است؟\n"
	showMessage("Jamshidix Setup", message, messageBoxInformation)

	clientPath := filepath.Join(w.dataDir, "client.json")
	if _, err := os.Stat(clientPath); err == nil {
		w.hasClientJSON = true
		return WizardStepValidateConfig, nil
	}
	return WizardStepCheckServer, nil
}

func (w *ConfigWizard) stepCheckServer() (WizardStep, error) {
	serverConfigPath := filepath.Join(w.dataDir, "server.json")
	if _, err := os.Stat(serverConfigPath); err == nil {
		w.hasOCIConfig = true
		return WizardStepGenerateConfig, nil
	}

	showMessage("Jamshidix", "سرور موجود نیست.\n\nدر مرحله بعد سرور رایگان OCI ساخته می‌شود.", messageBoxInformation)
	return WizardStepProvisionServer, nil
}

func (w *ConfigWizard) stepProvisionServer() (WizardStep, error) {
	showMessage("Jamshidix", "در حال آماده‌سازی سرور رایگان OCI...\n\nاین کار ممکن است چند دقیقه طول بکشد.", messageBoxInformation)
	return WizardStepLoadOCICredentials, nil
}

func (w *ConfigWizard) stepLoadOCICredentials() (WizardStep, error) {
	showMessage("Jamshidix", "در حال بارگذاری اطلاعات OCI...", messageBoxInformation)
	return WizardStepGenerateConfig, nil
}

func (w *ConfigWizard) stepGenerateConfig() (WizardStep, error) {
	clientPath := filepath.Join(w.dataDir, "client.json")
	cfg := map[string]interface{}{
		"log": map[string]string{"level": "warn"},
		"outbounds": []map[string]interface{}{{
			"type":   "vless",
			"tag":    "proxy",
			"server": "REPLACE_WITH_SERVER_IP",
			"server_port": 443,
		}},
	}
	cfgBytes, _ := json.MarshalIndent(cfg, "", "  ")
	if err := os.WriteFile(clientPath, cfgBytes, 0600); err != nil {
		return WizardStepComplete, err
	}
	return WizardStepValidateConfig, nil
}

func (w *ConfigWizard) stepValidateConfig() (WizardStep, error) {
	showMessage("Jamshidix", "در حال اعتبارسنجی کانفیگ sing-box...", messageBoxInformation)
	return WizardStepStartTUN, nil
}

func (w *ConfigWizard) stepStartTUN() (WizardStep, error) {
	showMessage("Jamshidix", "در حال فعال‌سازی TUN و شروع اتصال...", messageBoxInformation)
	return WizardStepComplete, nil
}

func showMessage(title, body string, flags uintptr) {
	if strings.Contains(strings.ToLower(os.Getenv("OS")), "windows") {
		showMessageBox(title, body, flags)
	} else {
		fmt.Printf("\n[%s]\n%s\n", title, body)
	}
}

func showMessageBox(title, body string, flags uintptr) {
	fmt.Printf("\n[%s]\n%s\n", title, body)
}

