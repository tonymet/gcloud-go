package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"main/rest"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"

	// "golang.org/x/oauth2"
	// auth "golang.org/x/oauth2/google"
	jwt "golang.org/x/oauth2/jwt"
)

var site = "dev-isgithubipv6"

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

func main() {
	if f, err := os.Open("tonym-us-311af670bc42.json"); err != nil {
		panic(err)
	} else {
		cont := bufio.NewReader(f)
		if bytes, err := io.ReadAll(cont); err != nil {
			panic(err)
		} else {
			ctx := context.Background()
			var conf JWTConfig
			json.Unmarshal(bytes, &conf)
			fmt.Printf("%+v\n", conf)
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
			client := jConf.Client(ctx)
			if body, err := getResource(client, "https://firebasehosting.googleapis.com/v1beta1/projects/tonym-us/sites"); err != nil {
				panic(err)
			} else {
				fmt.Printf("%s\n", body)
			}
			if body, err := getResource(client, "https://firebasehosting.googleapis.com/v1beta1/projects/tonym-us/sites/"+site+"/versions"); err != nil {
				panic(err)
			} else {
				fmt.Printf("%s\n", body)
			}
			if body, err := restCreateVersion(client, site); err != nil {
				panic(err)
			} else {
				fmt.Printf("%+v\n", body)
			}
		}
	}
}

func restDebugRequest(req *http.Request) {
	reqDump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("REQUEST:\n%s", string(reqDump))
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
