package main

import (
  "database/sql"
  "fmt"
  "net/url"

  "github.com/lib/pq"
  "github.com/rs/xid"
)

type Analytics struct {
    Domain string
    TimesHit int
    Guid string
}

func analyzeJob(job Job) {
    var db = connectToDatabase()
    defer db.Close()
    //Save job to jobs table
    saveJobInfo(job, db)
    addToAnalytics(job, db)
}

func saveJobInfo(job Job, db *sql.DB) {
    sqlStatement := `
        INSERT INTO jobs (url, selectors, success, capturama_code, guid)
        VALUES ($1, $2, $3, $4, $5)`
    _, err := db.Exec(
        sqlStatement,
        job.URL,
        pq.Array(job.Selectors),
        job.Success,
        job.CapturamaHttpCode,
        genID(),
    )
    if err != nil {
      fmt.Printf("Error saving Job info: %v\n", err)
    }
}

func addToAnalytics(job Job, db *sql.DB) {
    analytics := Analytics{}
    //Get domain from url
    urlObj, err := url.Parse(job.URL)
	if err != nil {
		fmt.Printf("Error parsing url: %v\n", err)
		return
	}

    analytics.Domain = urlObj.Host

    //Check for domain existing on DB
    sqlStatement := `SELECT domain, times_hit, guid FROM analytics WHERE domain=$1;`
    row := db.QueryRow(sqlStatement, analytics.Domain)
    analytics.TimesHit = 1
    analytics.Guid = ""
    err = row.Scan(&analytics.Domain, &analytics.TimesHit, &analytics.Guid)
    switch err {
    case sql.ErrNoRows:
        //Insert record
        sqlStatement := `
            INSERT INTO analytics (domain, times_hit, guid)
            VALUES ($1, $2, $3)`
        _, err := db.Exec(
            sqlStatement,
            analytics.Domain,
            analytics.TimesHit,
            genID(),
        )
        if err != nil {
          fmt.Printf("Error saving Analytics info: %v\n", err)
        }
        fmt.Println("Entry added!")
    case nil:
        //Update record
        fmt.Printf("Found entry! Times hit before updating: %v\n", analytics.TimesHit)
        sqlStatement := `
            UPDATE analytics
            SET times_hit = $2
            WHERE guid = $1;`

        _, err = db.Exec(sqlStatement, analytics.Guid, analytics.TimesHit + 1)
        if err != nil {
          panic(err)
        }
        fmt.Println("Entry updated!")
    default:
        fmt.Printf("Error scanning: %v\n", err)
    }
}

func genID() string {
    return xid.New().String()
}
