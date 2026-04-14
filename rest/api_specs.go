// JSON structs for Firebase Hosting API responses
package rest

import (
	"encoding/json"
	"time"
)

type FirebaseHeaderKV struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type FirebaseHeader struct {
	Source  string             `json:"source"`
	Glob    string             `json:"glob"`
	Headers []FirebaseHeaderKV `json:"headers"`
}

type FirebaseRedirect struct {
	Source      string `json:"source"`
	Glob        string `json:"glob"`
	Regex       string `json:"regex"`
	Destination string `json:"destination"`
	Location    string `json:"location"`
	Type        int    `json:"type"`
	StatusCode  int    `json:"statusCode"`
}

type HostingConfig struct {
	Site           string             `json:"site"`
	Target         string             `json:"target"`
	Public         string             `json:"public"`
	Ignore         []string           `json:"ignore"`
	Redirects      []FirebaseRedirect `json:"redirects"`
	Rewrites       []json.RawMessage  `json:"rewrites"`
	Headers        []FirebaseHeader  `json:"headers"`
	CleanUrls      bool              `json:"cleanUrls"`
	TrailingSlash  bool              `json:"trailingSlash"`
	AppAssociation string            `json:"appAssociation"`
}

type FirebaseConfig struct {
	Hosting json.RawMessage `json:"hosting"`
}

type Header struct {
	Glob    string            `json:"glob"`
	Headers map[string]string `json:"headers"`
}

type Redirect struct {
	Glob       string `json:"glob,omitempty"`
	Regex      string `json:"regex,omitempty"`
	StatusCode int    `json:"statusCode"`
	Location   string `json:"location"`
}

type ServingConfig struct {
	Headers   []Header   `json:"headers,omitempty"`
	Redirects []Redirect `json:"redirects,omitempty"`
}

// create version call return
type VersionCreateReturn struct {
	Name   string        `json:"name"`
	Status string        `json:"status"`
	Config ServingConfig `json:"config"`
}

type VersionCreateRequestBody struct {
	Config ServingConfig `json:"config"`
}

// Populate Files request
type VersionPopulateFilesRequestBody struct {
	Files map[string]string `json:"files"`
}

// Populate Files Response
type VersionPopulateFilesReturn struct {
	UploadRequiredHashes []string `json:"uploadRequiredHashes"`
	UploadURL            string   `json:"uploadUrl"`
}

// Version Status Update Request
type VersionStatusUpdateRequestBody struct {
	Status string `json:"status"`
}
type VersionStatusUpdateReturn struct {
	Name       string        `json:"name"`
	Status     string        `json:"status"`
	Config     ServingConfig `json:"config"`
	CreateTime time.Time     `json:"createTime"`
	CreateUser struct {
		Email string `json:"email"`
	} `json:"createUser"`
	FinalizeTime time.Time `json:"finalizeTime"`
	FinalizeUser struct {
		Email string `json:"email"`
	} `json:"finalizeUser"`
	FileCount    string `json:"fileCount"`
	VersionBytes string `json:"versionBytes"`
}

type ReleasesCreateReturn struct {
	Name    string `json:"name"`
	Version struct {
		Name   string        `json:"name"`
		Status string        `json:"status"`
		Config ServingConfig `json:"config"`
	} `json:"version"`
	Type        string    `json:"type"`
	ReleaseTime time.Time `json:"releaseTime"`
}

type ResponseMessage struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error"`
}
