package main

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/go-github/v53/github"
	"golang.org/x/oauth2"
	"log"
	"os"
	"time"
)

const (
	appIdEnvKey           = "APP_ID"
	appPrivateKeyEnvKey   = "APP_PRIVATE_KEY"
	githubOutputEnvKey    = "GITHUB_OUTPUT"
	userNameEnvKey        = "USER_NAME"
	privateRepoNameEnvKey = "PRIVATE_REPO_NAME"
)

func main() {
	privateKey := getEnv(appPrivateKeyEnvKey)
	appId := getEnv(appIdEnvKey)
	githubOutput := getEnv(githubOutputEnvKey)
	userName := getEnv(userNameEnvKey)
	privateRepo := getEnv(privateRepoNameEnvKey)

	if privateKey == "" || appId == "" {
		fmt.Println("[ERROR] Environment variable is empty")
	}

	// Generate JWT
	token, err := generateJWT(appId, privateKey)
	if err != nil {
		log.Fatalf("[ERROR] MESSAGE: Failed to generate JWT. ERROR: %v", err)
	}

	// Create GitHub client with JWT
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// Get Installation ID
	installation, _, err := client.Apps.FindRepositoryInstallation(ctx, userName, privateRepo)
	if err != nil {
		log.Fatalf("[ERROR] Failed to get installation ID: %v", err)
	}
	installationID := installation.GetID()

	// Get Access Token
	accessToken, _, err := client.Apps.CreateInstallationToken(ctx, installationID, nil)
	if err != nil {
		log.Fatalf("[ERROR] Failed to get access token: %v", err)
	}
	accessTokenValue := accessToken.GetToken()

	// Write to GITHUB_OUTPUT
	file, err := os.OpenFile(githubOutput, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("[ERROR] Failed to open file: ", err)
		return
	}
	defer file.Close()
	fmt.Fprintf(file, "accessToken=%s\n", accessTokenValue)
}

func getEnv(key string) string {
	env := os.Getenv(key)
	if env == "" {
		log.Fatalf("[ERROR] Can not read environment variable: key=%s", key)
	}
	return env
}

func generateJWT(appId string, privateKey string) (string, error) {
	token := jwt.New(jwt.GetSigningMethod("RS256"))
	token.Claims = jwt.MapClaims{
		"iss": appId,
		"iat": time.Now().Unix() - 60,
		"exp": time.Now().Unix() + (3 * 60),
	}

	key, err := parsePrivateKey(privateKey)
	if err != nil {
		fmt.Println("[ERROR] Failed to parse private key: ", err)
	}

	tokenString, err := token.SignedString(key)
	if err != nil {
		fmt.Println("[ERROR] Failed to generate JWT: ", err)
	}

	return tokenString, err
}

func parsePrivateKey(privateKey string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(privateKey))
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, fmt.Errorf("[ERROR] Failed to decode PEM block containing private key")
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("[ERROR] Failed to parse private key: %v", err)
	}

	return key, nil
}
