package config

import (
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {

	t.Run("successful config load", func(t *testing.T) {
		configYAML := `
server_port: 8080
data_store_type: "mysql"
kubernetes:
  namespace: "default"
  kubeconfig: "/path/to/kubeconfig"
crd:
  files: ["crd1.yaml", "crd2.yaml"]
  load: true
logging:
  level: "info"
  filename: "/var/log/app.log"
data_store: "k8s"
backend_type: "omc_rest_v1"
omc:
  url: "http://example.com"
  username: "admin"
  password: "password"
`
		tempFile, err := ioutil.TempFile("", "config*.yaml")
		if err != nil {
			t.Fatal("Failed to create temp file:", err)
		}
		defer os.Remove(tempFile.Name()) // Cleanup

		_, err = tempFile.WriteString(configYAML)
		if err != nil {
			t.Fatal("Failed to write temp file:", err)
		}
		tempFile.Close()

		config, err := LoadConfig(tempFile.Name())
		if err != nil {
			t.Fatalf("LoadConfig() failed: %v", err)
		}

		if config.ServerPort != 8080 {
			t.Errorf("Expected ServerPort = 8080, got %d", config.ServerPort)
		}
	})

	t.Run("file read error", func(t *testing.T) {
		_, err := LoadConfig("/invalid/path/to/config.yaml")
		if err == nil {
			t.Fatal("Expected error when reading non-existent file")
		}
	})

	t.Run("YAML unmarshal error", func(t *testing.T) {
		tempFile, _ := ioutil.TempFile("", "invalid*.yaml")
		defer os.Remove(tempFile.Name())

		_, err := tempFile.WriteString("invalid_yaml: : :")
		if err != nil {
			t.Fatal("Failed to write temp file:", err)
		}
		tempFile.Close()

		_, err = LoadConfig(tempFile.Name())
		if err == nil {
			t.Fatal("Expected error due to invalid YAML")
		}
	})

	t.Run("validation failure", func(t *testing.T) {
		config := Config{}

		err := config.Validate()
		if err == nil {
			t.Fatal("Expected validation error due to missing server_port")
		}
	})
}

func TestValidate(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		config := Config{ServerPort: 8080}
		err := config.Validate()
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
	})

	t.Run("missing server port", func(t *testing.T) {
		config := Config{}
		err := config.Validate()
		if err == nil || err.Error() != "server_port is required" {
			t.Fatalf("Expected 'server_port is required' error, got %v", err)
		}
	})
}

type mockFS struct {
	files map[string]bool
}

func (m *mockFS) statFile(path string) error {
	if m.files[path] {
		return nil
	}
	return os.ErrNotExist
}

func TestFindConfigFile(t *testing.T) {

	t.Run("config file found in /etc", func(t *testing.T) {
		defer resetOsStat() // Reset after test

		osStatFunc = func(name string) (os.FileInfo, error) {
			if name == "/etc/config.yaml" {
				return mockFileInfo{}, nil
			}
			return nil, os.ErrNotExist
		}

		got := findConfigFile()
		want := "/etc/config.yaml"
		if got != want {
			t.Errorf("findConfigFile() = %v, want %v", got, want)
		}
	})

	t.Run("config file not found", func(t *testing.T) {
		defer resetOsStat()

		osStatFunc = func(name string) (os.FileInfo, error) {
			return nil, os.ErrNotExist
		}

		got := findConfigFile()
		if got != "" {
			t.Errorf("Expected empty result, got %v", got)
		}
	})
}

type mockFileInfo struct{}

func (fi mockFileInfo) Name() string       { return "mockfile" }
func (fi mockFileInfo) Size() int64        { return 0 }
func (fi mockFileInfo) Mode() os.FileMode  { return 0644 }
func (fi mockFileInfo) ModTime() time.Time { return time.Now() }
func (fi mockFileInfo) IsDir() bool        { return false }
func (fi mockFileInfo) Sys() interface{}   { return nil }
