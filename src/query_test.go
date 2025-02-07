package main

// Tests for the /query endpoint and associated functionality.

import (
    "github.com/stretchr/testify/assert"
    "io"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"
)

func stringPointer(s string) *string {
    return &s
}

var highSeverityFixed Vulnerability = Vulnerability{
    ScanId: "scan1",
    VulnerabilityId: "vulnerability1",
    Severity: stringPointer("HIGH"),
    Status: stringPointer("fixed"),
}

var mediumSeverityActive Vulnerability = Vulnerability{
    ScanId: "scan1",
    VulnerabilityId: "vulnerability2",
    Severity: stringPointer("MEDIUM"),
    Status: stringPointer("active"),
}

var lowSeverityActive Vulnerability = Vulnerability{
    ScanId: "scan2",
    VulnerabilityId: "vulnerability3",
    Severity: stringPointer("LOW"),
    Status: stringPointer("active"),
}

var scan1 Scan = Scan{
    ScanId: "scan1",
    Vulnerabilities: []Vulnerability{highSeverityFixed, mediumSeverityActive},
}

var scan2 Scan = Scan{
    ScanId: "scan2",
    Vulnerabilities: []Vulnerability{lowSeverityActive},
}

var queryHandler http.Handler

func init() {
    db := OpenAndMigrateDb(":memory:")
    db.Save(&scan1)
    db.Save(&scan2)
    queryHandler = HandleQuery(db)
}

func TestHandleQuery_returnsAllResults(t *testing.T) {
    reader := strings.NewReader(`{"filters":{}}`)
    req, _ := http.NewRequest("POST", "/query", reader)

    rr := httptest.NewRecorder()
    queryHandler.ServeHTTP(rr, req)
    assert.Equal(t, http.StatusOK, rr.Code, "Unexpected status code.")

    expected := []QueryVulnerability{
        Convert(highSeverityFixed),
        Convert(mediumSeverityActive),
        Convert(lowSeverityActive)}
    var actual []QueryVulnerability
    err := Decode(io.NopCloser(rr.Body), &actual)
    if err != nil {
        t.Fatal(err)
    }
    assert.ElementsMatch(t, expected, actual, "Unexpected response")
}

func TestHandleQuery_appliesSeverityFilter(t *testing.T) {
    reader := strings.NewReader(`{"filters":{"severity":"HIGH"}}`)
    req, _ := http.NewRequest("POST", "/query", reader)

    rr := httptest.NewRecorder()
    queryHandler.ServeHTTP(rr, req)
    assert.Equal(t, http.StatusOK, rr.Code, "Unexpected status code.")

    expected := []QueryVulnerability{Convert(highSeverityFixed)}
    var actual []QueryVulnerability
    err := Decode(io.NopCloser(rr.Body), &actual)
    if err != nil {
        t.Fatal(err)
    }
    assert.ElementsMatch(t, expected, actual, "Unexpected response")
}

func TestHandleQuery_appliesStatusFilter(t *testing.T) {
    reader := strings.NewReader("{\"filters\":{\"status\":\"active\"}}")
    req, _ := http.NewRequest("POST", "/query", reader)
    
    rr := httptest.NewRecorder()
    queryHandler.ServeHTTP(rr, req)
    assert.Equal(t, http.StatusOK, rr.Code, "Unexpected status code.")
    
    expected := []QueryVulnerability{
        Convert(mediumSeverityActive),
        Convert(lowSeverityActive),
    }
    var actual []QueryVulnerability
    err := Decode(io.NopCloser(rr.Body), &actual)
    if err != nil {
        t.Fatal(err)
    }
    assert.ElementsMatch(t, expected, actual, "Unexpected response")
} 

func TestHandleQuery_reportsInvalidRequest(t *testing.T) {
    reader := strings.NewReader("invalid json")
    req, _ := http.NewRequest("POST", "/query", reader)

    rr := httptest.NewRecorder()
    queryHandler.ServeHTTP(rr, req)
    assert.Equal(t, http.StatusBadRequest, rr.Result().StatusCode, 
        "Unexpected status code.")
}

func TestConvert_setsField(t *testing.T) {
    expected := "fixed"
    assert.Equal(t, &expected, Convert(highSeverityFixed).Status, "Field not set")
}

