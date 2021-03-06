package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
)

func main() {
	accessToken := os.Getenv("ESA_ACCESS_TOKEN")
	client := NewClient(accessToken)

	res, err := client.Team.GetTeams()
	if err != nil {
		log.Fatal(err)
		return
	}
	output, _ := json.MarshalIndent(res, "", "  ")
	fmt.Println(string(output))
}

// esa's client
const (
	defaultBaseURL = "https://api.esa.io"
)

type Client struct {
	Client      *http.Client
	accessToken string
	baseURL     string
	Team        *TeamService
}

func NewClient(accessToken string) *Client {
	c := &Client{}
	c.Client = http.DefaultClient
	c.accessToken = accessToken
	c.baseURL = defaultBaseURL
	c.Team = &TeamService{client: c}

	return c
}

func (c *Client) createURL(esaURL string) string {
	return c.baseURL + esaURL + "?access_token=" + c.accessToken
}

func (c *Client) get(esaURL string, query url.Values, v interface{}) (resp *http.Response, err error) {
	path := c.createURL(esaURL)
	queries := query.Encode()
	if len(queries) != 0 {
		path += "&" + queries
	}

	res, err := c.Client.Get(path)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, errors.New(http.StatusText(res.StatusCode))
	}

	if err := responseUnmarshal(res.Body, v); err != nil {
		return nil, err
	}

	return res, err
}

func responseUnmarshal(body io.ReadCloser, v interface{}) error {
	data, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, v); err != nil {
		return err
	}
	return nil
}

// esa's team service
const (
	teamPath = "/v1/teams"
)

type TeamService struct {
	client *Client
}

type TeamResponse struct {
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Name        string `json:"name"`
	Privacy     string `json:"privacy"`
	URL         string `json:"url"`
}

type TeamsResponse struct {
	Teams      []TeamResponse `json:"teams"`
	PrevPage   interface{}    `json:"prev_page"`
	NextPage   interface{}    `json:"next_page"`
	TotalCount int            `json:"total_count"`
}

func (t *TeamService) GetTeams() (*TeamsResponse, error) {
	var teamsRes TeamsResponse
	_, err := t.client.get(teamPath, url.Values{}, &teamsRes)
	if err != nil {
		return nil, err
	}

	return &teamsRes, nil
}
