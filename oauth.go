package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	fs "main/fs"
	"main/rest"
	"net/http"
	"net/http/httputil"
	"os"
	ppath "path"
	"strings"

	jwt "golang.org/x/oauth2/jwt"
)

var (
	site       = "dev-isgithubipv6"
	version    = "aff8740d2d0aa2dc"
	stagingDir string
)

const (
	STATUS_FINALIZED = "FINALIZED"
	STATUS_CREATED   = "CREATED"
)

type JWTConfig struct {
	Type                    string `json:"type"`
	ProjectID               string `json:"project_id"`
	PrivateKeyID            string `json:"private_key_id"`
	PrivateKey              string `json:"private_key"`
	ClientEmail             string `json:"client_email"`
	ClientID                string `json:"client_id"`
	AuthURI                 string `json:"auth_uri"`
	TokenURI                string `json:"token_uri"`
	AuthProviderX509CertURL string `json:"auth_provider_x509_cert_url"`
	ClientX509CertURL       string `json:"client_x509_cert_url"`
	UniverseDomain          string `json:"universe_domain"`
}

var (
	flagSource string
	flagTemp   string
)

func init() {
	flag.StringVar(&flagSource, "source", "/content", "Source directory for content")
	flag.StringVar(&flagTemp, "temp", os.TempDir(), "temp directory for content")
	flag.Parse()
}

