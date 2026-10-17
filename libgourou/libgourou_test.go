package libgourou

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestActivate(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test utility uses a POSIX shell script")
	}
	for _, mode := range []string{"success", "command failure"} {
		t.Run(mode, func(t *testing.T) {
			root := t.TempDir()
			t.Chdir(root)
			t.Setenv("ADEPT_DIR", "")
			t.Setenv("PATH", root+string(os.PathListSeparator)+os.Getenv("PATH"))
			t.Setenv("LIBGOUROU_TEST_FAILURE", mode)
			writeTestFile(t, filepath.Join(root, "adept_activate"), `#!/bin/sh
set -eu
test "$#" -eq 3
test "$1" = '--anonymous'
test "$2" = '--output-dir'
printf 'called' > called
if test "$LIBGOUROU_TEST_FAILURE" = 'command failure'; then
    printf 'activation rejected\n' >&2
    exit 9
fi
mkdir "$3"
for name in activation.xml device.xml devicesalt; do
    printf 'activation data' > "$3/$name"
done
`, 0755)
			output := filepath.Join(root, "custom activation")
			err := Activate(output)
			switch mode {
			case "command failure":
				var exitErr *exec.ExitError
				if !errors.As(err, &exitErr) || exitErr.ExitCode() != 9 || !strings.Contains(err.Error(), "activation rejected") {
					t.Fatalf("missing activation failure diagnostics: %v", err)
				}
			default:
				if err != nil {
					t.Fatal(err)
				}
				for _, name := range []string{"activation.xml", "device.xml", "devicesalt"} {
					if _, err := os.Stat(filepath.Join(output, name)); err != nil {
						t.Fatal(err)
					}
				}
			}
		})
	}
	t.Run("missing executable", func(t *testing.T) {
		t.Chdir(t.TempDir())
		t.Setenv("PATH", t.TempDir())
		err := Activate(".adept")
		if !errors.Is(err, exec.ErrNotFound) || !strings.Contains(err.Error(), "find adept_activate") {
			t.Fatalf("expected missing executable error, got %v", err)
		}
	})
}

func TestActivateExistingDirectory(t *testing.T) {
	t.Chdir(t.TempDir())
	// An existing activation must be rejected before invoking the utility.
	t.Setenv("PATH", t.TempDir())
	if err := os.Mkdir(".adept", 0700); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, filepath.Join(".adept", "activation.xml"), "existing activation", 0600)
	if err := Activate(".adept"); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("expected existing directory error, got %v", err)
	}
	data, err := os.ReadFile(filepath.Join(".adept", "activation.xml"))
	if err != nil || string(data) != "existing activation" {
		t.Fatalf("activation changed: %q, %v", data, err)
	}
}

func TestActivateExistingFile(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("PATH", t.TempDir())
	writeTestFile(t, ".adept", "existing file", 0600)
	if err := Activate(".adept"); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("expected existing file error, got %v", err)
	}
}

func TestDownloadEPUB(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test utilities use POSIX shell scripts")
	}
	for _, tt := range []struct {
		name, failure, wantErr, wantCalls string
	}{
		{name: "success", wantCalls: "download\ndecrypt\n"},
		{name: "download failure", failure: "download", wantErr: "download EPUB", wantCalls: "download\n"},
		{name: "decrypt failure", failure: "decrypt", wantErr: "decrypt EPUB", wantCalls: "download\ndecrypt\n"},
		{name: "missing output", failure: "missing", wantErr: "read decrypted EPUB", wantCalls: "download\ndecrypt\n"},
		{name: "empty output", failure: "empty", wantErr: "decrypted EPUB is empty", wantCalls: "download\ndecrypt\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			binDir := filepath.Join(dir, "bin")
			if err := os.Mkdir(binDir, 0755); err != nil {
				t.Fatal(err)
			}
			writeTestFile(t, filepath.Join(binDir, "acsmdownloader"), `#!/bin/sh
set -eu
printf 'download\n' >> "$LIBGOUROU_TEST_CALLS"
test "$1" = '-f'
test "$2" = "$LIBGOUROU_TEST_INPUT"
test "$3" = '-o'
test "$#" -eq 6
test "$5" = '--adept-directory'
test "$6" = "$LIBGOUROU_TEST_ACTIVATION"
printf 'encrypted' > "$4"
if test "$LIBGOUROU_TEST_FAILURE" = download; then
    printf 'fulfillment failed\n' >&2
    exit 7
fi
`, 0755)
			writeTestFile(t, filepath.Join(binDir, "adept_remove"), `#!/bin/sh
set -eu
printf 'decrypt\n' >> "$LIBGOUROU_TEST_CALLS"
test "$1" = '-f'
test "$3" = '-o'
test "$#" -eq 6
test "$5" = '--adept-directory'
test "$6" = "$LIBGOUROU_TEST_ACTIVATION"
test "$(cat "$2")" = encrypted
case "$LIBGOUROU_TEST_FAILURE" in
    missing) exit 0 ;;
    empty) : > "$4"; exit 0 ;;
    decrypt) printf 'partial' > "$4"; printf 'decryption failed\n' >&2; exit 8 ;;
esac
printf 'decrypted EPUB' > "$4"
`, 0755)
			input := filepath.Join(dir, "book with spaces;$.acsm")
			outputDir := filepath.Join(dir, "books with spaces")
			output := filepath.Join(outputDir, "book.epub")
			calls := filepath.Join(dir, "calls")
			writeTestFile(t, input, "ACSM", 0600)
			if tt.wantErr != "" {
				if err := os.Mkdir(outputDir, 0755); err != nil {
					t.Fatal(err)
				}
				writeTestFile(t, output, "existing EPUB", 0600)
			}
			t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
			adeptDir := filepath.Join(dir, "custom activation;$")
			t.Setenv("ADEPT_DIR", filepath.Join(dir, "other activation"))
			t.Setenv("LIBGOUROU_TEST_ACTIVATION", adeptDir)
			t.Setenv("LIBGOUROU_TEST_INPUT", input)
			t.Setenv("LIBGOUROU_TEST_CALLS", calls)
			t.Setenv("LIBGOUROU_TEST_FAILURE", tt.failure)

			err := DownloadEPUB(adeptDir, input, output)
			wantOutput := "decrypted EPUB"
			if tt.wantErr == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else {
				wantOutput = "existing EPUB"
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %v, want %q", err, tt.wantErr)
				}
				if tt.failure == "download" || tt.failure == "decrypt" {
					var exitErr *exec.ExitError
					if !errors.As(err, &exitErr) || !strings.Contains(err.Error(), "failed") {
						t.Fatalf("error must retain exit status and diagnostics: %v", err)
					}
				}
			}
			for path, want := range map[string]string{output: wantOutput, input: "ACSM", calls: tt.wantCalls} {
				data, err := os.ReadFile(path)
				if err != nil || string(data) != want {
					t.Fatalf("%s = %q, %v; want %q", path, data, err, want)
				}
			}
			entries, err := os.ReadDir(outputDir)
			if err != nil || len(entries) != 1 || entries[0].Name() != "book.epub" {
				t.Fatalf("unexpected leftover files: %v, %v", entries, err)
			}
		})
	}
}

func writeTestFile(t *testing.T, path, contents string, mode os.FileMode) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), mode); err != nil {
		t.Fatal(err)
	}
}
