package main

// Contains schema and function for opening a database connection.

import (
    "database/sql/driver"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
    "gorm.io/gorm/schema"
    "log"
    "strings"
    "time"
)

// This code is to support storing an array of strings in a sqlite column
// needed for Vulnerability.RiskFactors.
type Strings []string

func (s *Strings) Scan(src interface{}) error {
	*s = strings.Split(src.(string), "|")
	return nil
}

func (s Strings) Value() (driver.Value, error) {
	if len(s) == 0 {
		return nil, nil
	}
	return strings.Join(s, "|"), nil
}

func (Strings) GormDataType() string {
	return "text"
}

func (Strings) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	return "text"
}

// Using GORM to handle persisting and querying
// Separate tables for scans and vulnerabilities
// Scans and vulnerabilites are uniquely identified by their ids.  Records can
// be updated, but not duplicated.  In some cases, this may result in a
// vulnerability moving from one scan to another.  Could be changed depending
// on the requirements.
type Scan struct {
    ScanId string `gorm:"primaryKey" json:"scan_id"`
    Filename string
    ProcessTime time.Time
    ScanTime *string `json:"timestamp"`
    Vulnerabilities []Vulnerability `gorm:"foreignKey:ScanId;references:ScanId" json:"vulnerabilities"`
}

type Vulnerability struct {
    ScanId string 
    VulnerabilityId string `gorm:"primaryKey" json:"id"`
    Severity *string `json:"severity"`
    CVSS *float32 `json:"cvss"`
    Status *string `json:"status"`
    PackageName *string `json:"package_name"`
    CurrentVersion *string `json:"current_version"`
    FixedVersion *string `json:"fixed_version"`
    Description *string `json:"description"`
    PublishedDate *string `json:"published_date"`
    RiskFactors *Strings `gorm:"type:text" json:"risk_factors"`
}

// Opens database and updates the schema.  For testing a memory database can
// be used.
func OpenAndMigrateDb(databaseFile string) *gorm.DB {
    db, err := gorm.Open(sqlite.Open(databaseFile), &gorm.Config{})
    if err != nil {
        log.Fatal(err)
    }

    err = db.AutoMigrate(&Scan{})
    if err != nil {
        log.Fatal(err)
    }

    err = db.AutoMigrate(&Vulnerability{})
    if err != nil {
        log.Fatal(err)
    }

    return db
}

