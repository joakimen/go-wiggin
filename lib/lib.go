package lib

import (
	"bytes"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

// Lib is the representation of the WigginLib folder
type Lib struct {
	repo string
}

// New returns a lib object with default repo path
func New() Lib {
	repo, ok := os.LookupEnv("WIGGIN_REPO")
	if !ok {
		log.Fatalln("Env var WIGGIN_REPO is not set, exiting.")
	}
	return Lib{repo}
}

// Build builds Intility.Wiggin using msbuild.
// returns error-message instead of error-object, as msbuild doesn't
func (l *Lib) Build() error {

	tmpDir, err := ioutil.TempDir("", "wigginlib")
	if err != nil {
		return err
	}

	// defer cleanup of build-folder
	defer os.RemoveAll(tmpDir)

	project := filepath.Join(
		l.repo,
		"Intility.Wiggin",
		"WigginLib",
		"Intility.Wiggin.csproj",
	)

	cmd := exec.Command(
		"msbuild",
		project,
		"/p:OutputPath="+tmpDir,
		"/clp:ErrorsOnly",
		"/verbosity:quiet",
	)

	// msbuild reports both err and output to stdout..
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	// run build
	err = cmd.Run()

	if err != nil {
		// we gotta create the error object manually, since msbuild doesn't
		// report error to stderr, leading to the original error-object (err)
		// having no error-message.
		return errors.New(string(stdout.Bytes()))
	}

	return nil
}

// Update library from database model data
func (l *Lib) Update() error {

	wyrmRepo, ok := os.LookupEnv("WYRM_REPO")
	if !ok {
		return errors.New("WYRM_REPO is not set")
	}
	var (
		server = os.Getenv("WIGGIN_SERVER")
		db     = os.Getenv("WIGGIN_DB")
		uid    = os.Getenv("WIGGIN_UID")
		pwd    = os.Getenv("WIGGIN_PWD")

		outputDir = filepath.Join(
			l.repo,
			"Intility.Wiggin",
			"WigginLib",
		)
		wyrmSrc    = filepath.Join(wyrmRepo, "src", "Wyrm")
		wyrmProj   = filepath.Join(wyrmSrc, "Wyrm.csproj")
		wyrmBinary = filepath.Join(wyrmSrc,
			"bin",
			"Debug",
			"netcoreapp2.0",
			"netcoreapp2.0",
			"wyrm.dll",
		)

		cmd *exec.Cmd
	)

	// use existing Wyrm binary if present
	if _, err := os.Stat(wyrmBinary); !os.IsNotExist(err) {
		cmd = exec.Command(
			"dotnet", wyrmBinary, "generate", "-s", server,
			"-d", db, "-u", uid, "-p", pwd, "-o", outputDir,
		)
	} else {
		cmd = exec.Command(
			"dotnet", "run", "-p", wyrmProj, "--", "generate",
			"-s", server, "-d", db, "-u", uid, "-p", pwd, "-o", outputDir,
		)
	}

	// run Wyrm, updating WigginLib-dir from db
	err := cmd.Run()

	if err != nil {
		return errors.New("didn't really run man")
	}
	return nil
}
