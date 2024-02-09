package services

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

type SpotifyService struct{}

type SpotifyArtist struct {
	Name      string `json:"name"`
	SpotifyID string `json:"id"`
}

type SpotifyAuth struct {
	AccessToken string
	ExpiresIn   time.Duration
	ObtainedAt  time.Time
}

var sp SpotifyAuth

func NewSpotifyService() *SpotifyService {
	return &SpotifyService{}
}

func GetSpotifyToken() string {
	if sp.AccessToken != "" && time.Since(sp.ObtainedAt) < sp.ExpiresIn {
		return sp.AccessToken
	}

	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequest("POST", "https://accounts.spotify.com/api/token", strings.NewReader(data.Encode()))
	if err != nil {
		log.Fatalf("error creating auth request: %v", err)
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(os.Getenv("SPOTIFY_CLIENT"), os.Getenv("SPOTIFY_SECRET"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("failed to request token: %v", err)
	}
	defer resp.Body.Close()

	var tokenReponse struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenReponse); err != nil {
		log.Fatalf("failed to decode token response: %v", err)
	}

	sp.AccessToken = tokenReponse.AccessToken
	sp.ExpiresIn = time.Duration(tokenReponse.ExpiresIn) * time.Second
	sp.ObtainedAt = time.Now()

	return sp.AccessToken
}

func (ss *SpotifyService) ExtractSpotifyID(spotifyURL string) string {
	re := regexp.MustCompile(`spotify\.com/artist/([0-9a-zA-Z]+)`)
	matches := re.FindStringSubmatch(spotifyURL)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func (ss *SpotifyService) FindArtistDetails(spotifyID string) (*SpotifyArtist, error) {
	token := GetSpotifyToken()

	url := fmt.Sprintf("https://api.spotify.com/v1/artists/%s", spotifyID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var artist SpotifyArtist
	if err := json.NewDecoder(resp.Body).Decode(&artist); err != nil {
		return nil, err
	}

	return &artist, nil
}
