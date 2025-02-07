package main

// Implementation of the /query endpoint.

import (
    "encoding/json"
    "gorm.io/gorm"
    "log"
    "net/http"
)

// Expected request format.
type QueryRequest struct {
    Filters QueryVulnerability `json:"filters"`
}

// API version of a Vulnerability.  Kept distinct from the database version.
type QueryVulnerability struct {
    VulnerabilityId string `json:"id"`
    Severity *string `json:"severity"`
    CVSS *float32 `json:"cvss"`
    Status *string `json:"status"`
    PackageName *string `json:"package_name"`
    CurrentVersion *string `json:"current_version"`
    FixedVersion *string `json:"fixed_version"`
    Description *string `json:"description"`
    PublishedDate *string `json:"published_date"`
    RiskFactors *Strings `json:"risk_factors"`
}

func HandleQuery(db *gorm.DB) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
        var qr QueryRequest
        err := DecodeDisallowUnknown(req.Body, &qr)
        if err != nil {
            log.Print(err)
            http.Error(w,
                "Request must be in the format {'filters': {'KEY': 'VALUE'}} " +
                "where KEY is the property of a vulnerability and VALUE is " +
                "the value to filter on",
                http.StatusBadRequest)
            return
        } 

        // Query and return results. It would be good to add a limit here or
        // paging.
        var vulnerabilities []Vulnerability
        db.Where(qr.Filters).Find(&vulnerabilities)

        results := make([]QueryVulnerability, len(vulnerabilities)) 
        for i, v := range vulnerabilities {
            results[i] = Convert(v)
        }

        w.Header().Set("Content-Type", "application/json")
        err = json.NewEncoder(w).Encode(results)
        if err != nil {
            log.Print(err)
        }
    })
}

// Convert to API version.
func Convert(v Vulnerability) QueryVulnerability {
    return QueryVulnerability {
            VulnerabilityId: v.VulnerabilityId, 
            Severity: v.Severity,
            CVSS: v.CVSS,
            Status: v.Status,
            PackageName: v.PackageName,
            CurrentVersion: v.CurrentVersion,
            FixedVersion: v.FixedVersion,
            Description: v.Description,
            PublishedDate: v.PublishedDate,
            RiskFactors: v.RiskFactors,
        }
}
