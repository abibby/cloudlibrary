package libgourou

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Activate registers an anonymous device using adept_activate from PATH.
// It writes activation.xml, device.xml, and devicesalt to adeptDir, which must
// not already exist.
// Reuse these files for subsequent downloads instead of activating each time.
func Activate(adeptDir string) error {
	if _, err := os.Lstat(adeptDir); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("libgourou: check activation directory: %w", err)
	}
	activator, err := exec.LookPath("adept_activate")
	if err != nil {
		return fmt.Errorf("libgourou: find adept_activate: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(adeptDir), 0700); err != nil {
		return fmt.Errorf("libgourou: create activation parent directory: %w", err)
	}
	if err := run(activator, "--anonymous", "--output-dir", adeptDir); err != nil {
		return fmt.Errorf("libgourou: activate Adobe device: %w", err)
	}
	return nil
}

// DownloadEPUB fulfills an ACSM file and saves the decrypted EPUB to epubPath.
// acsmdownloader and adept_remove must be on PATH. Both tools use the existing
// Adobe activation in adeptDir.
// The destination directory is created if needed. An existing destination is
// replaced only after both commands succeed.
func DownloadEPUB(adeptDir, acsmPath, epubPath string) error {
	if fileExists(epubPath) {
		return nil
	}
	if acsmPath == "" || epubPath == "" {
		return fmt.Errorf("libgourou: ACSM and EPUB paths are required")
	}
	acsmPath, err := filepath.Abs(acsmPath)
	if err != nil {
		return fmt.Errorf("libgourou: resolve ACSM path: %w", err)
	}
	info, err := os.Stat(acsmPath)
	if err != nil {
		return fmt.Errorf("libgourou: read ACSM file: %w", err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("libgourou: ACSM path is not a regular file")
	}

	// Check both utilities before consuming the fulfillment token.
	downloader, err := exec.LookPath("acsmdownloader")
	if err != nil {
		return fmt.Errorf("libgourou: find acsmdownloader: %w", err)
	}
	decryptor, err := exec.LookPath("adept_remove")
	if err != nil {
		return fmt.Errorf("libgourou: find adept_remove: %w", err)
	}

	dir := filepath.Dir(epubPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("libgourou: create EPUB directory: %w", err)
	}
	// Stage on the destination filesystem so the final rename is atomic.
	workDir, err := os.MkdirTemp(dir, ".libgourou-*")
	if err != nil {
		return fmt.Errorf("libgourou: create temporary directory: %w", err)
	}
	defer os.RemoveAll(workDir)

	encryptedPath := filepath.Join(workDir, "encrypted.epub")
	decryptedPath := filepath.Join(workDir, "decrypted.epub")
	if err := run(downloader, "-f", acsmPath, "-o", encryptedPath, "--adept-directory", adeptDir); err != nil {
		return fmt.Errorf("libgourou: download EPUB: %w", err)
	}
	if err := run(decryptor, "-f", encryptedPath, "-o", decryptedPath, "--adept-directory", adeptDir); err != nil {
		return fmt.Errorf("libgourou: decrypt EPUB: %w", err)
	}
	info, err = os.Stat(decryptedPath)
	if err != nil {
		return fmt.Errorf("libgourou: read decrypted EPUB: %w", err)
	}
	if !info.Mode().IsRegular() || info.Size() == 0 {
		return fmt.Errorf("libgourou: decrypted EPUB is empty or not a regular file")
	}
	if err := os.Rename(decryptedPath, epubPath); err != nil {
		return fmt.Errorf("libgourou: publish EPUB: %w", err)
	}
	return nil
}

func run(command string, args ...string) error {
	output, err := exec.Command(command, args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %w: %s", filepath.Base(command), err, strings.TrimSpace(string(output)))
	}
	return nil
}
