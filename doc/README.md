## TODO
* p1 X logging
* p1 X state var cleanup e.g. stagingDir
* p1 X nested dirs in content directory
* p2 paging (over 1000 file support)
* p2 progress bar using channel from file upload
* p3 X multi-core sha file upload 


## Authentication Overview
Google uses internal and external (development) authorization schemes to get an auth token. 
outside the cloud (development) , GOOGLE_DEFAULT_CREDENTIALS is used with path to service credential json file
Inside the cloud, metadata server is used. 

go api credentials.DetectDefault will test for the appropriate scheme. 

* [general auth topics](https://cloud.google.com/docs/authentication#service-accounts)
* [rest / metadata authentication](https://cloud.google.com/docs/authentication/rest#metadata-server)

