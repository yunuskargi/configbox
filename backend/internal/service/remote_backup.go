package service

import "log/slog"

type RemoteResult struct {
	S3Enabled      bool
	S3OK           bool
	S3Error        string
	GDriveEnabled  bool
	GDriveOK       bool
	GDriveError    string
}

func UploadToRemoteStorages(filePath, vendor, deviceName string) RemoteResult {
	r := RemoteResult{}

	s3 := getS3Settings()
	if s3.Enabled {
		r.S3Enabled = true
		if err := uploadToS3(filePath, vendor, deviceName); err != nil {
			r.S3Error = err.Error()
			slog.Error("remote backup to S3 failed", "error", err)
		} else {
			r.S3OK = true
		}
	}

	g := getGDriveSettings()
	if g.Enabled {
		r.GDriveEnabled = true
		if err := uploadToGDrive(filePath, vendor, deviceName); err != nil {
			r.GDriveError = err.Error()
			slog.Error("remote backup to Google Drive failed", "error", err)
		} else {
			r.GDriveOK = true
		}
	}

	return r
}
