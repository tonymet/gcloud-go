package main

import (
	"context"
	"flag"
	"fmt"
	"main/rest"
	"os"

	_ "golang.org/x/crypto/x509roots/fallback"
)

const (
	STATUS_FINALIZED = "FINALIZED"
	STATUS_CREATED   = "CREATED"
	OP_DEPLOY        = "deploy"
)

var (
	flagSource, flagTemp, flagSite     *string
	flagBucket, flagPrefix, flagTarget *string
	cmdDeploy, cmdStorage              *flag.FlagSet
)

func init() {
	cmdDeploy = flag.NewFlagSet("deploy", flag.ExitOnError)
	flagTemp = cmdDeploy.String("temp", os.TempDir(), "temp directory for staging files prior to upload")
	rest.FlagConn = cmdDeploy.Int("connections", 8, "number of connections")
	flagSource = cmdDeploy.String("source", "content", "Source directory for content")
	flagSite = cmdDeploy.String("site", "default", "Name of site (not project)")
	cmdStorage = flag.NewFlagSet("storage", flag.ExitOnError)
	flagBucket = cmdStorage.String("bucket", "", "GCS Bucket")
	flagPrefix = cmdStorage.String("prefix", "/", "GCS Object Prefix")
	flagTarget = cmdStorage.String("target", ".", "Target Directory for download")
}
func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s deploy|storage [options]\n options:\n", os.Args[0])
	cmdDeploy.PrintDefaults()
	fmt.Fprintf(os.Stderr, "storage:\n")
	cmdStorage.PrintDefaults()
}
func main() {
	if len(os.Args) < 2 {
		usage()
		return
	}
	switch os.Args[1] {
	case "deploy":
		if err := cmdDeploy.Parse(os.Args[2:]); err != nil {
			panic(err)
		} else if client, _, err := rest.AuthorizeClientDefault(context.Background()); err != nil {
			panic(err)
		} else if cwd, err := os.Getwd(); err != nil {
			panic(err)
		} else if stagingDir, err := os.MkdirTemp(*flagTemp, "firebase-"); err != nil {
			panic(err)
		} else if err := os.Chdir(*flagSource); err != nil {
			panic(err)
		} else if statusVersionCreate, err := rest.RestCreateVersion(client, *flagSite); err != nil {
			panic(err)
		} else if statusVersionCreate.Status != STATUS_CREATED {
			panic("status not created")
		} else if popFiles, err := rest.RestCreateVersionPopulateFiles(client, stagingDir, statusVersionCreate.Name); err != nil {
			panic(err)
		} else if err := rest.RestUploadFileList(client, statusVersionCreate.Name, popFiles, stagingDir); len(err) != 0 {
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
	case "storage":
		if err := cmdStorage.Parse(os.Args[2:]); err != nil {
			panic(err)
		} else if cmdStorage.NFlag() != 3 {
			usage()
			os.Exit(2)
		}
		if _, credsPackage, err := rest.AuthorizeClientDefault(context.Background()); err != nil {
			panic(err)
		} else if err := rest.StorageDownload(credsPackage.GoogleCredentials, *flagBucket, *flagPrefix, *flagTarget, rest.StorageFilterImages); err != nil {
			panic(err)
		}
	default:
		usage()
	}
}
