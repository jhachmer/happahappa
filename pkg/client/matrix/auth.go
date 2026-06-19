package matrix

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/jhachmer/happahappa/pkg/config"
	"golang.org/x/crypto/ssh/terminal"
)

type MatrixLoginRequest struct {
	Identifier Identifier `json:"identifier"`
	Password   string     `json:"password"`
	Type       string     `json:"type"`

	BaseURL string
}

type MatrixLoginResponse struct {
	UserId      string `json:"user_id"`
	AccessToken string `json:"access_token"`
	HomeServer  string `json:"home_server"`
	DeviceId    string `json:"device_id"`
	WellKnown   struct {
		MHomeserver struct {
			BaseUrl string `json:"base_url"`
		} `json:"m.homeserver"`
	} `json:"well_known"`
}

type Identifier struct {
	Type string `json:"type"`
	User string `json:"user"`
}

func NewMatrixLogin(cfg *config.Config) *MatrixLoginRequest {
	password := readCredentials()
	return &MatrixLoginRequest{
		Identifier: Identifier{
			Type: "m.id.user",
			User: cfg.Matrix.UserID,
		},
		Password: password,
		Type:     "m.login.password",

		BaseURL: cfg.Matrix.BaseUrl,
	}
}

func (ml *MatrixLoginRequest) GetMatrixLogin() (*MatrixLoginResponse, error) {
	const loginEndpoint = "/_matrix/client/v3/login"

	reqBody, err := json.Marshal(ml)
	if err != nil {
		return nil, err
	}
	loginURL := ml.BaseURL + loginEndpoint
	resp, err := http.Post(loginURL, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("could not get access token: %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var loginResponse MatrixLoginResponse
	err = json.Unmarshal(body, &loginResponse)
	if err != nil {
		return nil, err
	}
	return &loginResponse, nil
}

func readCredentials() string {
	fmt.Println("Enter the password for the Matrix Bot:")
	bytePassword, err := terminal.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		slog.Error("Error reading password from stdin")
		return ""
	}
	return string(bytePassword)
}

type Credentials struct {
	UserID      string
	AccessToken string
}

func NewCredentials(cfg *config.Config) (*Credentials, error) {
	return &Credentials{
		UserID:      cfg.Matrix.UserID,
		AccessToken: cfg.Matrix.AccessToken,
	}, nil
}

func GenerateAccessToken(cfg *config.Config) (*MatrixLoginResponse, error) {
	loginRequest := NewMatrixLogin(cfg)
	loginResponse, err := loginRequest.GetMatrixLogin()
	if err != nil {
		return nil, err
	}
	return loginResponse, nil
}
