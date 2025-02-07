package main

// Implmentation of the /scan endpoint.

import (
    "fmt"
    "gorm.io/gorm"
    "gorm.io/gorm/clause"
    "log"
    "net/http"
    "strings"
    "sync"
    "time"
)

// Expected request format.
type ScanRequest struct {
    Repo string `json:"repo"`
    Files []string `json:"files"`
}

func HandleScan(db *gorm.DB, client HttpClient) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
        var sr ScanRequest
        err := Decode(req.Body, &sr)
        if err != nil || sr.Repo == "" {
            log.Print(err)
            http.Error(w, 
                "Request must be in the format {'repo': 'URL_OF_REPOSITORY'," +
                " 'files': ['FILE_1', ...]}",
                http.StatusBadRequest)
            return
        }

        repo, err := ExtractGitHubRepo(sr.Repo)
        if err != nil {
            http.Error(w, 
                "URL must be in the format https://github.com/OWNER/REPOSITORY",
                http.StatusBadRequest)
            return
        }

        // HTTP request does not block.  Actual scanning is done in a
        // Goroutine.  A more robust approach might return an id and allow
        // for the caller to check back on the status of the scan. 
        go ScanRepository(repo, sr.Files, db, client)
        fmt.Fprintln(w, "Repository scan underway")
    })
}

// Format of information fetched about the repository.
type RepositoryFile struct {
    Name string `json:"name"`
    DownloadUrl string `json:"download_url"`
}

// If no filenames are passed in, this function will attempt to load all
// ".json" files.  Otherwise, only exact filename matches will be used.
// Loading files is done in parallel, with each file loaded in a separate
// Goroutine.  This function waits for all Goroutines to finish before
// returning.  
func ScanRepository(
        repository string,
        filenames []string,
        db *gorm.DB,
        client HttpClient) {
    log.Print("Scanning " + repository)

    var files []RepositoryFile
    url := "https://api.github.com/repos/" + repository + "/contents"
    err := FetchJson(client, url, &files)
    if err != nil {
        log.Print(err)
        return
    }

    // For the filename check.
    target := map[string]bool{}
    for _, filename := range filenames {
        target[filename] = true
    }

    var wg sync.WaitGroup
    for _, file := range files {
        _, present := target[file.Name]
        if present || 
            (len(filenames) == 0 && strings.HasSuffix(file.Name, ".json")) {
            // Load file in Goroutine.
            wg.Add(1)
            go func() {
                defer wg.Done()
                LoadFile(file, db, client)
            }()
        }
    }

    // Wait for all file loading to complete.
    wg.Wait()
    log.Print("Scanning complete")
}

// Expected format of file contents.  Process and schema could be made more
// robust as needed.
type Entry struct {
    ScanResult Scan `json:"scanResults"`
}

// Fetches file contents and writes to the database.
func LoadFile(rf RepositoryFile, db *gorm.DB, client HttpClient) {
    log.Print("Loading file " + rf.Name)
    var entries []Entry
    err := FetchJson(client, rf.DownloadUrl, &entries)
    if err != nil {
        log.Print(err)
    }

    // Writes each scan to the database.  Vulnerabilites will be created as a
    // part of this process.
    for _, entry := range entries {
        entry.ScanResult.Filename = rf.Name
        entry.ScanResult.ProcessTime = time.Now()
        db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&entry.ScanResult)
    }
    log.Print("Loading file completed")
}

