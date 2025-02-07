package main

// Starts the server.

import (
    "net/http"
    "github.com/hashicorp/go-retryablehttp"
    "log"
)

// Configurations
const databaseFile string = "database.db"
const retries int = 2

// Custom client can be used for testing.
type HttpClient interface {
    Do(req *http.Request) (*http.Response, error)
}

func main() {
    db := OpenAndMigrateDb(databaseFile)

    // This client will retry API calls.
    retryClient := retryablehttp.NewClient()
    retryClient.RetryMax = retries
    standardClient := retryClient.StandardClient()

    http.Handle("/scan", HandleScan(db, standardClient))
    http.Handle("/query", HandleQuery(db))
    log.Fatal(http.ListenAndServe(":8080", nil))
}

