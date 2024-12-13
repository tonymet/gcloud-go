// JSON structs for Firebase Hosting API responses
package rest

import (
	"time"
)

// create version call
type VersionCreateReturn struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Config struct {
		Headers []struct {
			Glob    string `json:"glob"`
			Headers struct {
				CacheControl string `json:"Cache-Control"`
			} `json:"headers"`
		} `json:"headers"`
	} `json:"config"`
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
	Name   string `json:"name"`
	Status string `json:"status"`
	Config struct {
		Headers []struct {
			Glob    string `json:"glob"`
			Headers struct {
				CacheControl string `json:"Cache-Control"`
			} `json:"headers"`
		} `json:"headers"`
	} `json:"config"`
	CreateTime time.Time `json:"createTime"`
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
		Name   string `json:"name"`
		Status string `json:"status"`
		Config struct {
			Headers []struct {
				Glob    string `json:"glob"`
				Headers struct {
					CacheControl string `json:"Cache-Control"`
				} `json:"headers"`
			} `json:"headers"`
		} `json:"config"`
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
