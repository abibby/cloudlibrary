package libgourou

import "os"

func fileExists(p string) bool {
	_, err := os.Lstat(p)
	return err == nil
}
