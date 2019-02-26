package main

import (
	"flag"
	"fmt"
	"github.com/krystah/cid"
	"github.com/krystah/wiggin/db"
	"github.com/krystah/wiggin/lib"
	"github.com/krystah/wiggin/repo"
	"log"
	"os"
)

func main() {

	// use user-supplied connection-string if desired
	conStr := flag.String("conStr", "", "Connection String")
	flag.Parse()

	// ensure required env var is set
	path, ok := os.LookupEnv("WIGGIN_REPO")
	if !ok {
		log.Fatalln("Environment variable WIGGIN_REPO is not set, exiting")
	}

	// create waitgroup for thread grouping
	r := cid.NewRunner()

	/* Repo-checks */
	rep := repo.New(path)

	r.AddWork(rep.CheckEmptyFiles, "repo", "empty files")
	r.AddWork(rep.CheckFunctionMissingTable, "repo", "rls functions missing tables")
	r.AddWork(rep.CheckPolicyMissingTable, "repo", "policies missing tables")
	r.AddWork(rep.CheckLibsMissingSchema, "repo", "libs missing schemas")

	/* Lib-checks */
	l := lib.New()

	// these must be run synchronously
	queue := cid.Queue{
		cid.NewWork(l.Update, "lib", "update"),
		cid.NewWork(l.Build, "lib", "build"),
	}
	r.AddQueue(queue)

	/* DB-checks */
	var c db.ConnMgr
	if *conStr != "" {
		c.ConnStr = *conStr
	} else {
		c.ConnStr = db.GetConnStr(db.GetDefaults())
	}
	c.Connect()
	r.AddWork(c.RunTests, "db", "Running Tests")

	// execute registered queues
	r.Exec()

	fmt.Println()

	// print test results
	c.PrintResults()   // db
	rep.PrintResults() // repo

	fmt.Println("\ndone.")

}
