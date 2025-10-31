package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/spf13/pflag"
	"github.com/vilasle/backilli/internal/report"
)

const (
	errStatus = "error"
)

type Args struct {
	ReportDir string
	Remove    bool
	ShowHelp  bool
}

type Result struct {
	Date     time.Time
	HasError bool
	Errors   []string
}

func cliInit() (args Args) {
	pflag.StringVarP(&args.ReportDir, "dir", "d", "", "directory with reports")
	pflag.BoolVarP(&args.Remove, "rm", "", false, "remove reports after handling")
	pflag.BoolVarP(&args.ShowHelp, "help", "h", false, "show help information")
	pflag.Parse()

	return args
}

// TODO send information about error to Telegram
// TODO read task file and check there are all necessary copies
// TODO read task file and check all tasks were executed
func main() {
	args := cliInit()

	if args.ShowHelp {
		pflag.Usage()
		os.Exit(0)
	}

	if args.ReportDir == "" {
		fmt.Println("does not pass 'dir' argument")
		pflag.Usage()
		os.Exit(1)
	}

	reports, err := getPathOfReports(args.ReportDir)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	results := make([]Result, 0, len(reports))
	for _, p := range reports {
		result, err := handleReport(p, args.Remove)
		if err != nil {
			fmt.Printf("could not handle report '%s' by reason %v\n", p, err)
			continue
		}
		results = append(results, result)
	}

	//flush to stdout
	hasError := false
	for _, r := range results {
		if !r.HasError {
			continue
		}
		hasError = true
		fmt.Printf("\nreport %s has errors:\n", r.Date.Format("02-01-2006"))
		for _, err := range r.Errors {
			fmt.Printf("\t%s\n", err)
		}
	}

	if not(hasError) {
		fmt.Println("\nreports do not have errors")
	}
}

func not(exp bool) bool {
	return !exp
}

func getPathOfReports(reportPath string) ([]string, error) {
	s, err := os.Stat(reportPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.Join(err, fmt.Errorf("directory %s is not exists", reportPath))
		} else {
			return nil, errors.Join(err, fmt.Errorf("unexpected error"))
		}
	}

	if !s.IsDir() {
		return nil, fmt.Errorf("file %s is not directory", reportPath)
	}

	ls, err := os.ReadDir(reportPath)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("could not read directory '%s'", reportPath))
	}

	exp := regexp.MustCompile("^report_[0-9]{2}-[0-9]{2}-[0-9]{4}.json$")
	result := make([]string, 0, len(ls))
	for _, f := range ls {
		if f.IsDir() {
			continue
		}

		n := f.Name()

		if !exp.MatchString(n) {
			continue
		}

		result = append(result, filepath.Join(reportPath, n))
	}
	return result, nil
}

func handleReport(path string, remove bool) (result Result, err error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return result, err
	}
	//remove report
	defer func() {
		if remove {
			if err := os.Remove(path); err != nil {
				fmt.Printf("could not remove file '%s' by reason %v\n", path, err)
			}
		}
	}()

	//get report as object
	reports, err := parseReport(raw)
	if err != nil {
		return result, err
	}

	//define date of report. expected date format dd.mm.yyyy
	exp := regexp.MustCompile(`(3[01]|[12][0-9]|0?[1-9])-(0[1-9]|1[0-2])-(19|20)[0-9]{2}`)
	d := exp.FindString(path)
	if d == "" {
		return result, fmt.Errorf("not found date in report's name")
	}

	//parse date as dd.mm.yyyy
	date, err := time.Parse("02-01-2006", d)
	if err != nil {
		return result, errors.Join(err, fmt.Errorf("could not parse date '%s' by format 'dd-mm-yyyy'", d))
	}
	// select detail from report where status equal 'error'
	result.Date = date
	result.Errors, result.HasError = lookForError(reports)

	return result, err
}

func parseReport(raw []byte) (rep []report.Report, err error) {
	rep = make([]report.Report, 0)
	if err = json.Unmarshal(raw, &rep); err != nil {
		return
	}
	return
}

func lookForError(reports []report.Report) (problems []string, hasError bool) {
	problems = make([]string, 0, len(reports))
	for _, r := range reports {
		if r.Status == errStatus {
			problems = append(problems, r.Details)
		}
	}
	return problems, len(problems) > 0
}
