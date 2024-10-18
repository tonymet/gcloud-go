package main

import (
	"context"
	"flag"
	"fmt"
	fs "main/fs"
	"main/rest"
	"os"
)

const (
	STATUS_FINALIZED = "FINALIZED"
	STATUS_CREATED   = "CREATED"
	OP_DEPLOY        = "deploy"
)

var (
	flagSource, flagTemp, flagCred, flagSite *string
	cmdDeploy                                *flag.FlagSet
)

func init() {
	cmdDeploy = flag.NewFlagSet("deploy", flag.ExitOnError)
	flagTemp = cmdDeploy.String("temp", os.TempDir(), "temp directory for staging files prior to upload")
	flagCred = cmdDeploy.String("cred", os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"),
		"path to service principal. Use ENV var GOOGLE_APPLICATION_CREDENTAILS by default. "+
			"Within GCP, metadata server will be used")
	rest.FlagConn = cmdDeploy.Int("connections", 8, "number of connections")
	flagSource = cmdDeploy.String("source", "content", "Source directory for content")
	flagSite = cmdDeploy.String("site", "default", "Name of site (not project)")
}
func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s deploy [options]\n options:\n", os.Args[0])
	cmdDeploy.PrintDefaults()
}
func main() {
	if len(os.Args) < 2 {
		usage()
		return
	}
	switch os.Args[1] {
	case "deploy":
		cmdDeploy.Parse(os.Args[2:])
		if client, err := rest.AuthorizeClientDefault(context.Background(), *flagCred); err != nil {
			panic(err)
		} else if cwd, err := os.Getwd(); err != nil {
			panic(err)
		} else if stagingDir, err := os.MkdirTemp(*flagTemp, "firebase-"); err != nil {
			panic(err)
		} else if err := os.Chdir(*flagSource); err != nil {
			panic(err)
		} else if ts, err := fs.ShaFiles("./", stagingDir); err != nil {
			panic(err)
		} else if statusVersionCreate, err := rest.RestCreateVersion(client, *flagSite); err != nil {
			panic(err)
		} else if statusVersionCreate.Status != STATUS_CREATED {
			panic("status not created")
		} else if popFiles, err := rest.RestCreateVersionPopulateFiles(client, ts, statusVersionCreate.Name); err != nil {
			panic(err)
		} else if err := rest.RestUploadFileList(client, statusVersionCreate.Name, popFiles, stagingDir); err != nil {
			panic(err)
		} else if statusReturn, err := rest.RestVersionSetStatus(client, statusVersionCreate.Name, STATUS_FINALIZED); err != nil {
			panic(err)
		} else if statusRelease, err := rest.RestReleasesCreate(client, *flagSite, statusVersionCreate.Name); err != nil {
			panic(err)
		} else if err := os.Chdir(cwd); err != nil {
			panic(err)
		} else {
			_ = statusReturn
			_ = statusRelease
			_ = statusVersionCreate
		}
	default:
		usage()
	}
}
