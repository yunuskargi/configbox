package auth

import (
	"encoding/base64"
	"fmt"

	"github.com/pquerna/otp/totp"
	"github.com/skip2/go-qrcode"
)

func GenerateTOTP(username string) (secret string, qrBase64 string, err error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "ConfigBox",
		AccountName: username,
	})
	if err != nil {
		return "", "", err
	}

	png, err := qrcode.Encode(key.URL(), qrcode.Medium, 256)
	if err != nil {
		return "", "", err
	}

	b64 := fmt.Sprintf("data:image/png;base64,%s", base64.StdEncoding.EncodeToString(png))
	return key.Secret(), b64, nil
}

func ValidateTOTP(code, secret string) bool {
	return totp.Validate(code, secret)
}
