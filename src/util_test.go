package main

// Tests for utility methods.

import (
    "github.com/stretchr/testify/assert"
    "testing"
)

func TestExtractGitHubRepo_extractsRespository(t *testing.T) {
    repo, err := ExtractGitHubRepo("https://github.com/owner/repository")
    assert.Equal(t, "owner/repository", repo, "Unexpected repository")
    assert.Empty(t, err, "No error expected")
}

func TestExtractGitHubRepo_reportsInvalidFormat(t *testing.T) {
    repo, err := ExtractGitHubRepo("incorrect")
    assert.Error(t, err, "URL does not match expected GitHub pattern")
    assert.Empty(t, repo, "No repostiory expected")
}

