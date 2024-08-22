package ui

import "net/http"

type HtmlOps interface {
	ListenAndServe(addr string) error
	HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request))
}

type HttpHtmlOps struct{}

func (h *HttpHtmlOps) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, nil)
}

func (h *HttpHtmlOps) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	http.HandleFunc(pattern, handler)
}

type MockHtmlOps struct {
	BindAddr string
	handlers map[string]func(http.ResponseWriter, *http.Request)
}

func (m *MockHtmlOps) ListenAndServe(addr string) error {
	m.BindAddr = addr
	return nil
}

func (m *MockHtmlOps) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	m.handlers[pattern] = handler
}