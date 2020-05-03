package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"text/tabwriter"
	"text/template"

	flag "github.com/spf13/pflag"
)

type commitDetail struct {
	urlTpl string
	name   string
}

var commitDetails = []commitDetail{
	{"https://%s-blue1.example.com", "server1 Blue"},
	{"https://%s-green1.example.com", "server1 Green"},
	{"https://%s-blue-r1.example.com", "server2r1 Blue"},
	{"https://%s-green-r1.example.com", "server2r1 Green"},
	{"https://%s-blue-r2.example.com", "server2r2 Blue"},
	{"https://%s-green-r2.example.com", "server2r2 Green"},
}

func main() {
	flags := flag.NewFlagSet("toynbee-tiles", flag.ContinueOnError)
	open := flags.BoolP("open", "b", false, "open in browser instead of printing")
	commitTpl := flags.StringP("template", "t", `{{printf "commitDetail: %s\tCommit: %s\n" .commitDetail .commit}}`, "how to extract the commit")
	path := flags.StringP("path", "p", "/-/info", "the path to open")
	flags.Usage = printHelpPrompt
	if err := flags.Parse(os.Args[1:]); err != nil {
		if err == flag.ErrHelp {
			os.Exit(0)
		}

		fmt.Println(err)
		os.Exit(1)
	}

	args := flags.Args()

	if len(args) < 1 {
		fmt.Println("at least one PROJECT must be provided")
		os.Exit(1)
	}

	printer, err := NewCommitDetailTemplatePrinter(*commitTpl)
	if err != nil {
		log.Fatal(err)
	}

	w := tabwriter.NewWriter(os.Stdout, 8, 8, 8, ' ', 0)

	for _, app := range args {
		for _, commitDetail := range commitDetails {
			endpoint := fmt.Sprintf(commitDetail.urlTpl, app) + *path
			if *open {
				openbrowser(endpoint)
				continue
			}

			resp, err := http.Get(endpoint)
			if err != nil {
				log.Printf("failed to get info for %s\n", endpoint)
				continue
			}
			defer resp.Body.Close()
			buf, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Printf("failed to read response for %s: %v\n", endpoint, err)
				continue
			}

			printer.Print(buf, commitDetail.name, w)
		}
	}
	w.Flush()
}

type commitDetailTemplatePrinter struct {
	rawTemplate string
	template    *template.Template
}

func NewCommitDetailTemplatePrinter(tmpl string) (*CommitDetailTemplatePrinter, error) {
	t, err := template.New("output").Parse(string(tmpl))
	if err != nil {
		return nil, err
	}
	return &commitDetailTemplatePrinter{
		rawTemplate: string(tmpl),
		template:    t,
	}, nil
}

func (p *commitDetailTemplatePrinter) Print(data []byte, commitDetailName string, w io.Writer) error {
	out := map[string]interface{}{}
	if err := json.Unmarshal(data, &out); err != nil {
		return err
	}

	out["commitDetail"] = commitDetailName

	if err := p.safeExecute(w, out); err != nil {
		fmt.Fprintf(w, "Error executing template: %v. Printing more information for debugging the template:\n", err)
		fmt.Fprintf(w, "\ttemplate was:\n\t\t%v\n", p.rawTemplate)
		fmt.Fprintf(w, "\traw data was:\n\t\t%v\n", string(data))
		fmt.Fprintf(w, "\tobject given to template engine was:\n\t\t%+v\n\n", out)
		return fmt.Errorf("error executing template %q: %v", p.rawTemplate, err)
	}
	return nil
}

func (p *commitDetailTemplatePrinter) safeExecute(w io.Writer, obj interface{}) error {
	var panicErr error
	retErr := func() error {
		defer func() {
			if x := recover(); x != nil {
				panicErr = fmt.Errorf("caught panic: %+v", x)
			}
		}()
		return p.template.Execute(w, obj)
	}()
	if panicErr != nil {
		return panicErr
	}
	return retErr
}

func openbrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("Sorry, this platform is not supported.")
	}
	if err != nil {
		log.Fatal(err)
	}

}

func printHelpPrompt() {
	helpText := `
Usage: toynbee-tiles [options] PROJECT [PROJECT ...]

  Helps you verify commit details in production. This will either launch in the browser or extract
  the commit from the following URL patterns.

	"https://{app}-blue1.example.com",
    "https://{app}-green1.example.com",
    "https://{app}-blue-r1.example.com",
    "https://{app}-green-r1.example.com",
    "https://{app}-blue-r2.example.com",
    "https://{app}-green-r2.example.com",

Options:

  -p,--path string       The URL path to either launch or to extract version info from.

  -t,--template string   A Go template that will be printed for each endpoint based on extracting
                         the version information. This is mutually exclusive to the --open option.
                         A {{.commitDetail}} is available with the datacenter name.

  -b,--open              Open every production endpoint in the browser.

`

	_, _ = fmt.Fprintln(os.Stdout, strings.TrimSpace(helpText))
}
