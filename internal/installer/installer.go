package installer

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"

	"qubert/application"
	"qubert/internal/config"
	"qubert/internal/logger"
)

const (
	githubApiUrl      = "https://api.github.com/repos/dedalqq/qubert/releases/latest"
	binFolder         = "/usr/local/bin"
	systemdUnitFile   = "/etc/systemd/system/qubert.service"
	defaultConfigPath = "/etc/qubert/config.json"
)

type githubApiResult struct {
	Name   string `json:""`
	Assets []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:""`
}

func InstallCurrentVersion(cfg *application.Config, log *logger.Logger) error {
	log.Info("Move bin file")
	binFilePath, err := copyBinFile(os.Args[0])
	if err != nil {
		return err
	}

	err = os.Chmod(binFilePath, 0744)
	if err != nil {
		return err
	}

	// TODO move settings file

	cfg.SettingsFile = "/var/qubert/settings.json"

	log.Info("Save config")
	err = config.Save(defaultConfigPath, cfg)
	if err != nil {
		return err
	}

	// TODO if systemd
	err = installAndStartSystemdUnit(binFilePath, log)
	if err != nil {
		return err
	}

	return nil
}

func installAndStartSystemdUnit(binFilePath string, log *logger.Logger) error {
	log.Info("Install systemd unit file")
	err := installSystemdConfig(binFilePath, defaultConfigPath)
	if err != nil {
		return err
	}

	log.Info("Enable service")
	err = exec.Command("systemctl", "enable", "qubert").Run()
	if err != nil {
		return err
	}

	log.Info("Start service")
	err = exec.Command("systemctl", "restart", "qubert").Run()
	if err != nil {
		return err
	}

	return nil
}

func copyBinFile(file string) (string, error) {
	srcFile, err := os.Open(file)
	if err != nil {
		return "", err
	}

	defer srcFile.Close()

	newBinFilePath := path.Join(binFolder, path.Base(file))

	descFile, err := os.Create(newBinFilePath)
	if err != nil {
		return "", err
	}

	defer descFile.Close()

	_, err = io.Copy(descFile, srcFile)
	if err != nil {
		return "", err
	}

	return newBinFilePath, nil
}

func installLastVersion(currentVersion string, configFilePath string) error {
	//httpClient := &http.Client{
	//	Timeout: 10 * time.Second,
	//}
	//
	//err := downloadAndInstall(httpClient, asset.BrowserDownloadURL, asset.Name)
	//if err != nil {
	//	return err
	//}
	//
	//err = installSystemdConfig(asset.Name, configFilePath)
	//if err != nil {
	//	return err
	//}

	return nil
}

func findLastVersion(client *http.Client, currentVersion string) (string, error) {
	res, err := client.Get(githubApiUrl)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	apiRes := githubApiResult{}

	err = json.NewDecoder(res.Body).Decode(&apiRes)
	if err != nil {
		return "", err
	}

	if currentVersion == apiRes.Name {
		return "", fmt.Errorf("version [%s] already installed", apiRes.Name)
	}

	platform := fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)

	for _, asset := range apiRes.Assets {
		if strings.HasSuffix(asset.Name, platform) {
			return asset.BrowserDownloadURL, nil
		}
	}

	return "", fmt.Errorf("platform [%s] not foud", platform)
}

func downloadAndInstall(client *http.Client, url string, name string) error {
	res, err := client.Get(url)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	f, err := os.Create(path.Join(binFolder, name))
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = io.Copy(f, res.Body)
	return err
}

func installSystemdConfig(binFilePath string, configFilePath string) error {
	unitFile := []string{
		"[Unit]",
		"Description=Qubert",
		"After=network.target",
		"Wants=network-online.target",
		"",
		"[Service]",
		"Restart=always",
		"Type=simple",
		"ExecStart=%s -c %s",
		"",
		"[Install]",
		"WantedBy=multi-user.target",
		"",
	}

	unitFileContent := fmt.Sprintf(strings.Join(unitFile, "\n"), binFilePath, configFilePath)

	err := ioutil.WriteFile(systemdUnitFile, []byte(unitFileContent), 0644)
	if err != nil {
		return err
	}

	return exec.Command("/bin/systemctl", "daemon-reload").Run()
}
