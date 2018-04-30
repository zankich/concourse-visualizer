package concourse_test

import (
	"net/http"
	"net/http/httptest"

	"github.com/onsi/gomega/ghttp"
	"github.com/zankich/concourse-visualizer/concourse"
	"github.com/zankich/concourse-visualizer/concourse/concoursefakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	oldAPIPipelines = `[
  {
    "name": "p1",
    "url": "/teams/main/pipelines/p1",
    "paused": false,
    "public": true,
    "team_name": "main"
  },
  {
    "name": "p2",
    "url": "/teams/main/pipelines/p2",
    "paused": false,
    "public": true,
    "team_name": "main"
  },
  {
    "name": "p3",
    "url": "/teams/main/pipelines/p3",
    "paused": true,
    "public": true,
    "team_name": "main"
  }
]`

	oldAPIPipeline1Jobs = `[
  {
    "finished_build": {
      "status": "finished",
      "job_name": "g1j1",
      "url": "/teams/main/pipelines/p1/jobs/g1j1/builds/1",
      "pipeline_name": "p1"
    }
  },
  {
    "finished_build": {
      "status": "finished",
      "job_name": "g1j2",
      "url": "/teams/main/pipelines/p1/jobs/g1j2/builds/1",
      "pipeline_name": "p1"
    }
  },
  {
    "finished_build": {
      "status": "finished",
      "job_name": "gBothj3",
      "url": "/teams/main/pipelines/p1/jobs/gBothj3/builds/1",
      "pipeline_name": "p1"
    }
  },
  {
    "finished_build": {
      "status": "finished",
      "job_name": "g2j1",
      "url": "/teams/main/pipelines/p1/jobs/g2j1/builds/1",
      "pipeline_name": "p1"
    }
  },
  {
    "finished_build": {
      "status": "finished",
      "job_name": "g2j2",
      "url": "/teams/main/pipelines/p1/jobs/g2j2/builds/1",
      "pipeline_name": "p1"
    }
  }
]`

	oldAPIPipeline2Jobs = `[
  {
    "finished_build": {
      "status": "finished",
      "job_name": "g1j1",
      "url": "/teams/main/pipelines/p2/jobs/g1j1/builds/1",
      "pipeline_name": "p2"
    }
  },
  {
    "finished_build": {
      "status": "finished",
      "job_name": "g1j2",
      "url": "/teams/main/pipelines/p2/jobs/g1j2/builds/1",
      "pipeline_name": "p2"
    }
  }
]`

	newAPIpipelines = `[
  {
    "name": "newAPI",
    "url": "/teams/main/pipelines/newAPI",
    "paused": false,
    "public": true,
    "team_name": "main"
  }
]`

	newAPIJobs = `[
  {
    "finished_build": {
      "name": "1",
      "job_name": "newAPIJob",
      "team_name": "main",
      "pipeline_name": "newAPI",
      "status": "succeeded"
    }
  }
]`
)

var _ = Describe("GetJobs", func() {
	Context("New API", func() {
		It("handles the new concourse API", func() {
			ts := ghttp.NewServer()

			newAPIJob := concourse.Job{
				FinishedBuild: concourse.Build{
					Status:       "succeeded",
					JobName:      "newAPIJob",
					URL:          ts.URL() + "/teams/main/pipelines/newAPI/jobs/newAPIJob/builds/1",
					PipelineName: "newAPI",
					Name:         "1",
					Team:         "main",
				},
			}

			nilAuthHeader := http.Header{"Authorization": nil}
			ts.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/v1/teams/main/pipelines"),
					ghttp.VerifyHeader(nilAuthHeader),
					ghttp.RespondWith(http.StatusOK, newAPIpipelines),
				),

				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/v1/teams/main/pipelines/newAPI/jobs"),
					ghttp.VerifyHeader(nilAuthHeader),
					ghttp.RespondWith(http.StatusOK, newAPIJobs),
				),
			)

			client := concourse.New(ts.URL(), "main", nil)
			jobs, err := client.GetJobs()

			Expect(err).NotTo(HaveOccurred())
			Expect(jobs).To(ConsistOf(
				newAPIJob,
			))
		})
	})

	Context("Successful reponses", func() {
		var (
			job1           concourse.Job
			job2           concourse.Job
			job3           concourse.Job
			job4           concourse.Job
			job5           concourse.Job
			job6           concourse.Job
			jobInTwoGroups concourse.Job
			ts             *ghttp.Server
		)

		BeforeEach(func() {
			ts = ghttp.NewServer()
			job1 = concourse.Job{
				FinishedBuild: concourse.Build{
					Status:       "finished",
					JobName:      "g1j1",
					URL:          ts.URL() + "/teams/main/pipelines/p1/jobs/g1j1/builds/1",
					PipelineName: "p1",
					Name:         "",
					Team:         "",
				},
			}
			job2 = concourse.Job{
				FinishedBuild: concourse.Build{
					Status:       "finished",
					JobName:      "g1j2",
					URL:          ts.URL() + "/teams/main/pipelines/p1/jobs/g1j2/builds/1",
					PipelineName: "p1",
					Name:         "",
					Team:         "",
				},
			}
			job3 = concourse.Job{
				FinishedBuild: concourse.Build{
					Status:       "finished",
					JobName:      "g2j1",
					URL:          ts.URL() + "/teams/main/pipelines/p1/jobs/g2j1/builds/1",
					PipelineName: "p1",
					Name:         "",
					Team:         "",
				},
			}
			job4 = concourse.Job{
				FinishedBuild: concourse.Build{
					Status:       "finished",
					JobName:      "g2j2",
					URL:          ts.URL() + "/teams/main/pipelines/p1/jobs/g2j2/builds/1",
					PipelineName: "p1",
					Name:         "",
					Team:         "",
				},
			}
			jobInTwoGroups = concourse.Job{
				FinishedBuild: concourse.Build{
					Status:       "finished",
					JobName:      "gBothj3",
					URL:          ts.URL() + "/teams/main/pipelines/p1/jobs/gBothj3/builds/1",
					PipelineName: "p1",
					Name:         "",
					Team:         "",
				},
			}

			job5 = concourse.Job{
				FinishedBuild: concourse.Build{
					Status:       "finished",
					JobName:      "g1j1",
					URL:          ts.URL() + "/teams/main/pipelines/p2/jobs/g1j1/builds/1",
					PipelineName: "p2",
					Name:         "",
					Team:         "",
				},
			}
			job6 = concourse.Job{
				FinishedBuild: concourse.Build{
					Status:       "finished",
					JobName:      "g1j2",
					URL:          ts.URL() + "/teams/main/pipelines/p2/jobs/g1j2/builds/1",
					PipelineName: "p2",
					Name:         "",
					Team:         "",
				},
			}
		})

		AfterEach(func() {
			ts.Close()
		})

		It("returns a list of public jobs for a given team", func() {
			nilAuthHeader := http.Header{"Authorization": nil}
			ts.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/v1/teams/main/pipelines"),
					ghttp.VerifyHeader(nilAuthHeader),
					ghttp.RespondWith(http.StatusOK, oldAPIPipelines),
				),

				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/v1/teams/main/pipelines/p1/jobs"),
					ghttp.VerifyHeader(nilAuthHeader),
					ghttp.RespondWith(http.StatusOK, oldAPIPipeline1Jobs),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/v1/teams/main/pipelines/p2/jobs"),
					ghttp.VerifyHeader(nilAuthHeader),
					ghttp.RespondWith(http.StatusOK, oldAPIPipeline2Jobs),
				),
			)

			client := concourse.New(ts.URL(), "main", nil)
			jobs, err := client.GetJobs()

			Expect(err).NotTo(HaveOccurred())
			Expect(jobs).To(ConsistOf(
				job1,
				job2,
				jobInTwoGroups,
				job3,
				job4,
				job5,
				job6,
			))
		})

		It("returns a list of jobs when authorization is required", func() {
			ts.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/v1/teams/main/pipelines"),
					ghttp.VerifyHeaderKV("Authorization", "Bearer token"),
					ghttp.RespondWith(http.StatusOK, oldAPIPipelines),
				),

				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/v1/teams/main/pipelines/p1/jobs"),
					ghttp.VerifyHeaderKV("Authorization", "Bearer token"),
					ghttp.RespondWith(http.StatusOK, oldAPIPipeline1Jobs),
				),

				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/v1/teams/main/pipelines/p2/jobs"),
					ghttp.VerifyHeaderKV("Authorization", "Bearer token"),
					ghttp.RespondWith(http.StatusOK, oldAPIPipeline2Jobs),
				),
			)

			tokenProvider := &concoursefakes.FakeTokenProvider{}
			tokenProvider.GetAuthorizationHeaderReturns("Bearer token", nil)

			client := concourse.New(ts.URL(), "main", tokenProvider)
			jobs, err := client.GetJobs()

			Expect(err).NotTo(HaveOccurred())
			Expect(tokenProvider.GetAuthorizationHeaderCallCount()).To(Equal(3))
			Expect(jobs).To(ConsistOf(
				job1,
				job2,
				jobInTwoGroups,
				job3,
				job4,
				job5,
				job6,
			))
		})
	})

	Context("failure cases", func() {
		It("returns an error when the host is bad", func() {
			client := concourse.New("%%%%%", "main", nil)
			_, err := client.GetJobs()

			Expect(err).To(MatchError(ContainSubstring("invalid URL escape")))
		})

		It("returns an error when the pipeline json is bad", func() {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == "GET" && r.URL.Path == "/api/v1/teams/main/pipelines" {
					w.Write([]byte("%%%%%%%%"))
					return
				}

				w.WriteHeader(http.StatusTeapot)
			}))
			defer ts.Close()
			client := concourse.New(ts.URL, "main", nil)

			_, err := client.GetJobs()
			Expect(err).To(MatchError("invalid character '%' looking for beginning of value"))
		})

		It("returns an error when the pipeline job json is bad", func() {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == "GET" && r.URL.Path == "/api/v1/teams/main/pipelines" {
					w.Write([]byte(oldAPIPipelines))
					return
				} else if r.Method == "GET" && r.URL.Path == "/api/v1/teams/main/pipelines/p1/jobs" {
					w.Write([]byte("%%%%%%%%"))
					return
				}

				w.WriteHeader(http.StatusTeapot)
			}))
			defer ts.Close()
			client := concourse.New(ts.URL, "main", nil)

			_, err := client.GetJobs()
			Expect(err).To(MatchError("invalid character '%' looking for beginning of value"))
		})
	})
})
