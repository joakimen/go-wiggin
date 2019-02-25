package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"

	_ "github.com/denisenkom/go-mssqldb" // mssql implementation
)

// Prop holds the connection string attributes
type Prop struct {
	server string
	db     string
	uid    string
	pwd    string
}

// ConnMgr holds the connection object and connection string
type ConnMgr struct {
	DB      *sql.DB
	ConnStr string
	Errors  []string
}

// GetDefaults returns default connection string properties
func GetDefaults() Prop {

	var p Prop

	p.server = os.Getenv("WIGGIN_SERVER")
	if p.server == "" {
		log.Fatalln("Env var WIGGIN_SERVER is not set")
	}

	p.db = os.Getenv("WIGGIN_DB")
	if p.db == "" {
		log.Fatalln("Env var WIGGIN_DB is not set")
	}

	p.uid = os.Getenv("WIGGIN_UID")
	if p.uid == "" {
		log.Fatalln("Env var WIGGIN_UID is not set")
	}

	p.pwd = os.Getenv("WIGGIN_PWD")
	if p.pwd == "" {
		log.Fatalln("Env var WIGGIN_PWD is not set")
	}

	return p
}

// GetConnStr returns a formatted Connection String using Props-struct
func GetConnStr(p Prop) string {
	return fmt.Sprintf("sqlserver://%s:%s@%s?database=%s",
		p.uid, p.pwd, p.server, p.db,
	)
}

// Connect connects to instance
func (c *ConnMgr) Connect() {

	// connect to sql server
	var err error
	c.DB, err = sql.Open("sqlserver", c.ConnStr)

	if err != nil {
		log.Fatalln(err)
	}
}

// RunTests executes database tests to verify integrity
func (c *ConnMgr) RunTests() error {

	// wake up connection if closed
	c.DB.Ping()
	defer c.DB.Close()
	query := "SET NOCOUNT ON; EXEC tSQLt.RunAll"

	// run tests
	var ctx = context.Background()
	_, err := c.DB.ExecContext(ctx, query)

	// If tests fail, run them again using os.exec to capture test details.
	// The reason for this is that the rest results, in the case of a failed
	// test, the errors
	if err != nil {
		c.Errors = append(c.Errors, err.Error())

		prop := GetDefaults()
		cmd := exec.Command(
			"sqlcmd",
			"-S", prop.server,
			"-d", prop.db,
			"-U", prop.uid,
			"-P", prop.pwd,
			"-Q", query,
		)

		// capture output from test execution
		out, err := cmd.CombinedOutput()
		if err != nil {
			c.Errors = append(c.Errors, err.Error())
		}

		// print output
		c.Errors = append(c.Errors, string(out))
	}

	return err
}

// PrintResults prints resulting errors, if any
func (c *ConnMgr) PrintResults() {
	if len(c.Errors) == 0 {
		return
	}

	fmt.Println(("db errors:"))
	for _, e := range c.Errors {
		fmt.Println("*", e)
	}
	fmt.Println()
}
