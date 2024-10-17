package rest

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	fs "main/fs"
	"net/http"
	"net/http/httputil"
	"os"
	ppath "path"
	"strings"

	jwt "golang.org/x/oauth2/jwt"
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

func AuthorizeClient(ctx context.Context, flagCred string) (*http.Client, error) {
	var conf JWTConfig
	if f, err := os.Open(flagCred); err != nil {
		panic(err)
	} else {
		cont := bufio.NewReader(f)
		if bytes, err := io.ReadAll(cont); err != nil {
			panic(err)
		} else {
			json.Unmarshal(bytes, &conf)
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
		return jConf.Client(ctx), nil
	}
}

/**
func authorizeClientIdToken(ctx context.Context) (*http.Client, error) {
	// client is a http.Client that automatically adds an "Authorization" header
	// to any requests made.
	client, err := idtoken.NewClient(ctx, "https://tonym.us/")
	if err != nil {
		return nil, fmt.Errorf("idtoken.NewClient: %w", err)
	}
	return client, nil

}
**/

func RestDebugRequest(req *http.Request) {
	reqDump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("REQUEST:\n%s", string(reqDump))
}

func RestUploadFileList(client *http.Client, versionId string, manifest VersionPopulateFilesReturn, stagingDir string) error {
	for _, shaHash := range manifest.UploadRequiredHashes {
		// open file reader
		if f, err := os.Open(ppath.Join(stagingDir, shaHash)); err != nil {
			panic(err)
		} else if err := RestUploadFile(client, f, shaHash, versionId); err != nil {
			panic(err)
		}
	}
	return nil
}

//	func restVersionSetStatus(client *http.Client, site, version, status string) (r rest.VersionStatusUpdateReturn, e error) {
//		resource := "https://firebasehosting.googleapis.com/v1beta1/sites/" + site + "/versions/" + version + "?update_mask=status"
func RestVersionSetStatus(client *http.Client, versionId string, status string) (r VersionStatusUpdateReturn, e error) {
	resource := "https://firebasehosting.googleapis.com/v1beta1/" + versionId + "?update_mask=status"
	// set up shas
	var bodyJson VersionStatusUpdateRequestBody
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
func RestReleasesCreate(client *http.Client, site, versionId string) (r ReleasesCreateReturn, e error) {
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
func RestUploadFile(client *http.Client, bodyFile io.Reader, shaHash, versionId string) error {
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
func RestCreateVersionPopulateFiles(client *http.Client, shas fs.ShaList, versionId string) (r VersionPopulateFilesReturn, err error) {
	resource := "https://firebasehosting.googleapis.com/v1beta1/" + versionId + ":populateFiles"
	// set up shas
	var bodyJson VersionPopulateFilesRequestBody
	bodyJson.Files = make(map[string]string)
	for _, s := range shas {
		bodyJson.Files[s.RelPath] = s.Shasum
	}
	if bodyBuffer, err := json.Marshal(bodyJson); err != nil {
		panic(err)
	} else if req, err := http.NewRequest(http.MethodPost, resource, bytes.NewReader(bodyBuffer)); err != nil {
		panic(err)
	} else {
		RestDebugRequest(req)
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

func RestCreateVersion(client *http.Client, site string) (r VersionCreateReturn, e error) {
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
	} else if resp, err := client.Do(req); err != nil {
		return VersionCreateReturn{}, err
	} else if body, err := io.ReadAll(resp.Body); err != nil {
		return VersionCreateReturn{}, err
	} else if err := json.Unmarshal(body, &r); err != nil {
		panic(err)
	}
	return r, nil
}

func GetResource(client *http.Client, resource string) (string, error) {
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
