package rest

import (
	"time"
)

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

type VersionPopulateFilesRequestBody struct {
	Files map[string]string `json:"files"`
}

type VersionPopulateFilesReturn struct {
	UploadRequiredHashes []string `json:"uploadRequiredHashes"`
	UploadURL            string   `json:"uploadUrl"`
}

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
