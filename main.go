package main

import (
	"context"
	"flag"
	fs "main/fs"
	"main/rest"
	"os"
)

const (
	STATUS_FINALIZED = "FINALIZED"
	STATUS_CREATED   = "CREATED"
)

var (
	flagSource, flagTemp, flagCred, flagSite string
)

func init() {
	flag.StringVar(&flagSource, "source", "/content", "Source directory for content")
	flag.StringVar(&flagSite, "site", "default", "Name of site (not project)")
	flag.StringVar(&flagTemp, "temp", os.TempDir(), "temp directory for staging files prior to upload")
	flag.StringVar(&flagCred, "cred", os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"), "path to service principal")
	flag.IntVar(&rest.FlagConn, "connections", 8, "number of connections")
	flag.Parse()
}

func main() {
	var (
		cwd string
	)
	if client, err := rest.AuthorizeClientDefault(context.Background(), flagCred); err != nil {
		panic(err)
	} else if cwd, err = os.Getwd(); err != nil {
		panic(err)
	} else if stagingDir, err := os.MkdirTemp(flagTemp, "firebase-"); err != nil {
		panic(err)
	} else if err := os.Chdir(flagSource); err != nil {
		panic(err)
	} else if ts, err := fs.ShaFiles("./", stagingDir); err != nil {
		panic(err)
	} else if statusVersionCreate, err := rest.RestCreateVersion(client, flagSite); err != nil {
		panic(err)
	} else if statusVersionCreate.Status != STATUS_CREATED {
		panic("status not created")
	} else if popFiles, err := rest.RestCreateVersionPopulateFiles(client, ts, statusVersionCreate.Name); err != nil {
		panic(err)
	} else if err := rest.RestUploadFileList(client, statusVersionCreate.Name, popFiles, stagingDir); err != nil {
		panic(err)
	} else if statusReturn, err := rest.RestVersionSetStatus(client, statusVersionCreate.Name, STATUS_FINALIZED); err != nil {
		panic(err)
	} else if statusRelease, err := rest.RestReleasesCreate(client, flagSite, statusVersionCreate.Name); err != nil {
		panic(err)
	} else if err := os.Chdir(cwd); err != nil {
		panic(err)
	} else {
		_ = statusReturn
		_ = statusRelease
		_ = statusVersionCreate
	}
}
