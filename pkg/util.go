package troll

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"sort"
)

func hashFiles(files []string) (string, error) {
	sort.Strings(files)
	hash := sha256.New()
	for _, fileName := range files {
		f, err := os.Open(fileName)
		if err != nil {
			return "", err
		}
		defer f.Close()
		if _, err := io.Copy(hash, f); err != nil {
			return "", err
		}
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
