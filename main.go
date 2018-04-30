package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/benbjohnson/clock"
	"github.com/zankich/concourse-visualizer/concourse"
)

func main() {
	host := os.Getenv("CONCOURSE_HOST")
	team := os.Getenv("CONCOURSE_TEAM")
	username := os.Getenv("CONCOURSE_USERNAME")
	password := os.Getenv("CONCOURSE_PASSWORD")

	tokenProvider := concourse.NewTokenProvider(host, team, username, password, clock.New())

	buildNumber := os.Args[1]
	pipeline := os.Args[2]

	concourseClient := concourse.New(host, team, tokenProvider)

	jobs, err := concourseClient.GetJobs(pipeline)
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup

	flow := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		flow <- true
	}

	fmt.Println(len(jobs))
	for _, job := range jobs {
		wg.Add(1)
		go checkProductVersion(job.Name, buildNumber, wg, pipeline, concourseClient, flow)
	}

	wg.Wait()
}

func checkProductVersion(jobName string, buildNumber string, wg sync.WaitGroup, pipeline string, concourseClient concourse.Client, flow chan bool) {
	defer func() {
		flow <- true
		wg.Done()
	}()

	<-flow

	builds, err := concourseClient.JobBuilds(pipeline, jobName)
	if err != nil {
		panic(err)
	}

	for _, build := range builds {
		found := false

		r, e := concourseClient.BuildResources(build.ID)
		if e != nil {
			panic(err)
		}

		for _, input := range r.Inputs {
			if input.Resource == "product-version" {
				if input.Version.Number == buildNumber {
					found = true
					fmt.Printf("%s/%s has %s\n", pipeline, jobName, build.Status)
					break
				}
			}
		}

		if found {
			break
		}
	}
}