func main() {
	client := authorizeClient(context.Background())
	if body, err := getResource(client, "https://firebasehosting.googleapis.com/v1beta1/projects/tonym-us/sites"); err != nil {
		panic(err)
	} else {
		// fmt.Printf("%s\n", body)
		_ = body
	}
	if body, err := getResource(client, "https://firebasehosting.googleapis.com/v1beta1/projects/tonym-us/sites/"+site+"/versions?filter=status%3D%22CREATED%22"); err != nil {
		panic(err)
	} else {
		// fmt.Printf("%s\n", body)
		_ = body
	}
	// build sha index
	// get tmp dir
	var (
		cwd string
		err error
	)

	if cwd, err = os.Getwd(); err != nil {
		panic(err)
	} else if stagingDir, err := os.MkdirTemp(flagTemp, "firebase-"); err != nil {
		panic(err)
	} else if err := os.Chdir(flagSource); err != nil {
		panic(err)
	} else if ts, err := fs.ShaFiles("./", stagingDir); err != nil {
		panic(err)
	} else if statusVersionCreate, err := restCreateVersion(client, site); err != nil {
		panic(err)
	} else if statusVersionCreate.Status != STATUS_CREATED {
		panic("status not created")
	} else if popFiles, err := restCreateVersionPopulateFiles(client, ts, statusVersionCreate.Name); err != nil {
		panic(err)
	} else if err := restUploadFileList(client, statusVersionCreate.Name, popFiles); err != nil {
		panic(err)
	} else if statusReturn, err := restVersionSetStatus(client, statusVersionCreate.Name, STATUS_FINALIZED); err != nil {
		panic(err)
	} else if statusRelease, err := restReleasesCreate(client, site, statusVersionCreate.Name); err != nil {
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

func authorizeClient(ctx context.Context) *http.Client {
	var conf JWTConfig
	if f, err := os.Open("tonym-us-311af670bc42.json"); err != nil {
		panic(err)
	} else {
		cont := bufio.NewReader(f)
		if bytes, err := io.ReadAll(cont); err != nil {
			panic(err)
		} else {
			json.Unmarshal(bytes, &conf)
			// fmt.Printf("%+v\n", conf)
		}
		var jConf jwt.Config
		jConf.Scopes = []string{
			"https://www.googleapis.com/auth/cloud-platform",
			"https://www.googleapis.com/auth/cloud-platform.read-only",
			"https://www.googleapis.com/auth/firebase",
			"https://www.googleapis.com/auth/firebase.readonly",
		}
		jConf.TokenURL = conf.TokenURI
		jConf.Email = conf.ClientEmail
		jConf.PrivateKeyID = conf.PrivateKeyID
		jConf.PrivateKey = []byte(conf.PrivateKey)
		jConf.Subject = conf.ClientEmail
		return jConf.Client(ctx)
	}
}
func restDebugRequest(req *http.Request) {
	reqDump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("REQUEST:\n%s", string(reqDump))
}

func restUploadFileList(client *http.Client, versionId string, manifest rest.VersionPopulateFilesReturn) error {
	for _, shaHash := range manifest.UploadRequiredHashes {
		// open file reader
		if f, err := os.Open(ppath.Join(stagingDir, shaHash)); err != nil {
			panic(err)
		} else if err := restUploadFile(client, f, shaHash, versionId); err != nil {
			panic(err)
		}
	}
	return nil
}

//	func restVersionSetStatus(client *http.Client, site, version, status string) (r rest.VersionStatusUpdateReturn, e error) {
//		resource := "https://firebasehosting.googleapis.com/v1beta1/sites/" + site + "/versions/" + version + "?update_mask=status"
func restVersionSetStatus(client *http.Client, versionId string, status string) (r rest.VersionStatusUpdateReturn, e error) {
	resource := "https://firebasehosting.googleapis.com/v1beta1/" + versionId + "?update_mask=status"
	// set up shas
	var bodyJson rest.VersionStatusUpdateRequestBody
	bodyJson.Status = status
	if bodyBuffer, err := json.Marshal(bodyJson); err != nil {
		panic(err)
	} else if req, err := http.NewRequest(http.MethodPatch, resource, bytes.NewReader(bodyBuffer)); err != nil {
		panic(err)
	} else {
		req.Header.Add("Content-Type", "application/json")
		if res, err := client.Do(req); err != nil {
			panic(err)
		} else if res.StatusCode < 200 || res.StatusCode > 299 {
			panic(fmt.Sprintf("http error: status = %s, resource = %s\n", res.Status, res.Request.URL))
		} else if bodyBytes, err := io.ReadAll(res.Body); err != nil {
			panic(err)
		} else if err := json.Unmarshal(bodyBytes, &r); err != nil {
			panic(err)
		} else {
			return r, nil
		}
	}
}

// func restReleasesCreate(client *http.Client, site, version string) (r rest.ReleasesCreateReturn, e error) {
// resource := "https://firebasehosting.googleapis.com/v1beta1/sites/" + site + "/releases?versionName=sites/" + site + "/versions/" + version
func restReleasesCreate(client *http.Client, site, versionId string) (r rest.ReleasesCreateReturn, e error) {
	resource := "https://firebasehosting.googleapis.com/v1beta1/sites/" + site + "/releases?versionName=" + versionId
	// set up shas
	if req, err := http.NewRequest(http.MethodPost, resource, nil); err != nil {
		panic(err)
	} else if res, err := client.Do(req); err != nil {
		panic(err)
	} else if res.StatusCode < 200 || res.StatusCode > 299 {
		panic(fmt.Sprintf("http error: status = %s, resource = %s\n", res.Status, res.Request.URL))
	} else if bodyBytes, err := io.ReadAll(res.Body); err != nil {
		panic(err)
	} else if err := json.Unmarshal(bodyBytes, &r); err != nil {
		panic(err)
	} else {
		return r, nil
	}
}

// func restUploadFile(client *http.Client, bodyFile io.Reader, shaHash, site, version string) error {
// resource := "https://upload-firebasehosting.googleapis.com/upload/sites/" + site + "/versions/" + version + "/files/" + shaHash
func restUploadFile(client *http.Client, bodyFile io.Reader, shaHash, versionId string) error {
	resource := "https://upload-firebasehosting.googleapis.com/upload/" + versionId + "/files/" + shaHash
	// set up shas
	if req, err := http.NewRequest(http.MethodPost, resource, bodyFile); err != nil {
		panic(err)
	} else {
		req.Header.Add("Content-Type", "application/octet-stream")
		if res, err := client.Do(req); err != nil {
			panic(err)
		} else if res.StatusCode < 200 || res.StatusCode > 299 {
			panic(fmt.Sprintf("http error: status = %s, resource = %s\n", res.Status, res.Request.URL))
		}
	}
	return nil
}
func restCreateVersionPopulateFiles(client *http.Client, shas fs.ShaList, versionId string) (r rest.VersionPopulateFilesReturn, err error) {
	resource := "https://firebasehosting.googleapis.com/v1beta1/" + versionId + ":populateFiles"
	// set up shas
	var bodyJson rest.VersionPopulateFilesRequestBody
	bodyJson.Files = make(map[string]string)
	for _, s := range shas {
		bodyJson.Files[s.RelPath] = s.Shasum
	}
	if bodyBuffer, err := json.Marshal(bodyJson); err != nil {
		panic(err)
	} else {
		if req, err := http.NewRequest(http.MethodPost, resource, bytes.NewReader(bodyBuffer)); err != nil {
			panic(err)
		} else {
			restDebugRequest(req)
			req.Header.Add("Content-Type", "application/json")
			if res, err := client.Do(req); err != nil {
				panic(err)
			} else if res.StatusCode < 200 || res.StatusCode > 299 {
				panic(fmt.Sprintf("http error: status = %s, resource = %s\n", res.Status, res.Request.URL))
			} else if bodyBytes, err := io.ReadAll(res.Body); err != nil {
				panic(err)
			} else {
				if err := json.Unmarshal(bodyBytes, &r); err != nil {
					panic(err)
				} else {
					return r, nil
				}
			}
		}

	}
}

func restCreateVersion(client *http.Client, site string) (r rest.VersionCreateReturn, e error) {
	reqBody := ` 
	{
             "config": {
               "headers": [{
                 "glob": "**",
                 "headers": {
                   "Cache-Control": "max-age=1800"
                 }
               }]
             }
           }
	`
	resource := "https://firebasehosting.googleapis.com/v1beta1/projects/tonym-us/sites/" + site + "/versions"
	if req, err := http.NewRequest("POST", resource, strings.NewReader(reqBody)); err != nil {
		panic(err)
	} else {
		if resp, err := client.Do(req); err != nil {
			return rest.VersionCreateReturn{}, err
		} else {
			if body, err := io.ReadAll(resp.Body); err != nil {
				return rest.VersionCreateReturn{}, err
			} else {
				if err := json.Unmarshal(body, &r); err != nil {
					panic(err)
				}
				return r, nil
			}
		}
	}
}

func getResource(client *http.Client, resource string) (string, error) {
	if resp, err := client.Get(resource); err != nil {
		return "", err
	} else {
		if body, err := io.ReadAll(resp.Body); err != nil {
			return "", err
		} else {
			return string(body), nil
		}
	}
}

//credentials, err := auth.FindDefaultCredentials(ctx, scopes...)
// if err == nil {
// 	log.Printf("found default credentials. %v", credentials)
// 	token, err = credentials.TokenSource.Token()
// 	log.Printf("token: %v, err: %v", token, err)
// 	if err != nil {
// 		log.Print(err)
// 	}
// }
