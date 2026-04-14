package rest

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

type redirectRoundTripper struct {
	targetURL string
}

func (r *redirectRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Reconstruct the URL to point to our test server
	// We keep the path and query string
	newReq := req.Clone(req.Context())
	newReq.URL.Scheme = "http"
	newReq.URL.Host = strings.TrimPrefix(r.targetURL, "http://")
	return http.DefaultTransport.RoundTrip(newReq)
}

func TestRestReleasesCreate(t *testing.T) {
	mockResponse := ReleasesCreateReturn{
		Name: "sites/test-site/releases/test-release",
		Type: "DEPLOY",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected method POST, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/sites/test-site/releases") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("versionName") != "test-version" {
			t.Errorf("expected versionName test-version, got %s", r.URL.Query().Get("versionName"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	client := &AuthorizedHTTPClient{
		Client: &http.Client{
			Transport: &redirectRoundTripper{targetURL: server.URL},
		},
	}

	res, err := client.RestReleasesCreate("test-site", "test-version")
	if err != nil {
		t.Fatalf("RestReleasesCreate failed: %v", err)
	}

	if res.Name != mockResponse.Name {
		t.Errorf("expected name %s, got %s", mockResponse.Name, res.Name)
	}
}

func TestRestCreateVersionWithMock(t *testing.T) {
	mockResponse := VersionCreateReturn{
		Name:   "sites/test-site/versions/test-version",
		Status: "CREATED",
	}

	var lastReqBody VersionCreateRequestBody

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected method POST, got %s", r.Method)
		}

		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &lastReqBody)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	client := &AuthorizedHTTPClient{
		Client: &http.Client{
			Transport: &redirectRoundTripper{targetURL: server.URL},
		},
	}

	// Test case 1: No firebase.json
	os.Remove("firebase.json")
	config := FirebaseConfigOrDefault("test-target")
	_, err := client.RestCreateVersion("test-site", config)
	if err != nil {
		t.Fatalf("RestCreateVersion failed: %v", err)
	}

	if len(lastReqBody.Config.Headers) != 1 || lastReqBody.Config.Headers[0].Headers["Cache-Control"] != "max-age=1800" {
		t.Errorf("unexpected default config: %+v", lastReqBody.Config)
	}

	// Test case 2: With firebase.json
	firebaseConfig := `{
		"hosting": {
			"headers": [{
				"glob": "**",
				"headers": [{
					"key": "Cache-Control",
					"value": "max-age=3600"
				}]
			}]
		}
	}`
	os.WriteFile("firebase.json", []byte(firebaseConfig), 0644)
	defer os.Remove("firebase.json")

	_, err = client.RestCreateVersion("test-site", FirebaseConfigOrDefault("default"))
	if err != nil {
		t.Fatalf("RestCreateVersion failed: %v", err)
	}

	if len(lastReqBody.Config.Headers) != 1 || lastReqBody.Config.Headers[0].Headers["Cache-Control"] != "max-age=3600" {
		t.Errorf("unexpected config from firebase.json: %+v", lastReqBody.Config)
	}
}

func TestFirebaseConfigOrDefault(t *testing.T) {
	// Re-implementing the previous test in the main test file for consolidation
	tests := []struct {
		name            string
		firebaseJson    string
		target          string
		expectHeaders   int
		expectGlob      string
		expectCacheCtl  string
		expectRedirects int
		expectRedirGlob string
		expectRedirLoc  string
		expectRedirCode int
		skip            bool
	}{
		{
			name: "Single site object",
			firebaseJson: `{
				"hosting": {
					"headers": [{
						"glob": "**",
						"headers": [{
							"key": "Cache-Control",
							"value": "max-age=3600"
						}]
					}]
				}
			}`,
			target:         "default",
			expectHeaders:  1,
			expectGlob:     "**",
			expectCacheCtl: "max-age=3600",
		},
		{
			name: "Multiple sites array",
			firebaseJson: `{
				"hosting": [
					{
						"target": "site-a",
						"headers": [{
							"glob": "/a/**",
							"headers": [{
								"key": "Cache-Control",
								"value": "max-age=100"
							}]
						}]
					},
					{
						"target": "site-b",
						"headers": [{
							"glob": "/b/**",
							"headers": [{
								"key": "Cache-Control",
								"value": "max-age=200"
							}]
						}]
					}
				]
			}`,
			target:         "site-b",
			expectHeaders:  1,
			expectGlob:     "/b/**",
			expectCacheCtl: "max-age=200",
			skip:           false,
		},
		{
			name: "Firebase source field mapping",
			firebaseJson: `{
				"hosting": {
					"headers": [{
						"source": "/src/**",
						"headers": [
							{"key": "X-Custom", "value": "val"}
						]
					}]
				}
			}`,
			target:         "default",
			expectHeaders:  1,
			expectGlob:     "/src/**",
			expectCacheCtl: "",
		},
		{
			name: "Redirects mapping",
			firebaseJson: `{
				"hosting": {
					"redirects": [{
						"source": "/foo/**",
						"destination": "/bar",
						"type": 301
					}]
				}
			}`,
			target:          "default",
			expectHeaders:   1,
			expectGlob:      "**",
			expectCacheCtl:  "max-age=1800",
			expectRedirects: 1,
			expectRedirGlob: "/foo/**",
			expectRedirLoc:  "/bar",
			expectRedirCode: 301,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skip {
				t.Skip("skipped")
			}
			if err := os.WriteFile("firebase.json", []byte(tt.firebaseJson), 0644); err != nil {
				t.Fatalf("failed to write mock firebase.json: %v", err)
			}
			defer os.Remove("firebase.json")

			config := FirebaseConfigOrDefault(tt.target)

			if len(config.Headers) != tt.expectHeaders {
				t.Errorf("expected %d headers, got %d", tt.expectHeaders, len(config.Headers))
			}
			if len(config.Redirects) != tt.expectRedirects {
				t.Errorf("expected %d redirects, got %d", tt.expectRedirects, len(config.Redirects))
			}

			if tt.expectRedirects > 0 {
				r := config.Redirects[0]
				if r.Glob != tt.expectRedirGlob {
					t.Errorf("expected redirect glob %s, got %s", tt.expectRedirGlob, r.Glob)
				}
				if r.Location != tt.expectRedirLoc {
					t.Errorf("expected redirect location %s, got %s", tt.expectRedirLoc, r.Location)
				}
				if r.StatusCode != tt.expectRedirCode {
					t.Errorf("expected redirect code %d, got %d", tt.expectRedirCode, r.StatusCode)
				}
			}

			if tt.expectHeaders > 0 {
				h := config.Headers[0]
				if h.Glob != tt.expectGlob {
					t.Errorf("expected glob %s, got %s", tt.expectGlob, h.Glob)
				}
				if tt.expectCacheCtl != "" {
					if val, ok := h.Headers["Cache-Control"]; !ok || val != tt.expectCacheCtl {
						t.Errorf("expected Cache-Control %s, got %s", tt.expectCacheCtl, val)
					}
				}
				if tt.name == "Firebase source field mapping" {
					if val, ok := h.Headers["X-Custom"]; !ok || val != "val" {
						t.Errorf("expected X-Custom val, got %s", val)
					}
				}
			}
		})
	}
}
