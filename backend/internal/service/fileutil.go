package service

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"strings"
)

const maxDecompressedSize = 50 * 1024 * 1024 // 50MB

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

		limited := io.LimitReader(gr, maxDecompressedSize+1)
		data, err := io.ReadAll(limited)
		if err != nil {
			return nil, err
		}
		if len(data) > maxDecompressedSize {
			return nil, fmt.Errorf("decompressed file exceeds %d MB limit", maxDecompressedSize/(1024*1024))
		}
		return data, nil
	}
	return os.ReadFile(filePath)
}
