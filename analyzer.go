package main

import (
  "database/sql"
  "fmt"

  "github.com/lib/pq"
)

func analyzeJob(job Job) {
    var db = connectToDatabase()
    fmt.Printf("\ndb info: %+v\n\n", db)
    defer db.Close()
    //Save job to jobs table
    saveJobInfo(job, db)
    // err := addToAnalysis(job, db)
}

func saveJobInfo(job Job, db *sql.DB) {
    sqlStatement := `
        INSERT INTO Jobs (url, selectors, success, capturama_code)
        VALUES ($1, $2, $3, $4)`
    _, err := db.Exec(
        sqlStatement,
        job.URL,
        pq.Array(job.Selectors),
        job.Success,
        job.CapturamaHttpCode,
    )
    if err != nil {
      fmt.Printf("Error saving Job info: %v\n", err)
    }
}

func addToAnalysis(job Job, db sql.DB) {
    //Check for domain existing on DB to add to
}
