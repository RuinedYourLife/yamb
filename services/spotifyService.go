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

type SpotifyRelease struct {
	Name        string
	ReleaseDate time.Time
	SpotifyID   string
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
	defer resp.Body.Close()

	if err != nil {
		log.Printf("failed to get artist details: %v", err)
		return nil, err
	}

	var artist *SpotifyArtist
	if err := json.NewDecoder(resp.Body).Decode(&artist); err != nil {
		log.Printf("failed to decode artist details: %v", err)
		return nil, err
	}

	return artist, nil
}

func (ss *SpotifyService) FindLatestRelease(spotifyID string) (*SpotifyRelease, error) {
	token := GetSpotifyToken()

	albumTypes := []string{"album", "single", "appears_on"}
	var latestRelease *SpotifyRelease

	for _, albumType := range albumTypes {
		url := fmt.Sprintf("https://api.spotify.com/v1/artists/%s/albums?include_groups=%s&limit=1", spotifyID, albumType)
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Add("Authorization", "Bearer "+token)

		client := &http.Client{}
		resp, err := client.Do(req)
		defer resp.Body.Close()

		if err != nil {
			log.Printf("failed to get albums: %v", err)
			return nil, err
		}

		var releasesReponse struct {
			Items []struct {
				Name        string `json:"name"`
				ReleaseDate string `json:"release_date"`
				SpotifyID   string `json:"id"`
			} `json:"items"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&releasesReponse); err != nil {
			log.Printf("failed to decode album response: %v", err)
			return nil, err
		}

		if len(releasesReponse.Items) > 0 {
			releaseDate, _ := time.Parse("2006-01-02", releasesReponse.Items[0].ReleaseDate)
			if latestRelease == nil || releaseDate.After(latestRelease.ReleaseDate) {
				latestRelease = &SpotifyRelease{
					Name:        releasesReponse.Items[0].Name,
					ReleaseDate: releaseDate,
					SpotifyID:   releasesReponse.Items[0].SpotifyID,
				}
			}
		}
	}

	return latestRelease, nil
}
