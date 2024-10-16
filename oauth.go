package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	_ "log"
	"os"
	// "golang.org/x/oauth2"
	// auth "golang.org/x/oauth2/google"
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

func main() {
	// var token *oauth2.Token
	// ctx := context.Background()
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
			if resp, err := client.Get("https://firebasehosting.googleapis.com/v1beta1/projects/tonym-us/sites"); err != nil {
				panic(err)
			} else {
				if body, err := io.ReadAll(resp.Body); err != nil {
					panic(err)
				} else {
					fmt.Printf("%s\n", string(body))
				}
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
}
