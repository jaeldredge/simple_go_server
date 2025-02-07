package main

// Utility functions that may be reused.

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "regexp"
)

func DecodeDisallowUnknown(input io.ReadCloser, result any) error {
    return decode(input, result, true)
}

func Decode(input io.ReadCloser, result any) error {
    return decode(input, result, false)
}

func decode(input io.ReadCloser, result any, disallowUnknown bool) error {
    d := json.NewDecoder(input)
    if disallowUnknown {
        d.DisallowUnknownFields()
    }

    err := d.Decode(result)
    if err != nil {
        return err
    }
    return nil
}

// Requiring a specific format for GitHub urls.
var gitHubPattern = regexp.MustCompile(`https://github.com/([A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+)`)

func ExtractGitHubRepo(url string) (string, error) {
    matches := gitHubPattern.FindStringSubmatch(url)
    if len(matches) == 0 {
        return "", fmt.Errorf("URL does not match expected GitHub pattern")
    }
    return matches[1], nil
}

func FetchJson(client HttpClient, url string, result any) error {
    req, err := http.NewRequest(http.MethodGet, url, nil)
    if err != nil {
        return err
    }

    // When using the retryablehttp client, failed requests will be retried.
    res, err := client.Do(req)
    if err != nil {
        return err
    }
    defer res.Body.Close()
    err = Decode(res.Body, result)
    return err
}

