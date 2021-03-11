// Package httpmock contains code for creating simple mock HTTP servers.
package httpmock

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/codahale/gubbins/assert"
	"github.com/google/go-cmp/cmp/cmpopts"
)

type Server struct {
	m            sync.Mutex
	srv          *httptest.Server
	tb           testing.TB
	expectations []expectation
}

type expectation struct {
	url      url.URL
	method   string
	req      string
	status   int
	resp     string
	optional bool
	called   bool
}

func NewServer(tb testing.TB) *Server {
	tb.Helper()

	server := &Server{tb: tb}
	server.srv = httptest.NewServer(http.HandlerFunc(server.handle))

	return server
}

func (s *Server) handle(w http.ResponseWriter, r *http.Request) {
	s.tb.Helper()
	s.m.Lock()
	defer s.m.Unlock()

	for i, exp := range s.expectations {
		if !exp.called && *r.URL == exp.url {
			s.expectations[i].called = true
			s.checkExpectation(w, r, exp)

			return
		}
	}

	http.NotFound(w, r)
	s.tb.Errorf("Unexpected request for %q", r.URL.String())
}

func (s *Server) checkExpectation(w http.ResponseWriter, r *http.Request, exp expectation) {
	if exp.method != "" {
		assert.Equal(s.tb, "method", exp.method, r.Method,
			cmpopts.AcyclicTransformer("ToUpper", strings.ToUpper))
	}

	if exp.req != "" {
		req, err := ioutil.ReadAll(r.Body)
		if err != nil {
			s.tb.Fatal(err)
		}

		_ = r.Body.Close()

		assert.Equal(s.tb, "request", exp.req, string(req),
			cmpopts.AcyclicTransformer("TrimSpace", strings.TrimSpace))
	}

	if exp.status != 0 {
		w.WriteHeader(exp.status)
	}

	if exp.resp != "" {
		_, err := io.WriteString(w, exp.resp)
		if err != nil {
			s.tb.Fatal(err)
		}
	}
}

func (s *Server) URL() string {
	s.tb.Helper()
	return s.srv.URL
}

func (s *Server) Client() *http.Client {
	s.tb.Helper()
	return s.srv.Client()
}

func (s *Server) Expect(reqURL string, opt ...Option) {
	s.tb.Helper()
	s.m.Lock()
	defer s.m.Unlock()

	u, err := url.Parse(reqURL)
	if err != nil {
		s.tb.Fatal(err)
	}

	e := expectation{
		url: *u,
	}

	for _, f := range opt {
		f(&e)
	}

	s.expectations = append(s.expectations, e)
}

func (s *Server) Finish() {
	s.tb.Helper()
	s.m.Lock()
	defer s.m.Unlock()

	for _, exp := range s.expectations {
		if !exp.optional && !exp.called {
			s.tb.Errorf("No request for %q", exp.url.String())
		}
	}
}

type Option func(e *expectation)

func RespJSON(resp interface{}) Option {
	j, err := json.Marshal(resp)
	if err != nil {
		panic(err)
	}

	return func(e *expectation) {
		e.resp = string(j)
	}
}

func Method(method string) Option {
	return func(e *expectation) {
		e.method = method
	}
}

func Status(status int) Option {
	return func(e *expectation) {
		e.status = status
	}
}

func ReqJSON(req interface{}) Option {
	j, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}

	return func(e *expectation) {
		e.req = string(j)
	}
}

func Optional() Option {
	return func(e *expectation) {
		e.optional = true
	}
}
