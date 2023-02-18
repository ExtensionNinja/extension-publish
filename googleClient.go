package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type TokenRequest struct {
	RefreshToken string `json:"refresh_token"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	GrantType    string `json:"grant_type"`
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
}

type ItemError struct {
	ErrorCode   string `json:"error_code"`
	ErrorDetail string `json:"error_detail"`
}

type ExtensionItemResponse struct {
	Kind        string      `json:"kind"`
	ID          string      `json:"id"`
	PublicKey   string      `json:"publicKey"`
	UploadState string      `json:"uploadState"`
	CrxVersion  string      `json:"crxVersion"`
	ItemError   []ItemError `json:"itemError"`
}

type ExtensionPublishResponse struct {
	Kind         string   `json:"kind"`
	ID           string   `json:"item_id"`
	Status       []string `json:"status"`
	StatusDetail []string `json:"statusDetail"`
}

func refreshAccessToken(ctx context.Context, clientID string, clientSecret string, refreshToken string) (string, error) {
	request := TokenRequest{
		RefreshToken: refreshToken,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		GrantType:    "refresh_token",
	}
	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	// Send access token refresh request
	resp, err := http.Post("https://www.googleapis.com/oauth2/v4/token", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		errContent, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("request failed. %v", string(errContent))
	}

	// Decode response
	respData := TokenResponse{}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&respData)
	if err != nil {
		return "", err
	}

	if !strings.Contains(respData.Scope, "https://www.googleapis.com/auth/chromewebstore") {
		return "", fmt.Errorf("chromewebstore scope missing. Actual: %v", respData.Scope)
	}

	if respData.TokenType != "Bearer" {
		return "", fmt.Errorf("token type must be bearer. Actual: %v", respData.TokenType)
	}

	return respData.AccessToken, nil
}

func getExtensionItem(ctx context.Context, accessToken string, extensionID string) (*ExtensionItemResponse, error) {
	req, err := http.NewRequest("GET", "https://www.googleapis.com/chromewebstore/v1.1/items/"+extensionID+"?projection=DRAFT", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+accessToken)
	req.Header.Add("x-goog-api-version", "2")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		errResp, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed. Err: %v", string(errResp))
	}

	decoder := json.NewDecoder(resp.Body)
	extResponse := ExtensionItemResponse{}
	err = decoder.Decode(&extResponse)
	if err != nil {
		return nil, err
	}

	return &extResponse, nil
}

func uploadExtension(ctx context.Context, accessToken string, extensionID string, extensionFile []byte) (*ExtensionItemResponse, error) {
	req, err := http.NewRequest("PUT", "https://www.googleapis.com/upload/chromewebstore/v1.1/items/"+extensionID+"?uploadType=media", bytes.NewBuffer(extensionFile))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+accessToken)
	req.Header.Add("x-goog-api-version", "2")
	req.Header.Add("Content-Type", "application/zip")
	req.Header.Add("Content-Length", fmt.Sprintf("%d", len(extensionFile)))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		errResp, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed. Err: %v", string(errResp))
	}

	decoder := json.NewDecoder(resp.Body)
	extResponse := ExtensionItemResponse{}
	err = decoder.Decode(&extResponse)
	if err != nil {
		return nil, err
	}
	if extResponse.UploadState == "FAILURE" {
		return nil, fmt.Errorf("upload failed: %v", extResponse.ItemError)
	}

	return &extResponse, nil
}

func publishExtension(ctx context.Context, accessToken string, extensionID string) (*ExtensionPublishResponse, error) {
	resp, err := http.Post("https://www.googleapis.com/chromewebstore/v1.1/items/"+extensionID+"/publish", "application/json", nil)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		errResp, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed. Err: %v", string(errResp))
	}

	decoder := json.NewDecoder(resp.Body)
	publishResponse := ExtensionPublishResponse{}
	err = decoder.Decode(&publishResponse)
	if err != nil {
		return nil, err
	}

	return &publishResponse, nil
}
