package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yunuskargi/confbox/internal/crypto"
	"github.com/yunuskargi/confbox/internal/database"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

type gdriveSettings struct {
	Enabled      bool
	ClientID     string
	ClientSecret string
	RefreshToken string
	FolderID     string
}

func getGDriveSettings() gdriveSettings {
	get := func(key, def string) string {
		var val *string
		err := database.DB.Get(&val, "SELECT value FROM settings WHERE key = ?", key)
		if err != nil || val == nil {
			return def
		}
		return *val
	}

	return gdriveSettings{
		Enabled:      get("gdrive_enabled", "false") == "true",
		ClientID:     get("gdrive_client_id", ""),
		ClientSecret: crypto.Decrypt(get("gdrive_client_secret", "")),
		RefreshToken: crypto.Decrypt(get("gdrive_refresh_token", "")),
		FolderID:     get("gdrive_folder_id", ""),
	}
}

func gdriveOAuthConfig(clientID, clientSecret string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
		RedirectURL:  "urn:ietf:wg:oauth:2.0:oob",
		Scopes:       []string{drive.DriveFileScope},
	}
}

func GetGDriveAuthURL() (string, error) {
	g := getGDriveSettings()
	if g.ClientID == "" || g.ClientSecret == "" {
		return "", fmt.Errorf("client ID and secret are required")
	}
	cfg := gdriveOAuthConfig(g.ClientID, g.ClientSecret)
	return cfg.AuthCodeURL("state-token", oauth2.AccessTypeOffline, oauth2.ApprovalForce), nil
}

func ExchangeGDriveCode(code string) error {
	g := getGDriveSettings()
	if g.ClientID == "" || g.ClientSecret == "" {
		return fmt.Errorf("client ID and secret are required")
	}

	cfg := gdriveOAuthConfig(g.ClientID, g.ClientSecret)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	token, err := cfg.Exchange(ctx, code)
	if err != nil {
		return fmt.Errorf("failed to exchange code: %v", err)
	}

	if token.RefreshToken == "" {
		return fmt.Errorf("no refresh token received, try revoking access and authorizing again")
	}

	tokenJSON, _ := json.Marshal(token)
	database.DB.Exec("INSERT OR REPLACE INTO settings (key, value) VALUES (?, ?)", "gdrive_refresh_token", crypto.Encrypt(string(tokenJSON)))

	return nil
}

func newDriveService(g gdriveSettings) (*drive.Service, error) {
	cfg := gdriveOAuthConfig(g.ClientID, g.ClientSecret)

	var token oauth2.Token
	if err := json.Unmarshal([]byte(g.RefreshToken), &token); err != nil {
		return nil, fmt.Errorf("invalid saved token: %v", err)
	}

	ctx := context.Background()
	client := cfg.Client(ctx, &token)

	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("Drive service error: %v", err)
	}
	return srv, nil
}

func ensureSubFolder(srv *drive.Service, parentID, folderName string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	safeName := strings.ReplaceAll(folderName, `\`, `\\`)
	safeName = strings.ReplaceAll(safeName, `'`, `\'`)
	q := fmt.Sprintf("'%s' in parents and name = '%s' and mimeType = 'application/vnd.google-apps.folder' and trashed = false", parentID, safeName)
	list, err := srv.Files.List().Q(q).Fields("files(id)").Context(ctx).Do()
	if err != nil {
		return "", err
	}
	if len(list.Files) > 0 {
		return list.Files[0].Id, nil
	}

	folder, err := srv.Files.Create(&drive.File{
		Name:     folderName,
		MimeType: "application/vnd.google-apps.folder",
		Parents:  []string{parentID},
	}).Fields("id").Context(ctx).Do()
	if err != nil {
		return "", err
	}
	return folder.Id, nil
}

func uploadToGDrive(filePath, vendor, deviceName string) error {
	g := getGDriveSettings()
	if !g.Enabled || g.RefreshToken == "" || g.FolderID == "" {
		return nil
	}

	srv, err := newDriveService(g)
	if err != nil {
		return err
	}

	vendorFolder, err := ensureSubFolder(srv, g.FolderID, vendor)
	if err != nil {
		return fmt.Errorf("failed to create vendor folder: %v", err)
	}

	deviceFolder, err := ensureSubFolder(srv, vendorFolder, deviceName)
	if err != nil {
		return fmt.Errorf("failed to create device folder: %v", err)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	fileName := filepath.Base(filePath)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	_, err = srv.Files.Create(&drive.File{
		Name:    fileName,
		Parents: []string{deviceFolder},
	}).Media(file).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("Google Drive upload failed: %v", err)
	}

	slog.Info("backup uploaded to Google Drive", "folder", deviceFolder, "file", fileName)
	return nil
}

func TestGDriveConnection() error {
	g := getGDriveSettings()
	if g.RefreshToken == "" || g.FolderID == "" {
		return fmt.Errorf("authorize Google Drive first and set folder ID")
	}

	srv, err := newDriveService(g)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = srv.Files.Get(g.FolderID).Fields("id, name").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("cannot access folder: %v", err)
	}

	return nil
}
