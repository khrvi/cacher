package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	l "log"
	"net/http"
	"os"
	"strings"
	"testing"

	"./cache"
	"github.com/labstack/echo"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/ghttp"
)

// Port Cacher listens on for testing...differs from default port in dev
const (
	Port = "8081"
)

var authToken = "token"

// Run once for all tests
var _ = BeforeSuite(func() {
	prepareLogger()
	l.SetOutput(os.Stdout)
	server := httpServer()
	go server.Start("localhost:" + Port)
})

func TestHTTPServer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cacher HTTP Interface Suite")
}

// basic Cacherg HTTP client
type CacherClient struct {
	client *http.Client
	port   string
}

// Read HTTP response
type HTTPResponse struct {
	Body    string
	Status  int
	Headers http.Header
	Cookies []*http.Cookie
}

var _ = Describe("keys", func() {

	var do *ghttp.Server
	var client *CacherClient
	var response *HTTPResponse
	var err error

	BeforeEach(func() {
		do = ghttp.NewServer()
		client = NewCacherClient()
		cacheManager, _ = cache.New("mutex-map", log, false, 60, false)
	})

	AfterEach(func() {
		do.Close()
	})

	Describe("listing keys", func() {
		BeforeEach(func() {
			response, err = client.Get("/keys")
		})

		It("no error occured", func() {
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns 200 status code", func() {
			Ω(response.Status).Should(Equal(200))
		})

		It("return empty list", func() {
			keys := []string{}

			r := Response{
				Status: "ok",
				Value:  keys,
			}
			expected, err := json.Marshal(r)
			Expect(err).NotTo(HaveOccurred())
			Ω(response.Body).Should(MatchJSON(expected))
		})
	})

	Describe("listing keys (not empty)", func() {
		BeforeEach(func() {
			cacheManager.Set("test", 3, 0)
			response, err = client.Get("/keys")
		})

		It("no error occured", func() {
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns 200 status code", func() {
			Ω(response.Status).Should(Equal(200))
		})

		It("lists one key", func() {
			keys := []string{"test"}

			r := Response{
				Status: "ok",
				Value:  keys,
			}
			expected, err := json.Marshal(r)
			Expect(err).NotTo(HaveOccurred())
			Ω(response.Body).Should(MatchJSON(expected))
		})
	})

	Describe("getting key", func() {
		BeforeEach(func() {
			cacheManager.Set("test", 3, 0)
			response, err = client.Get("/test")
		})

		It("no error occured", func() {
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns 200 status code", func() {
			Ω(response.Status).Should(Equal(200))
		})

		It("return stored valie", func() {
			value := 3

			r := Response{
				Status: "ok",
				Value:  value,
			}
			expected, err := json.Marshal(r)
			Expect(err).NotTo(HaveOccurred())
			Ω(response.Body).Should(MatchJSON(expected))
		})
	})

	Describe("getting missed key", func() {
		BeforeEach(func() {
			cacheManager.Set("another_key", 3, 0)
			response, err = client.Get("/test")
		})

		It("no error occured", func() {
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns 400 status code", func() {
			Ω(response.Status).Should(Equal(400))
		})

		It("return error message", func() {
			value := "Key 'test' not found in cache."

			r := Response{
				Status:       "error",
				ErrorMessage: value,
			}
			expected, err := json.Marshal(r)
			Expect(err).NotTo(HaveOccurred())
			Ω(response.Body).Should(MatchJSON(expected))
		})
	})

	Describe("setting key", func() {
		BeforeEach(func() {
			response, err = client.Post("/", "{\"key\":\"test_string\",\"value\":4,\"ttl\":0}")
		})

		It("no error occured", func() {
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns 200 status code", func() {
			Ω(response.Status).Should(Equal(200))
		})

		It("store value in cache", func() {
			r := Response{Status: "ok"}
			expected, err := json.Marshal(r)
			Expect(err).NotTo(HaveOccurred())
			Ω(response.Body).Should(MatchJSON(expected))
		})
	})

	Describe("setting unprocessable data", func() {
		BeforeEach(func() {
			response, err = client.Post("/", "wrong_json")
		})

		It("returns 200 status code", func() {
			Ω(response.Status).Should(Equal(400))
		})
	})

	Describe("deliting key", func() {
		BeforeEach(func() {
			cacheManager.Set("test", 3, 0)
			response, err = client.Delete("/test")
		})

		It("no error occured", func() {
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns 200 status code", func() {
			Ω(response.Status).Should(Equal(200))
		})

		It("return stored valie", func() {
			_, _, found, _ := cacheManager.Get("test")
			Expect(found).To(BeFalse())

			r := Response{Status: "ok"}
			expected, err := json.Marshal(r)
			Expect(err).NotTo(HaveOccurred())
			Ω(response.Body).Should(MatchJSON(expected))
		})
	})
})

// Instantiate new http client
func NewCacherClient() *CacherClient {
	return &CacherClient{
		client: http.DefaultClient,
		port:   Port,
	}
}

// Send GET request
func (c *CacherClient) Get(url string) (*HTTPResponse, error) {
	return c.do("GET", url, "")
}

// Send POST request
func (c *CacherClient) Post(url, body string) (*HTTPResponse, error) {
	return c.do("POST", url, body)
}

// Send DELETE request
func (c *CacherClient) Delete(url string) (*HTTPResponse, error) {
	return c.do("DELETE", url, "")
}

// Helper generic send request method
func (c *CacherClient) do(verb, url, body string) (*HTTPResponse, error) {
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req, err := http.NewRequest(verb, "http://localhost:"+c.port+url, reader)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+authToken)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return &HTTPResponse{
		Body:    string(respBody),
		Status:  resp.StatusCode,
		Headers: resp.Header,
		Cookies: resp.Cookies(),
	}, nil
}

// Factory method for application
// Makes it possible to do integration testing.
func httpServer() *echo.Echo {
	e := echo.New()

	// Routes
	// e.GET("/", healthCheck)
	e.GET("/:key", getValue)
	e.POST("/", setValue)
	e.DELETE("/:key", deleteValue)
	e.GET("/keys", getAllKeys)

	return e
}
