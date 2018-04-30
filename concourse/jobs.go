package concourse

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"strings"
)

type Pipeline struct {
	Name   string  `json:"name"`
	URL    string  `json:"url"`
	Paused bool    `json:"paused"`
	Groups []Group `json:"groups"`
}

type Group struct {
	Name     string   `json:"name"`
	JobNames []string `json:"jobs"`
}

type Build struct {
	Status       string `json:"status"`
	JobName      string `json:"job_name"`
	URL          string `json:"url"`
	PipelineName string `json:"pipeline_name"`
	Name         string `json:"name"`
	Team         string `json:"team_name"`
	ID           int    `json:"id"`
	TeamName     string `json:"team_name"`
	APIURL       string `json:"api_url"`
	StartTime    int    `json:"start_time"`
	EndTime      int    `json:"end_time"`
}

type Job struct {
	ID            int     `json:"id"`
	Name          string  `json:"name"`
	URL           string  `json:"url"`
	FinishedBuild Build   `json:"finished_build"`
	NextBuild     Build   `json:"next_build"`
	Inputs        []Input `json:"inputs"`
	//Groups        []Group  `json:"groups"`
	Outputs []Output `json:"output"`
}

type Resources struct {
	Inputs  []Input  `json:"inputs"`
	Outputs []Output `json:"outputs"`
}

type Metadata struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Version struct {
	Path   string `json:"path"`
	Number string `json:"number"`
}

type Input struct {
	Name            string     `json:"name"`
	Resource        string     `json:"resource"`
	Type            string     `json:"type"`
	Version         Version    `json:"version"`
	Metadata        []Metadata `json:"metadata"`
	PipelineID      int        `json:"pipeline_id"`
	FirstOccurrence bool       `json:"first_occurrence"`
	Trigger         bool       `json:"trigger"`
	Passed          []string   `json:"passed"`
}

type Output struct {
	Name     string     `json:"name"`
	ID       int        `json:"id"`
	Type     string     `json:"type"`
	Resource string     `json:"resource"`
	Version  Version    `json:"version"`
	Enabled  bool       `json:"enabled"`
	Metadata []Metadata `json:"metadata"`
}

type Client struct {
	host string
	team string

	tokenProvider TokenProvider
}

func New(host, team string, tokenProvider TokenProvider) Client {
	return Client{
		host:          strings.TrimSuffix(host, "/"),
		team:          team,
		tokenProvider: tokenProvider,
	}
}

func (c Client) performRequest(url string) (*http.Response, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := http.Client{Transport: tr, Timeout: 30 * time.Second}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	if c.tokenProvider != nil {
		authHeader, err := c.tokenProvider.GetAuthorizationHeader()
		if err != nil {
			return nil, err
		}

		req.Header.Set("Authorization", authHeader)
	}

	return client.Do(req)
}

func (c Client) JobBuilds(pipeline, job string) ([]Build, error) {
	resp, err := c.performRequest(fmt.Sprintf("%s/api/v1/teams/%s/pipelines/%s/jobs/%s/builds", c.host, c.team, pipeline, job))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var builds []Build
	if err := json.NewDecoder(resp.Body).Decode(&builds); err != nil {
		return nil, err
	}

	return builds, nil
}

func (c Client) BuildResources(id int) (Resources, error) {
	resp, err := c.performRequest(fmt.Sprintf("%s/api/v1/builds/%v/resources", c.host, id))
	if err != nil {
		return Resources{}, err
	}
	defer resp.Body.Close()

	var resources Resources
	if err := json.NewDecoder(resp.Body).Decode(&resources); err != nil {
		return Resources{}, err
	}

	return resources, nil
}

func (c Client) GetJobs(pipelineName string) ([]Job, error) {
	resp, err := c.performRequest(fmt.Sprintf("%s/api/v1/teams/%s/pipelines/%s/jobs", c.host, c.team, pipelineName))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var jobs []Job
	err = json.NewDecoder(resp.Body).Decode(&jobs)
	if err != nil {
		return nil, err
	}

	return jobs, nil
}
