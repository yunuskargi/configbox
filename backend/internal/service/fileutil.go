package service

import (
	"compress/gzip"
	"io"
	"os"
	"strings"
)

func ReadBackupFile(filePath string) ([]byte, error) {
	if strings.HasSuffix(filePath, ".gz") {
		f, err := os.Open(filePath)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		gr, err := gzip.NewReader(f)
		if err != nil {
			return nil, err
		}
		defer gr.Close()

		return io.ReadAll(gr)
	}
	return os.ReadFile(filePath)
}
