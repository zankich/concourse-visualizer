package concourse_test

import (
	"github.com/zankich/concourse-visualizer/concourse"

	"net/http"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	"github.com/benbjohnson/clock"
)

var _ = Describe("Auth", func() {
	Context("#GetAuthorizationHeader", func() {
		var (
			authServer *ghttp.Server
			mockClock  *clock.Mock

			tp concourse.TokenProvider
		)

		BeforeEach(func() {
			authServer = ghttp.NewServer()
			authServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/v1/teams/main/auth/token"),
					ghttp.VerifyBasicAuth("test-user", "test-password"),
					ghttp.RespondWith(http.StatusOK, `{"type":"Bearer","value":"token"}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/v1/teams/main/auth/token"),
					ghttp.VerifyBasicAuth("test-user", "test-password"),
					ghttp.RespondWith(http.StatusOK, `{"type":"Bearer","value":"updated-token"}`),
				),
			)
			mockClock = clock.NewMock()

			tp = concourse.NewTokenProvider(authServer.URL(), "main", "test-user", "test-password", mockClock)

		})

		It("returns oauth token to be used for requests to the concourse API", func() {
			token, err := tp.GetAuthorizationHeader()
			Expect(err).ToNot(HaveOccurred())

			Expect(token).To(Equal("Bearer token"))
		})

		It("caches the token", func() {
			for i := 0; i < 10; i++ {
				token, err := tp.GetAuthorizationHeader()
				Expect(err).ToNot(HaveOccurred())
				Expect(token).To(Equal("Bearer token"))
			}

			Expect(authServer.ReceivedRequests()).To(HaveLen(1))
		})

		It("refreshes the cached token after 23 hours", func() {
			token, err := tp.GetAuthorizationHeader()
			Expect(err).ToNot(HaveOccurred())
			Expect(token).To(Equal("Bearer token"))

			mockClock.Add(23 * time.Hour)
			mockClock.Add(time.Nanosecond)

			token, err = tp.GetAuthorizationHeader()
			Expect(err).ToNot(HaveOccurred())
			Expect(token).To(Equal("Bearer updated-token"))

			Expect(authServer.ReceivedRequests()).To(HaveLen(2))
		})
	})

})
