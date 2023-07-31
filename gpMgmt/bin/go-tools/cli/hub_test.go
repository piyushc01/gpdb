package cli

import (
	"errors"
	"github.com/greenplum-db/gp-common-go-libs/testhelper"
	"testing"
)

func TestCreateConfigFile(t *testing.T) {
	caCertPath := "/tmp/test"
	caKeyPath := "/tmp/test"
	serverCertPath := "/tmp/test"
	serverKeyPath := "/tmp/test"
	hubPort := 80000
	agentPort := 90000
	hostnames := []string{"host1", "host2", "hostn"}
	hubLogDir := "/tmp/test"
	serviceDir := "/tmp/test"
	testhelper.SetupTestLogger()
	t.Run("Check createConfigFile creates config file when no error", func(t *testing.T) {
		origMarshalIndet := MasrshalIndent
		defer func() { MasrshalIndent = origMarshalIndet }()
		MasrshalIndent = func(v any, prefix, indent string) ([]byte, error) {
			return nil, nil
		}
		origwriteConfigContent := writeConfigContent
		defer func() { writeConfigContent = origwriteConfigContent }()
		writeConfigContent = func(filepath string, configContents []byte) error {
			return nil
		}
		origcopyConfigFileToAgents := copyConfigFileToAgents
		defer func() { copyConfigFileToAgents = origcopyConfigFileToAgents }()
		copyConfigFileToAgents = func(hostList []string) error {
			return nil
		}
		err := CreateConfigFile(caCertPath, caKeyPath, serverCertPath, serverKeyPath, hubPort, agentPort, hostnames, hubLogDir, serviceName, serviceDir)
		if err != nil {
			t.Fatalf("Expected no error, got error:%v", err)
		}
	})
	t.Run("CreateConfigFile returns error when json marshalling fails", func(t *testing.T) {
		origMarshalIndet := MasrshalIndent
		defer func() { MasrshalIndent = origMarshalIndet }()
		MasrshalIndent = func(v any, prefix, indent string) ([]byte, error) {
			return nil, errors.New("Test json parsing error")
		}
		err := CreateConfigFile(caCertPath, caKeyPath, serverCertPath, serverKeyPath, hubPort, agentPort, hostnames, hubLogDir, serviceName, serviceDir)
		if err == nil {
			t.Fatalf("Expected json parsing error, got no error")
		}
	})
	t.Run("CreateConfigFile return error when writeConfigContent fails", func(t *testing.T) {
		origMarshalIndet := MasrshalIndent
		defer func() { MasrshalIndent = origMarshalIndet }()
		MasrshalIndent = func(v any, prefix, indent string) ([]byte, error) {
			return nil, nil
		}
		origwriteConfigContent := writeConfigContent
		defer func() { writeConfigContent = origwriteConfigContent }()
		writeConfigContent = func(filepath string, configContents []byte) error {
			return errors.New("WriteConfig error")
		}
		err := CreateConfigFile(caCertPath, caKeyPath, serverCertPath, serverKeyPath, hubPort, agentPort, hostnames, hubLogDir, serviceName, serviceDir)
		if err == nil {
			t.Fatalf("Expected writeConfig error, got no error")
		}

	})
	t.Run("CreateConfigFile copyConfigFile fails to copy", func(t *testing.T) {
		origMarshalIndent := MasrshalIndent
		defer func() { MasrshalIndent = origMarshalIndent }()
		MasrshalIndent = func(v any, prefix, indent string) ([]byte, error) {
			return nil, nil
		}
		origwriteConfigContent := writeConfigContent
		defer func() { writeConfigContent = origwriteConfigContent }()
		writeConfigContent = func(filepath string, configContents []byte) error {
			return nil
		}
		origcopyConfigFileToAgents := copyConfigFileToAgents
		defer func() { copyConfigFileToAgents = origcopyConfigFileToAgents }()
		copyConfigFileToAgents = func(hostList []string) error {
			return errors.New("Copy file error")
		}
		err := CreateConfigFile(caCertPath, caKeyPath, serverCertPath, serverKeyPath, hubPort, agentPort, hostnames, hubLogDir, serviceName, serviceDir)
		if err == nil {
			t.Fatalf("Expected copy file error, got no error")
		}
	})
}
