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
	"strconv"
	"strings"
)

func main() {
	accessToken := os.Getenv("ESA_ACCESS_TOKEN")
	client := NewClient(accessToken)

	// team api request
	teamRes, teamErr := client.Team.GetTeams()
	if teamErr != nil {
		log.Fatal(teamErr)
		return
	}
	teamOutput, _ := json.MarshalIndent(teamRes, "", "  ")
	fmt.Println(string(teamOutput))

	// post api request
	i, _ := strconv.Atoi(os.Args[2])
	res, err := client.Post.GetPost(os.Args[1], i)
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
	Post        *PostService
}

func NewClient(accessToken string) *Client {
	c := &Client{}
	c.Client = http.DefaultClient
	c.accessToken = accessToken
	c.baseURL = defaultBaseURL
	c.Team = &TeamService{client: c}
	c.Post = &PostService{client: c}

	return c
}

func (c *Client) createURL(path string) string {
	return c.baseURL + path + "?access_token=" + c.accessToken
}

func (c *Client) get(path string, query url.Values, v interface{}) (resp *http.Response, err error) {
	esaURL := c.createURL(path)
	queries := query.Encode()
	if len(queries) != 0 {
		esaURL += "&" + queries
	}

	res, err := c.Client.Get(esaURL)
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

// esa's post service
const (
	postPath = "/v1/teams"
)

type PostService struct {
	client *Client
}

// Post 記事
type Post struct {
	BodyMd   string   `json:"body_md"`
	Category string   `json:"category"`
	Message  string   `json:"message"`
	Name     string   `json:"name"`
	Tags     []string `json:"tags"`
	Wip      bool     `json:"wip"`
}

// PostResponse 記事のレスポンス
type PostResponse struct {
	BodyHTML      string `json:"body_html"`
	BodyMd        string `json:"body_md"`
	Category      string `json:"category"`
	CommentsCount int    `json:"comments_count"`
	CreatedAt     string `json:"created_at"`
	CreatedBy     struct {
		Icon       string `json:"icon"`
		Name       string `json:"name"`
		ScreenName string `json:"screen_name"`
	} `json:"created_by"`
	DoneTasksCount  int      `json:"done_tasks_count"`
	FullName        string   `json:"full_name"`
	Kind            string   `json:"kind"`
	Message         string   `json:"message"`
	Name            string   `json:"name"`
	Number          int      `json:"number"`
	OverLapped      bool     `json:"overlapped"`
	RevisionNumber  int      `json:"revision_number"`
	Star            bool     `json:"star"`
	StargazersCount int      `json:"stargazers_count"`
	Tags            []string `json:"tags"`
	TasksCount      int      `json:"tasks_count"`
	UpdatedAt       string   `json:"updated_at"`
	UpdatedBy       struct {
		Icon       string `json:"icon"`
		Name       string `json:"name"`
		ScreenName string `json:"screen_name"`
	} `json:"updated_by"`
	URL           string `json:"url"`
	Watch         bool   `json:"watch"`
	WatchersCount int    `json:"watchers_count"`
	Wip           bool   `json:"wip"`
}

// PostsResponse 複数記事のレスポンス
type PostsResponse struct {
	NextPage   interface{}    `json:"next_page"`
	Posts      []PostResponse `json:"posts"`
	PrevPage   interface{}    `json:"prev_page"`
	TotalCount int            `json:"total_count"`
}

// SharedPost 公開された記事
type SharedPost struct {
	HTML   string `json:"html"`
	Slides string `json:"slides"`
}

func createSearchQuery(query url.Values) string {
	var queries []string
	for key, values := range query {
		for _, value := range values {
			query := value
			if key != "" {
				query = key + ":" + query
			}
			queries = append(queries, query)
		}
	}

	return strings.Join(queries, " ")
}

func createQuery(query url.Values) url.Values {
	queries := url.Values{}
	searchQuery := query

	queryKey := []string{"page", "per_page", "q", "include", "sort", "order"}
	for _, key := range queryKey {
		if value := query.Get(key); value != "" {
			queries.Add(key, value)
			searchQuery.Del(key)
		}
	}

	queries.Add("q", createSearchQuery(searchQuery))
	return queries
}

// GetPosts チーム名とクエリを指定して記事を取得する
func (p *PostService) GetPosts(teamName string, query url.Values) (*PostsResponse, error) {
	var postsRes PostsResponse
	queries := createQuery(query)

	path := postPath + "/" + teamName + "/posts"
	res, err := p.client.get(path, queries, &postsRes)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	return &postsRes, nil

}

// GetPost チーム名と記事番号を指定して記事を取得する
func (p *PostService) GetPost(teamName string, postNumber int) (*PostResponse, error) {
	var postRes PostResponse

	postNumberStr := strconv.Itoa(postNumber)

	path := postPath + "/" + teamName + "/posts" + "/" + postNumberStr
	res, err := p.client.get(path, url.Values{}, &postRes)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	return &postRes, nil
}
