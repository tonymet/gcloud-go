package main

import (
	"context"
	"flag"
	"fmt"
	fs "main/fs"
	"main/rest"
	"os"
	"strings"
)

var (
	site       = "dev-isgithubipv6"
	stagingDir string
)

const (
	STATUS_FINALIZED = "FINALIZED"
	STATUS_CREATED   = "CREATED"
)

var (
	flagSource string
	flagTemp   string
	flagCred   string
)

func init() {
	flag.StringVar(&flagSource, "source", "/content", "Source directory for content")
	flag.StringVar(&flagTemp, "temp", os.TempDir(), "temp directory for staging files prior to upload")
	flag.StringVar(&flagCred, "cred", os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"), "path to service principal")
	flag.Parse()
}

func main() {
	var (
		cwd string
	)
	if client, err := rest.AuthorizeClient(context.Background(), flagCred); err != nil {
		panic(err)
	} else if cwd, err = os.Getwd(); err != nil {
		panic(err)
	} else if stagingDir, err := os.MkdirTemp(flagTemp, "firebase-"); err != nil {
		panic(err)
	} else if err := os.Chdir(flagSource); err != nil {
		panic(err)
	} else if ts, err := fs.ShaFiles("./", stagingDir); err != nil {
		panic(err)
	} else if statusVersionCreate, err := rest.RestCreateVersion(client, site); err != nil {
		panic(err)
	} else if statusVersionCreate.Status != STATUS_CREATED {
		panic("status not created")
	} else if popFiles, err := rest.RestCreateVersionPopulateFiles(client, ts, statusVersionCreate.Name); err != nil {
		panic(err)
	} else if err := rest.RestUploadFileList(client, statusVersionCreate.Name, popFiles, stagingDir); err != nil {
		panic(err)
	} else if statusReturn, err := rest.RestVersionSetStatus(client, statusVersionCreate.Name, STATUS_FINALIZED); err != nil {
		panic(err)
	} else if statusRelease, err := rest.RestReleasesCreate(client, site, statusVersionCreate.Name); err != nil {
		panic(err)
	} else if err := os.Chdir(cwd); err != nil {
		panic(err)
	} else {
		_ = statusReturn
		_ = statusRelease
		_ = statusVersionCreate
	}
}

func siteVersionId(name string) (site string, version string, err error) {
	if s := strings.Split(name, "/"); len(s) < 4 {
		return "", "", fmt.Errorf("siteVersion parse error")
	} else {
		return s[1], s[3], nil
	}
}
