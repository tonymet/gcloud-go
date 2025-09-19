package misc

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/storage"
	"github.com/tonymet/gcloud-go/github"
	"github.com/tonymet/gcloud-go/kms"
	_ "golang.org/x/crypto/x509roots/fallback"
	"google.golang.org/api/iterator"
)

var (
	ErrReleaseExists = fmt.Errorf("release already exists")
)

func init() {
	mime.AddExtensionType(".sig", "application/octet-stream")         //nolint:errcheck
	mime.AddExtensionType(".gz", "application/x-gtar-compressed")     //nolint:errcheck
	mime.AddExtensionType(".tar.gz", "application/x-gtar-compressed") //nolint:errcheck
}

func logErr(format string, params ...any) {
	if len(params) > 0 {
		fmt.Fprintf(os.Stderr, format, params...)
	} else {
		fmt.Fprint(os.Stderr, format)
	}
}

type BuildCommand struct {
	Command string `json:"command"`
	Version string `json:"cloud_sdk_version"`
}

func (bc BuildCommand) toJson() ([]byte, error) {
	if b, err := json.Marshal(bc); err != nil {
		return []byte{}, err
	} else {
		return b, nil
	}
}

// uploadFile uploads an object.
func SetObject(bucket, object, contents string) error {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		panic(err)
	}
	defer client.Close() //nolint:errcheck
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	f := strings.NewReader(contents)
	o := client.Bucket(bucket).Object(object)
	//o = o.If(storage.Conditions{DoesNotExist: true})
	wc := o.NewWriter(ctx)
	if _, err = io.Copy(wc, f); err != nil {
		panic(err)
	}
	if err := wc.Close(); err != nil {
		panic(err)
	}
	return nil
}

func SyncDown(bucket, prefix string) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	client, err := storage.NewClient(ctx)
	if err != nil {
		panic(err)
	}
	defer client.Close() //nolint:errcheck
	bkt := client.Bucket(bucket)
	query := &storage.Query{Prefix: prefix}
	it := bkt.Objects(ctx, query)
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		logErr("name: %s\n", attrs.Name)
		switch {
		case strings.HasSuffix(attrs.Name, "/"):
			if err := os.MkdirAll(attrs.Name, 0750); err != nil {
				panic(err)
			}
		default:
			if h, err := bkt.Object(attrs.Name).NewReader(ctx); err != nil {
				panic(err)
			} else if f, err := os.Create(attrs.Name); err != nil {
				panic(err)
			} else if _, err := io.Copy(f, h); err != nil {
				panic(err)
			}
		}
	}
}
func GetObjectStdout(bucket, object string) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	obj := GetObject(ctx, bucket, object)
	if r, err := obj.NewReader(ctx); err != nil {
		panic(err)
	} else if _, err := io.Copy(os.Stdout, r); err != nil {
		panic(err)
	}
}

func GetObjectContents(bucket, object string) string {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	obj := GetObject(ctx, bucket, object)
	r, err := obj.NewReader(ctx)
	if err != nil {
		panic(err)
	}
	var result strings.Builder
	_, err = io.Copy(&result, r)
	if err != nil {
		panic(err)
	}
	return result.String()
}

func GetObject(ctx context.Context, bucket, object string) *storage.ObjectHandle {
	client, err := storage.NewClient(ctx)
	if err != nil {
		panic(err)
	}
	defer client.Close() //nolint:errcheck
	bkt := client.Bucket(bucket)
	return bkt.Object(object)
}

func PubSubPush(project, topic string, object any) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	c, err := pubsub.NewClient(ctx, project)
	if err != nil {
		panic(err)
	}
	t := c.Topic(topic)
	defer t.Stop()
	if marshalled, err := json.Marshal(object); err != nil {
		return err
	} else {
		pr := t.Publish(ctx, &pubsub.Message{Data: marshalled})
		if _, err := pr.Get(ctx); err != nil {
			return err
		}
	}
	return nil
}

func IncrementVersion(v string) string {
	major := strings.Split(v, ".")
	if val, err := strconv.ParseInt(major[0], 10, 32); err != nil {
		panic(err)
	} else {
		return fmt.Sprintf("%d.0.0", val+1)
	}
}

func GetActiveVersion(bucket, object string) string {
	return IncrementVersion(GetObjectContents(bucket, object))
}

func GithubRelease(args github.GithubArgs) error {
	owner, repo, file, commit := args.Owner, args.Repo, args.File, args.Commit
	var tagValue string
	if args.Tag != "" {
		tagValue = args.Tag
	} else {
		tagValue = time.Now().Format(time.DateOnly) + "-" + commit[0:7]
	}
	ctx := context.Background()
	r := github.CreateReleaseResponse{
		TagName:         tagValue,
		TargetCommitish: commit,
	}
	gc := github.AuthorizeClient(args.Token)
	if _, res, err := gc.GetReleaseByTag(owner, repo, tagValue); res != nil && res.StatusCode == 200 {
		return ErrReleaseExists
	} else if (res != nil && res.StatusCode != 404) || (err != nil && res == nil) {
		panic(err)
	}
	if repoObj, _, err := gc.GithubCreateRelease(owner, repo, r); err != nil {
		panic(err)
	} else if fileHandle, err := os.Open(file); err != nil {
		panic(err)
	} else {
		fInfo, err := fileHandle.Stat()
		if err != nil {
			panic(err)
		}
		digest := sha256.New()
		tReader := io.TeeReader(fileHandle, digest)
		filename := path.Base(file)
		ext := path.Ext(filename)
		asset, _, err := gc.UploadReleaseAsset(owner, repo, repoObj.ID, filename, tReader, fInfo.Size(), mime.TypeByExtension(ext))
		if err != nil {
			panic(err)
		}
		fileHandle.Close() //nolint:errcheck
		logErr("release ID: %+d\n", repoObj.ID)
		logErr("asset ID: %+x\n", asset.ID)
		if args.KeyPath != "" {
			var outWriter bytes.Buffer
			err = kms.SignAsymmetricDigest(ctx, &outWriter, args.KeyPath, digest)
			if err != nil {
				return err
			}
			assetSig, _, err := gc.UploadReleaseAsset(owner, repo, repoObj.ID, filename+".sig", &outWriter, int64(outWriter.Len()), mime.TypeByExtension(".sig"))
			if err != nil {
				return err
			}
			logErr("sig asset ID: %+x\n", assetSig.ID)
			// generate sig and write to file
		}
	}
	return nil
}

type KMSArgs struct {
	Filename, Output, Keypath string
}
