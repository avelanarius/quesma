// Copyright Quesma, licensed under the Elastic License 2.0.
// SPDX-License-Identifier: Elastic-2.0
package licensing

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"os"
	"quesma/quesma/config"
	"slices"
)

type LicenseModule struct {
	InstallationID string
	LicenseKey     []byte
	License        *License
	Config         *config.QuesmaConfiguration
}

const (
	installationIdFile = ".installation_id"
)

func Init(config *config.QuesmaConfiguration) *LicenseModule {
	l := &LicenseModule{
		Config:     config,
		LicenseKey: []byte(config.LicenseKey),
	}
	l.Run()
	return l
}

func (l *LicenseModule) Run() {
	if len(l.LicenseKey) > 0 {
		l.logInfo("License key [%s] already present, skipping license key obtainment.", l.LicenseKey)
	} else {
		l.setInstallationID()
		if err := l.obtainLicenseKey(); err != nil {
			PanicWithLicenseViolation(fmt.Errorf("failed to obtain license key: %v", err))
		}
	}
	if err := l.processLicense(); err != nil {
		PanicWithLicenseViolation(fmt.Errorf("failed to process license: %v", err))
	}
	if err := l.validateConfig(); err != nil {
		PanicWithLicenseViolation(fmt.Errorf("failed to validate configuration: %v", err))
	}
}

func (l *LicenseModule) validateConfig() error {
	// Check if connectors are allowed
	for _, conn := range l.Config.Connectors {
		if !slices.Contains(l.License.Connectors, conn.ConnectorType) {
			return fmt.Errorf("connector [%s] is not allowed within the current license", conn.ConnectorType)
		}
	}
	return nil
}

func (l *LicenseModule) setInstallationID() {
	if l.Config.InstallationId != "" {
		l.logInfo("Installation ID provided in the configuration [%s]", l.Config.InstallationId)
		l.InstallationID = l.Config.InstallationId
		return
	}

	if data, err := os.ReadFile(installationIdFile); err != nil {
		l.logDebug("Reading Installation ID failed [%v], generating new one", err)
		generatedID := uuid.New().String()
		l.logDebug("Generated Installation ID of [%s]", generatedID)
		l.tryStoringInstallationIdInFile(generatedID)
		l.InstallationID = generatedID
	} else {
		installationID := string(data)
		l.logDebug("Installation ID found in file [%s]", installationID)
		l.InstallationID = installationID
	}
}

func (l *LicenseModule) tryStoringInstallationIdInFile(installationID string) {
	if err := os.WriteFile(installationIdFile, []byte(installationID), 0644); err != nil {
		l.logDebug("Failed to store Installation ID in file: %v", err)
	} else {
		l.logDebug("Stored Installation ID in file [%s]", installationIdFile)
	}
}

func (l *LicenseModule) logInfo(msg string, args ...interface{}) {
	fmt.Printf(msg+"\n", args...)
}

func (l *LicenseModule) logDebug(msg string, args ...interface{}) {
	if l.Config.Logging.Level == zerolog.DebugLevel {
		fmt.Printf(msg+"\n", args...)
	}
}