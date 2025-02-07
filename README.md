# Simple Go Server

## Requirements
Status: complete\
Language: Go\
Database: sqlite\
Concurrency: processes files in parallel\
Error Handling: retries for API calls\
Docker: functional Docker container\
Testing: tests implemented\
Scan endpoint: complete, scans all .json files if no files specified, otherwise looks for file matches\
Query endpoint: complete, supports filtering by additional fields beyond "severity"

## Instructions for Docker and manual testing
Build and run image, mapping ports as needed:
```
docker build . -t simple_go_server
docker run -p 8080:8080 --name simple_go_server simple_go_server
```

Make requests from the command line to test out functionality:
```
curl -X POST -d '{"repo":"https://github.com/OWNER/REPOSITORY", "files":[]}' http://localhost:8080/scan
curl -X POST -d '{"repo":"https://github.com/OWNER/REPOSITORY", "files":["FILENAME"]}' http://localhost:8080/scan
curl -X POST -d '{"filters":{"severity":"HIGH"}}' http://localhost:8080/query
curl -X POST -d '{"filters":{"status":"active"}}' http://localhost:8080/query
```

Stop and remove container when finished:
```
docker stop simple_go_server
docker rm simple_go_server
```

## Instructions for running automated tests
From the src directory, run:
```
go test
```

## Code structure
server.go - starts the server\
database.go - defines the schema and provides database connections\
scan.go - implementation of the /scan endpoint\
query.go - implementation of the /query endpoint\
util.go - utility methods\
Test files correspond to the non-test files.

## Additional improvements
Given additional time, there would be further opportunities to clarify requirements, update the code structure, make the code more robust, and improve the tests.
