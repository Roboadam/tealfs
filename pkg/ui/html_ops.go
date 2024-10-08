// Copyright (C) 2024 Adam Hess
//
// This program is free software: you can redistribute it and/or modify it under
// the terms of the GNU Affero General Public License as published by the Free
// Software Foundation, version 3.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE. See the GNU Affero General Public License
// for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

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
	Handlers map[string]func(http.ResponseWriter, *http.Request)
}

func (m *MockHtmlOps) ListenAndServe(addr string) error {
	m.BindAddr = addr
	return nil
}

func (m *MockHtmlOps) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	m.Handlers[pattern] = handler
}

type MockResponseWriter struct {
	WrittenData string
}

func (m *MockResponseWriter) Header() http.Header {
	return make(http.Header)
}

func (m *MockResponseWriter) Write(data []byte) (int, error) {
	m.WrittenData = string(data)
	return len(data), nil
}

func (m *MockResponseWriter) WriteHeader(statusCode int) {}
