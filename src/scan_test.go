package main

// Tests for the /scan endpoint and associated functionality.

import (
    "github.com/stretchr/testify/assert"
    "gorm.io/gorm"
    "io"
    "net/http"
    "net/http/httptest"
    "testing"
    "strings"
)

type TestClient struct {}

func (m *TestClient) Do(req *http.Request) (*http.Response, error) {
    if strings.HasSuffix(req.URL.Path, "contents") {
        // call from ScanRepository to get repository contents
        return &http.Response{
            StatusCode: http.StatusOK,
            Body: io.NopCloser(strings.NewReader(
                `[{
                    "name": "example.json",
                    "download_url": "http://github.com/owner/repository/example.json"
                }]`)),
        }, nil
    } else {
        // call from LoadFile to get file contents
        return &http.Response{
            StatusCode: http.StatusOK,
            Body: io.NopCloser(strings.NewReader(
                `[{
                    "scanResults": {
                        "scan_id": "scan1",
                        "vulnerabilities": [{"id": "vulnerability1"}]
                    }
                }]`)),
        }, nil
    } 
}

func getDb() *gorm.DB {
    return OpenAndMigrateDb(":memory:")
}

func getClient() *TestClient {
    return &TestClient{}
}

func TestHandleScan_processesAllFiles(t *testing.T) {
    db := getDb()
    client := getClient()

    ScanRepository("owner/repository", []string{}, db, client)
    var vulnerabilities []Vulnerability
    db.Find(&vulnerabilities)
    assert.Equal(t, 1, len(vulnerabilities), "Unexpected number of results")
}

func TestHandleScan_deduplicates(t *testing.T) {
    db := getDb()
    client := getClient()

    ScanRepository("owner/repository", []string{}, db, client)
    ScanRepository("owner/repository", []string{}, db, client)
    var vulnerabilities []Vulnerability
    db.Find(&vulnerabilities)
    assert.Equal(t, 1, len(vulnerabilities), "Unexpected number of results")
}

func TestHandleScan_processesSpecificFile(t *testing.T) {
    db := getDb()
    client := getClient()

    ScanRepository("owner/repository", []string{"example.json"}, db, client)
    var vulnerabilities []Vulnerability
    db.Find(&vulnerabilities)
    assert.Equal(t, 1, len(vulnerabilities), "Unexpected number of results")
}

func TestHandleScan_skipsUnmatchedFiles(t *testing.T) {
    db := getDb()
    client := getClient()

    ScanRepository("owner/repository", []string{"example2.json"}, db, client)
    var vulnerabilities []Vulnerability
    db.Find(&vulnerabilities)
    assert.Equal(t, 0, len(vulnerabilities), "Unexpected number of results")
}

func TestHandleScan_validRequestSucceeds(t *testing.T) {
    db := getDb()
    client := getClient()
    scanHandler := HandleScan(db, client)

    reader := strings.NewReader(`{"repo":"https://github.com/owner/repository"}`)
    req, _ := http.NewRequest("POST", "/scan", reader)

    rr := httptest.NewRecorder()
    scanHandler.ServeHTTP(rr, req)
    assert.Equal(t, http.StatusOK, rr.Result().StatusCode,
        "Unexpected status code.")
}

func TestHandleScan_reportsInvalidRequest(t *testing.T) {
    db := getDb()
    client := getClient()
    scanHandler := HandleScan(db, client)

    reader := strings.NewReader("invalid json")
    req, _ := http.NewRequest("POST", "/scan", reader)

    rr := httptest.NewRecorder()
    scanHandler.ServeHTTP(rr, req)
    assert.Equal(t, http.StatusBadRequest, rr.Result().StatusCode,
        "Unexpected status code.")
}

func TestHandleScan_reportsInvalidRepository(t *testing.T) {
    db := getDb()
    client := getClient()
    scanHandler := HandleScan(db, client)

    reader := strings.NewReader(`{"repo": "invalid"}`)
    req, _ := http.NewRequest("POST", "/scan", reader)

    rr := httptest.NewRecorder()
    scanHandler.ServeHTTP(rr, req)
    assert.Equal(t, http.StatusBadRequest, rr.Result().StatusCode,
        "Unexpected status code.")
}

