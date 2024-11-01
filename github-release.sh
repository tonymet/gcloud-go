#!/bin/zsh
owner=tonymet
repo="gcloud-go"
VERSION=$(date +%Y-%m-%d)-${COMMIT}
build(){
    make build/gcloud-go.tgz
    if [[ $? -ne 0 ]]; then
        echo "ERROR building"
        exit 2
    fi
}

github_list_releases(){
    if [[ -z $GH_TOKEN ]]; then
        echo "GH_TOKEN is unset"
        exit 1
    fi
    TAG=$VERSION
    res=$(\
        curl -s -L -f \
    -X GET \
    -H "Accept: application/vnd.github+json" \
    -H "Authorization: Bearer $GH_TOKEN" \
    -H "X-GitHub-Api-Version: 2022-11-28" \
    "https://api.github.com/repos/${owner}/${repo}/releases?per_page=1" \
    -d "{\"tag_name\":\"${TAG}\",\"target_commitish\":\"${SHA}\",\"name\":\"${TAG}\",\"body\":\"gcloud-go cli release\",\"draft\":false,\"prerelease\":false,\"generate_release_notes\":false}"\
    )
    #echo $res |jq '[0].tag_name'
    echo $res
}

github_latest_tag(){
    echo "checking ${VERSION}"
    latest=$(github_list_releases | jq -r '.[0].tag_name') 
    echo "latest=${latest}"
    if [[ $latest = $VERSION ]]; then
        echo "Release ${VERSION} exists"
        exit 0
    else
        echo "${VERSION} not yet published"
    fi
}

github_release(){
    if [[ -z $GH_TOKEN ]]; then
        echo "GH_TOKEN is unset"
        exit 1
    fi
    TAG=$VERSION
    echo "creating release version=$VERSION"
    res=$(\
        curl -s -L -f \
    -X POST \
    -H "Accept: application/vnd.github+json" \
    -H "Authorization: Bearer $GH_TOKEN" \
    -H "X-GitHub-Api-Version: 2022-11-28" \
    https://api.github.com/repos/${owner}/${repo}/releases \
    -d "{\"tag_name\":\"${TAG}\",\"target_commitish\":\"${SHA}\",\"name\":\"${TAG}\",\"body\":\"gcloud-go cli release\",\"draft\":false,\"prerelease\":false,\"generate_release_notes\":false}"\
    )
    if [[ $? -ne 0 ]]; then
        echo "ERROR: create release fail"
        echo "res=$res"
        exit 1
    fi
    ID=$(echo $res | jq .id)
    echo "uploading asset id=$ID" 
    curl -s -L \
    -X POST \
    -H "Accept: application/vnd.github+json" \
    -H "Authorization: Bearer $GH_TOKEN" \
    -H "X-GitHub-Api-Version: 2022-11-28" \
    -H "Content-Type: application/octet-stream" \
    "https://uploads.github.com/repos/${owner}/${repo}/releases/$ID/assets?name=gcloud-go-$TAG-linux-x86_64.tar.gz" \
    --data-binary "@build/gcloud-go.tgz"
    cd ..
}

build
github_latest_tag
github_release