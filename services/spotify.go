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
	Name   string `json:"name"`
	Images []struct {
		URL string `json:"url"`
	} `json:"images"`
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

func getSpotifyToken() string {
	if sp.AccessToken != "" && time.Since(sp.ObtainedAt) < sp.ExpiresIn {
		return sp.AccessToken
	}

	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequest("POST", "https://accounts.spotify.com/api/token", strings.NewReader(data.Encode()))
	if err != nil {
		log.Fatalf("failed to create auth request: %v", err)
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

func (ss *SpotifyService) makeSpotifyRequest(endpoint string, result interface{}) error {
	token := getSpotifyToken()

	req, _ := http.NewRequest("GET", "https://api.spotify.com/v1/"+endpoint, nil)
	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-ok http status code: %s", resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

func (ss *SpotifyService) FindArtistDetails(spotifyID string) (*SpotifyArtist, error) {
	var artist SpotifyArtist
	if err := ss.makeSpotifyRequest(fmt.Sprintf("artists/%s", spotifyID), &artist); err != nil {
		log.Println(err)
		return nil, err
	}

	return &artist, nil
}

func (ss *SpotifyService) FindArtistLatestRelease(spotifyID string) (*SpotifyRelease, error) {
	albumTypes := []string{"album", "single", "appears_on"}
	var latestRelease SpotifyRelease

	for _, albumType := range albumTypes {
		var releasesReponse struct {
			Items []struct {
				Name        string `json:"name"`
				ReleaseDate string `json:"release_date"`
				SpotifyID   string `json:"id"`
			} `json:"items"`
		}
		endpoint := fmt.Sprintf("artists/%s/albums?include_groups=%s&limit=1", spotifyID, albumType)
		if err := ss.makeSpotifyRequest(endpoint, &releasesReponse); err != nil {
			log.Println(err)
			return nil, err
		}

		if len(releasesReponse.Items) > 0 {
			releaseDate, _ := time.Parse("2006-01-02", releasesReponse.Items[0].ReleaseDate)
			if releaseDate.After(latestRelease.ReleaseDate) {
				latestRelease = SpotifyRelease{
					Name:        releasesReponse.Items[0].Name,
					ReleaseDate: releaseDate,
					SpotifyID:   releasesReponse.Items[0].SpotifyID,
				}
			}
		}
	}

	return &latestRelease, nil
}

func (ss *SpotifyService) FindAlbumDetails(spotifyID string) (map[string]string, error) {
	var albumDetails struct {
		ExternalURL map[string]string `json:"external_urls"`
		Artists     []struct {
			ID string `json:"id"`
		} `json:"artists"`
		Images []struct {
			URL string `json:"url"`
		} `json:"images"`
	}

	if err := ss.makeSpotifyRequest(fmt.Sprintf("albums/%s", spotifyID), &albumDetails); err != nil {
		log.Println(err)
		return nil, err
	}

	result := map[string]string{
		"spotifyURL": albumDetails.ExternalURL["spotify"],
	}

	if len(albumDetails.Images) > 0 {
		result["imageURL"] = albumDetails.Images[0].URL
	}

	if len(albumDetails.Artists) > 0 {
		artistID := albumDetails.Artists[0].ID

		artistDetails, err := ss.FindArtistDetails(artistID)
		if err != nil {
			return nil, err
		}

		if len(artistDetails.Images) > 0 {
			result["artistImageURL"] = artistDetails.Images[0].URL
		}
	}

	return result, nil
}

func (ss *SpotifyService) ExtractSpotifyID(spotifyURL string) string {
	re := regexp.MustCompile(`spotify\.com/artist/([0-9a-zA-Z]+)`)
	matches := re.FindStringSubmatch(spotifyURL)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}
