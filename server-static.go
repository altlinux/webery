/*
* Copyright (C) 2015 Alexey Gladkov <gladkov.alexey@gmail.com>
*
* This file is covered by the GNU General Public License,
* which should be included with webery as the file COPYING.
 */

package main

import (
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func httpStatusError(err error) int {
	//	if _, ok := err.(KhpError); ok {
	//		return http.StatusServiceUnavailable
	//	}
	return http.StatusInternalServerError
}

func (s *Server) rootHandler(w *HTTPResponse, r *http.Request, p *url.Values) {
	filename := s.Cfg.Content.Path + "/index.html"

	file, err := os.Open(filename)
	if err != nil {
		if os.IsExist(err) {
			s.errorResponse(w, httpStatusError(err), "%+v", err)
		} else {
			s.notFoundHandler(w, r, p)
		}
		return
	}
	defer file.Close()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	s.writeStatus(w, http.StatusOK)
	io.Copy(w, file)
}

func (s *Server) staticHandler(w *HTTPResponse, r *http.Request, p *url.Values) {
	filename := s.Cfg.Content.Path + "/" + p.Get("path")

	file, err := os.Open(filename)
	if err != nil {
		if os.IsExist(err) {
			s.errorResponse(w, httpStatusError(err), "%+v", err)
		} else {
			s.notFoundHandler(w, r, p)
		}
		return
	}
	defer file.Close()

	if strings.HasSuffix(p.Get("path"), ".css") {
		w.Header().Set("Content-Type", "text/css")
	} else if strings.HasSuffix(p.Get("path"), ".js") {
		w.Header().Set("Content-Type", "text/javascript")
	} else {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	}

	s.writeStatus(w, http.StatusOK)
	io.Copy(w, file)
}

func (s *Server) pingHandler(w *HTTPResponse, r *http.Request, p *url.Values) {
	w.WriteHeader(http.StatusOK)
}

func (s *Server) notFoundHandler(w *HTTPResponse, r *http.Request, p *url.Values) {
	s.errorResponse(w, http.StatusNotFound, "404 page not found")
}

func (s *Server) notAllowedHandler(w *HTTPResponse, r *http.Request, p *url.Values) {
	s.errorResponse(w, http.StatusMethodNotAllowed, "405 Method Not Allowed")
}
