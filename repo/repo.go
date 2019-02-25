package repo

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
)

// Repo represents a git repository
type Repo struct {
	path                  string
	emptyFiles            []string
	functionsMissingTable []string
	policiesMissingTable  []string
	libsMissingSchema     []string
}

// New returns a new Repo object
func New(path string) Repo {
	return Repo{
		path: path,
	}
}

// CheckEmptyFiles returns empty files in repo
func (r *Repo) CheckEmptyFiles() error {

	validExtensions := []string{".sql", ".cs"}

	err := filepath.Walk(r.path,
		func(path string, f os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// only act on files with the desired extension and size
			if contains(filepath.Ext(f.Name()), validExtensions) && f.Size() < 3 {
				r.emptyFiles = append(r.emptyFiles, path)
			}
			return nil
		})

	if err != nil {
		return err
	}
	return nil
}

// CheckPolicyMissingTable returns policies missing associated table
func (r *Repo) CheckPolicyMissingTable() error {

	pat := `Security\.(?P<schema>\w+)_(?P<object>\w+)\.sql`
	re := regexp.MustCompile(pat)
	path := filepath.Join(
		r.path,
		"WigginDB",
		"Security Policies",
	)

	result, err := r.checkIfMissingTable(re, path)
	if err != nil {
		return err
	}
	r.policiesMissingTable = result

	return nil
}

// CheckFunctionMissingTable lists RLS functions missing associated table
func (r *Repo) CheckFunctionMissingTable() error {

	pat := `Security\.fn_RLS_Read_(?P<schema>\w+)_(?P<object>\w+)\.sql`
	re := regexp.MustCompile(pat)
	path := filepath.Join(
		r.path,
		"WigginDB",
		"Functions",
	)

	result, err := r.checkIfMissingTable(re, path)
	if err != nil {
		return err
	}
	r.functionsMissingTable = result

	return nil
}

// getMatchList returns schema- and object-name from the supplied string/regex
func (r *Repo) getMatchList(re *regexp.Regexp, str string) map[string]string {
	result := make(map[string]string)

	if !re.MatchString(str) {
		return result
	}

	match := re.FindStringSubmatch(str)
	for i, name := range re.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}
	return result
}

// checkIfMissingTable checks if any of the objects in the specified "path" is
// missing the associated table
func (r *Repo) checkIfMissingTable(re *regexp.Regexp, path string) ([]string, error) {

	var result []string

	tablePath := filepath.Join(r.path, "WigginDB", "Tables")
	files, err := listDir(path)

	if err != nil {
		return nil, err
	}

	for _, f := range files {

		// skip file if it doesn't match regex
		if !re.MatchString(f.Name()) {
			continue
		}

		// extract schema- and object-name and test if table exists
		schema, object := r.extractNameParts(f.Name(), re)
		if !tableExists(tablePath, schema, object) {
			// add object to results if table is missing
			result = append(result, f.Name())
		}
	}

	return result, nil
}

// CheckLibsMissingSchema lists libraries without associated db schema
func (r *Repo) CheckLibsMissingSchema() error {

	var (
		libPath = filepath.Join(
			r.path,
			"Intility.Wiggin",
			"WigginLib",
		)

		excludedItems = []string{
			"Intility.Wiggin.csproj", "Settings.cs", "app.config", "packages.config",
			"_Data", "_Entity", "bin", "obj", "_Repo", "Properties",
		}
	)

	files, err := listDir(libPath)
	if err != nil {
		return err
	}

	for _, f := range files {

		// skip regular files
		if !f.IsDir() {
			continue
		}

		// skip if item is in blacklist
		if contains(f.Name(), excludedItems) {
			continue
		}

		// check if schema exists
		if !fileExists(filepath.Join(
			r.path,
			"WigginDB",
			"Security",
			"Schemas",
			f.Name()+".sql",
		)) {
			r.libsMissingSchema = append(r.libsMissingSchema, f.Name())
		}
	}

	return nil
}

func contains(needle string, haystack []string) bool {
	for _, s := range haystack {
		if needle == s {
			return true
		}
	}
	return false
}

func (r *Repo) extractNameParts(filename string, re *regexp.Regexp) (schema string, object string) {

	matchList := make(map[string]string)
	matchList = r.getMatchList(re, filename)

	schema = matchList["schema"]
	object = matchList["object"]

	return
}

func tableExists(tableDir string, schema string, object string) bool {
	table := filepath.Join(tableDir, fmt.Sprintf("%s.%s.sql", schema, object))
	return fileExists(table)
}

func listDir(path string) ([]os.FileInfo, error) {
	return ioutil.ReadDir(path)
}

func fileExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

// PrintResults dumps result variables
func (r *Repo) PrintResults() {

	if len(r.emptyFiles) > 0 {
		fmt.Println("empty files:", len(r.emptyFiles))
		printSlice(r.emptyFiles)
	}

	if len(r.functionsMissingTable) > 0 {
		fmt.Println("rls functions missing table:", len(r.functionsMissingTable))
		printSlice(r.functionsMissingTable)
	}

	if len(r.policiesMissingTable) > 0 {
		fmt.Println("policies missing table:", len(r.policiesMissingTable))
		printSlice(r.policiesMissingTable)
	}

	if len(r.libsMissingSchema) > 0 {
		fmt.Println("libs missing schema:", len(r.libsMissingSchema))
		printSlice(r.libsMissingSchema)
	}
}

func printSlice(s []string) {
	if len(s) > 0 {
		for _, e := range s {
			fmt.Println("*", e)
		}
		fmt.Println()
	}
}
