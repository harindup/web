package web

import (
	"fmt"
	. "launchpad.net/gocheck"
	"net/http"
	"strings"
)

type ErrorTestSuite struct{}

var _ = Suite(&ErrorTestSuite{})

func ErrorHandlerWithNoContext(w ResponseWriter, r *Request, err interface{}) {
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, "Contextless Error")
}

func (s *ErrorTestSuite) TestNoErorHandler(c *C) {
	router := New(Context{})
	router.Get("/action", (*Context).ErrorAction)

	admin := router.Subrouter(AdminContext{}, "/admin")
	admin.Get("/action", (*AdminContext).ErrorAction)

	rw, req := newTestRequest("GET", "/action")
	router.ServeHTTP(rw, req)
	c.Assert(strings.TrimSpace(string(rw.Body.Bytes())), Equals, "Application Error")
	c.Assert(rw.Code, Equals, http.StatusInternalServerError)

	rw, req = newTestRequest("GET", "/admin/action")
	router.ServeHTTP(rw, req)
	c.Assert(strings.TrimSpace(string(rw.Body.Bytes())), Equals, "Application Error")
	c.Assert(rw.Code, Equals, http.StatusInternalServerError)
}

func (s *ErrorTestSuite) TestHandlerOnRoot(c *C) {
	router := New(Context{})
	router.Error((*Context).ErrorHandler)
	router.Get("/action", (*Context).ErrorAction)

	admin := router.Subrouter(AdminContext{}, "/admin")
	admin.Get("/action", (*AdminContext).ErrorAction)

	rw, req := newTestRequest("GET", "/action")
	router.ServeHTTP(rw, req)
	c.Assert(strings.TrimSpace(string(rw.Body.Bytes())), Equals, "My Error")
	c.Assert(rw.Code, Equals, http.StatusInternalServerError)

	rw, req = newTestRequest("GET", "/admin/action")
	router.ServeHTTP(rw, req)
	c.Assert(strings.TrimSpace(string(rw.Body.Bytes())), Equals, "My Error")
	c.Assert(rw.Code, Equals, http.StatusInternalServerError)
}

func (s *ErrorTestSuite) TestContextlessError(c *C) {
	router := New(Context{})
	router.Error(ErrorHandlerWithNoContext)
	router.Get("/action", (*Context).ErrorAction)

	admin := router.Subrouter(AdminContext{}, "/admin")
	admin.Get("/action", (*AdminContext).ErrorAction)

	rw, req := newTestRequest("GET", "/action")
	router.ServeHTTP(rw, req)
	c.Assert(strings.TrimSpace(string(rw.Body.Bytes())), Equals, "Contextless Error")
	c.Assert(rw.Code, Equals, http.StatusInternalServerError)

	rw, req = newTestRequest("GET", "/admin/action")
	router.ServeHTTP(rw, req)
	c.Assert(strings.TrimSpace(string(rw.Body.Bytes())), Equals, "Contextless Error")
	c.Assert(rw.Code, Equals, http.StatusInternalServerError)
}

func (s *ErrorTestSuite) TestMultipleErrorHandlers(c *C) {
	router := New(Context{})
	router.Error((*Context).ErrorHandler)
	router.Get("/action", (*Context).ErrorAction)

	admin := router.Subrouter(AdminContext{}, "/admin")
	admin.Error((*AdminContext).ErrorHandler)
	admin.Get("/action", (*AdminContext).ErrorAction)

	rw, req := newTestRequest("GET", "/action")
	router.ServeHTTP(rw, req)
	c.Assert(strings.TrimSpace(string(rw.Body.Bytes())), Equals, "My Error")
	c.Assert(rw.Code, Equals, http.StatusInternalServerError)

	rw, req = newTestRequest("GET", "/admin/action")
	router.ServeHTTP(rw, req)
	c.Assert(strings.TrimSpace(string(rw.Body.Bytes())), Equals, "Admin Error")
	c.Assert(rw.Code, Equals, http.StatusInternalServerError)
}

func (s *ErrorTestSuite) TestMultipleErrorHandlers2(c *C) {
	router := New(Context{})
	router.Get("/action", (*Context).ErrorAction)

	admin := router.Subrouter(AdminContext{}, "/admin")
	admin.Error((*AdminContext).ErrorHandler)
	admin.Get("/action", (*AdminContext).ErrorAction)

	api := router.Subrouter(APIContext{}, "/api")
	api.Error((*APIContext).ErrorHandler)
	api.Get("/action", (*APIContext).ErrorAction)

	rw, req := newTestRequest("GET", "/action")
	router.ServeHTTP(rw, req)
	c.Assert(strings.TrimSpace(string(rw.Body.Bytes())), Equals, "Application Error")
	c.Assert(rw.Code, Equals, http.StatusInternalServerError)

	rw, req = newTestRequest("GET", "/admin/action")
	router.ServeHTTP(rw, req)
	c.Assert(strings.TrimSpace(string(rw.Body.Bytes())), Equals, "Admin Error")
	c.Assert(rw.Code, Equals, http.StatusInternalServerError)

	rw, req = newTestRequest("GET", "/api/action")
	router.ServeHTTP(rw, req)
	c.Assert(strings.TrimSpace(string(rw.Body.Bytes())), Equals, "Api Error")
	c.Assert(rw.Code, Equals, http.StatusInternalServerError)
}

func (s *ErrorTestSuite) TestRootMiddlewarePanic(c *C) {
	router := New(Context{})
	router.Middleware((*Context).ErrorMiddleware)
	router.Error((*Context).ErrorHandler)
	admin := router.Subrouter(AdminContext{}, "/admin")
	admin.Error((*AdminContext).ErrorHandler)
	admin.Get("/action", (*AdminContext).ErrorAction)

	rw, req := newTestRequest("GET", "/admin/action")
	router.ServeHTTP(rw, req)
	assertResponse(c, rw, "My Error", 500)
}

func (s *ErrorTestSuite) TestNonRootMiddlewarePanic(c *C) {
	router := New(Context{})
	router.Error((*Context).ErrorHandler)
	admin := router.Subrouter(AdminContext{}, "/admin")
	admin.Middleware((*AdminContext).ErrorMiddleware)
	admin.Error((*AdminContext).ErrorHandler)
	admin.Get("/action", (*AdminContext).ErrorAction)

	rw, req := newTestRequest("GET", "/admin/action")
	router.ServeHTTP(rw, req)
	assertResponse(c, rw, "Admin Error", 500)
}

func (s *ErrorTestSuite) TestConsistentContext(c *C) {
	router := New(Context{})
	router.Error((*Context).ErrorHandler)
	admin := router.Subrouter(Context{}, "/admin")
	admin.Error((*Context).ErrorHandlerSecondary)
	admin.Get("/foo", (*Context).ErrorAction)

	rw, req := newTestRequest("GET", "/admin/foo")
	router.ServeHTTP(rw, req)
	assertResponse(c, rw, "My Secondary Error", 500)
}
