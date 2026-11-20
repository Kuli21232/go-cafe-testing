package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCafeNegative(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		request string
		status  int
		message string
	}{
		{"/cafe", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=omsk", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=tula&count=na", http.StatusBadRequest, "incorrect count"},
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v.request, nil)
		handler.ServeHTTP(response, req)

		assert.Equal(t, v.status, response.Code)
		assert.Equal(t, v.message, strings.TrimSpace(response.Body.String()))
	}
}

func TestCafeWhenOk(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []string{
		"/cafe?count=2&city=moscow",
		"/cafe?city=tula",
		"/cafe?city=moscow&search=ложка",
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v, nil)

		handler.ServeHTTP(response, req)

		assert.Equal(t, http.StatusOK, response.Code)
	}
}

func TestCafeCount(t *testing.T) {
	requests := []struct {
		count int
		want  int
	}{
		{count: 0, want: 0},
		{count: 1, want: 1},
		{count: 2, want: 2},
		{count: 100, want: len(cafeList["moscow"])},
	}

	handler := http.HandlerFunc(mainHandle)
	for _, v := range requests {
		t.Run("count="+strconv.Itoa(v.count), func(t *testing.T) {
			response := httptest.NewRecorder()
			requestURL := "/cafe?city=moscow&count=" + strconv.Itoa(v.count)
			req := httptest.NewRequest(http.MethodGet, requestURL, nil)
			handler.ServeHTTP(response, req)

			require.Equal(t, http.StatusOK, response.Code)

			var cafes []string
			if body := response.Body.String(); body != "" {
				cafes = strings.Split(body, ",")
			}
			assert.Len(t, cafes, v.want)
		})
	}
}

func TestCafeSearch(t *testing.T) {
	requests := []struct {
		search    string
		wantCount int
	}{
		{search: "фасоль", wantCount: 0},
		{search: "кофе", wantCount: 2},
		{search: "вилка", wantCount: 1},
		{search: "КОФЕ", wantCount: 2},
		{search: "ВиЛкА", wantCount: 1},
	}

	handler := http.HandlerFunc(mainHandle)
	for _, v := range requests {
		t.Run(v.search, func(t *testing.T) {
			response := httptest.NewRecorder()
			requestURL := "/cafe?city=moscow&search=" + url.QueryEscape(v.search)
			req := httptest.NewRequest(http.MethodGet, requestURL, nil)
			handler.ServeHTTP(response, req)

			require.Equal(t, http.StatusOK, response.Code)

			var cafes []string
			if body := response.Body.String(); body != "" {
				cafes = strings.Split(body, ",")
			}
			assert.Len(t, cafes, v.wantCount)
			for _, cafe := range cafes {
				assert.True(t, strings.Contains(strings.ToLower(cafe), strings.ToLower(v.search)),
					"кафе %q не содержит подстроку %q", cafe, v.search)
			}
		})
	}
}
