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

type SpotifyResourceDetails struct {
	Name           string
	URL            string
	ImageURL       string
	ReleaseDate    string
	ArtistName     string
	ArtistImageURL string
	OwnerName      string
	OwnerImageURL  string
	Public         string
}

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

func (ss *SpotifyService) FetchArtistDetails(spotifyID string) (*SpotifyArtist, error) {
	var artist SpotifyArtist
	if err := ss.makeSpotifyRequest(fmt.Sprintf("artists/%s", spotifyID), &artist); err != nil {
		log.Println(err)
		return nil, err
	}

	return &artist, nil
}

func (ss *SpotifyService) FetchArtistLatestRelease(spotifyID string) (*SpotifyRelease, error) {
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

func (ss *SpotifyService) FetchAlbumDetails(spotifyID string) (*SpotifyResourceDetails, error) {
	var albumDetails struct {
		Name        string            `json:"name"`
		ExternalURL map[string]string `json:"external_urls"`
		Artists     []struct {
			ID string `json:"id"`
		} `json:"artists"`
		Images []struct {
			URL string `json:"url"`
		} `json:"images"`
		ReleaseDate string `json:"release_date"`
	}

	if err := ss.makeSpotifyRequest(fmt.Sprintf("albums/%s", spotifyID), &albumDetails); err != nil {
		log.Println(err)
		return nil, err
	}

	details := &SpotifyResourceDetails{
		URL:         albumDetails.ExternalURL["spotify"],
		Name:        albumDetails.Name,
		ReleaseDate: albumDetails.ReleaseDate,
	}

	if len(albumDetails.Images) > 0 {
		details.ImageURL = albumDetails.Images[0].URL
	}

	if len(albumDetails.Artists) > 0 {
		artistID := albumDetails.Artists[0].ID
		artistDetails, err := ss.FetchArtistDetails(artistID)
		if err != nil {
			return nil, err
		}
		details.ArtistName = artistDetails.Name
		if len(artistDetails.Images) > 0 {
			details.ArtistImageURL = artistDetails.Images[0].URL
		}
	}

	return details, nil
}

func (ss *SpotifyService) FetchTrackDetails(spotifyID string) (*SpotifyResourceDetails, error) {
	var trackDetails struct {
		ExternalURL map[string]string `json:"external_urls"`
		Artists     []struct {
			ID string `json:"id"`
		} `json:"artists"`
		Album struct {
			Images []struct {
				URL string `json:"url"`
			} `json:"images"`
			ReleaseDate string `json:"release_date"`
		} `json:"album"`
		Name string `json:"name"`
	}

	if err := ss.makeSpotifyRequest(fmt.Sprintf("tracks/%s", spotifyID), &trackDetails); err != nil {
		log.Println(err)
		return nil, err
	}

	details := &SpotifyResourceDetails{
		URL:         trackDetails.ExternalURL["spotify"],
		Name:        trackDetails.Name,
		ReleaseDate: trackDetails.Album.ReleaseDate,
	}

	if len(trackDetails.Album.Images) > 0 {
		details.ImageURL = trackDetails.Album.Images[0].URL
	}

	if len(trackDetails.Artists) > 0 {
		artistID := trackDetails.Artists[0].ID

		artistDetails, err := ss.FetchArtistDetails(artistID)
		if err != nil {
			return nil, err
		}

		details.ArtistName = artistDetails.Name

		if len(artistDetails.Images) > 0 {
			details.ArtistImageURL = artistDetails.Images[0].URL
		}
	}

	return details, nil
}

func (ss *SpotifyService) FetchPlaylistDetails(spotifyID string) (*SpotifyResourceDetails, error) {
	var playlistDetails struct {
		ExternalURL map[string]string `json:"external_urls"`
		Images      []struct {
			URL string `json:"url"`
		} `json:"images"`
		Name  string `json:"name"`
		Owner struct {
			DisplayName string            `json:"display_name"`
			ExternalURL map[string]string `json:"external_urls"`
			ID          string            `json:"id"`
		} `json:"owner"`
		Public bool `json:"public"`
	}

	if err := ss.makeSpotifyRequest(fmt.Sprintf("playlists/%s", spotifyID), &playlistDetails); err != nil {
		log.Println(err)
		return nil, err
	}

	details := &SpotifyResourceDetails{
		URL:       playlistDetails.ExternalURL["spotify"],
		Name:      playlistDetails.Name,
		OwnerName: playlistDetails.Owner.DisplayName,
		Public:    "No",
	}

	if playlistDetails.Public {
		details.Public = "Yes"
	}

	if len(playlistDetails.Images) > 0 {
		details.ImageURL = playlistDetails.Images[0].URL
	}

	var ownerDetails struct {
		Images []struct {
			URL string `json:"url"`
		} `json:"images"`
	}

	if err := ss.makeSpotifyRequest(fmt.Sprintf("users/%s", playlistDetails.Owner.ID), &ownerDetails); err != nil {
		log.Println(err)
		return nil, err
	}

	if len(ownerDetails.Images) > 0 {
		details.OwnerImageURL = ownerDetails.Images[0].URL
	}

	return details, nil
}

func (ss *SpotifyService) ExtractSpotifyInfos(spotifyURL string) (string, string) {
	re := regexp.MustCompile(`spotify\.com/(\w+)/([0-9a-zA-Z]+)`)
	matches := re.FindStringSubmatch(spotifyURL)
	if len(matches) > 2 {
		return matches[1], matches[2]
	}
	return "", ""
}
