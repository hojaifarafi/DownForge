package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// GenerateSystemdUnit generates a systemd service file string for Linux
func GenerateSystemdUnit(execPath, user string) string {
	return fmt.Sprintf(`[Unit]
Description=DownForge Download Manager Service
After=network.target network-online.target
Wants=network-online.target

[Service]
Type=simple
User=%s
ExecStart=%s -port 8080
Restart=always
RestartSec=5s
LimitNOFILE=65536

# Environment
Environment=DOWNFORGE_PORT=8080

[Install]
WantedBy=multi-user.target
`, user, execPath)
}

// InstallLinuxService installs the current binary as a systemd service
func InstallLinuxService() error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("systemd service install is only supported on Linux")
	}

	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to determine executable path: %w", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("failed to resolve executable path: %w", err)
	}

	currentUser := os.Getenv("USER")
	if currentUser == "" {
		currentUser = "root"
	}

	serviceContent := GenerateSystemdUnit(execPath, currentUser)
	servicePath := "/etc/systemd/system/downforge.service"

	err = os.WriteFile("/tmp/downforge.service", []byte(serviceContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write temp service file: %w", err)
	}

	cmd := exec.Command("sudo", "cp", "/tmp/downforge.service", servicePath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to copy service to %s (requires sudo): %w", servicePath, err)
	}

	_ = exec.Command("sudo", "systemctl", "daemon-reload").Run()
	_ = exec.Command("sudo", "systemctl", "enable", "--now", "downforge.service").Run()

	return nil
}

// UninstallLinuxService disables and removes the systemd service
func UninstallLinuxService() error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("systemd uninstall is only supported on Linux")
	}

	_ = exec.Command("sudo", "systemctl", "stop", "downforge.service").Run()
	_ = exec.Command("sudo", "systemctl", "disable", "downforge.service").Run()
	_ = exec.Command("sudo", "rm", "-f", "/etc/systemd/system/downforge.service").Run()
	_ = exec.Command("sudo", "systemctl", "daemon-reload").Run()

	return nil
}
