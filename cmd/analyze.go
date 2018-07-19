package cmd

import (
	"fmt"
	"os"
	"text/template"

	"strings"

	"github.com/coreos/clair/api/v1"
	"github.com/coreos/clair/database"
	"github.com/fatih/color"
	"github.com/jgsqware/clairctl/clair"
	"github.com/jgsqware/clairctl/config"
	"github.com/jgsqware/clairctl/docker"
	"github.com/spf13/cobra"
)

const analyzeTplt = `
Image: {{.String}}
 {{range $v := vulns .MostRecentLayer}}
 {{$v | colorized}}{{end}}
`

var filters string
var noFail bool

var analyzeCmd = &cobra.Command{
	Use:   "analyze IMAGE",
	Short: "Analyze Docker image",
	Long:  `Analyze a Docker image with Clair, against Ubuntu, Red hat and Debian vulnerabilities databases`,
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) != 1 {
			fmt.Printf("clairctl: \"analyze\" requires a minimum of 1 argument")
			os.Exit(1)
		}

		config.ImageName = args[0]
		image, manifest, err := docker.RetrieveManifest(config.ImageName, true)
		if err != nil {
			fmt.Println(errInternalError)
			log.Fatalf("retrieving manifest for %q: %v", config.ImageName, err)
		}

		if config.IsLocal {
			startLocalServer()
		}

		if err := clair.Push(image, manifest); err != nil {
			if err != nil {
				fmt.Println(errInternalError)
				log.Fatalf("pushing image %q: %v", image.String(), err)
			}
		}

		analysis := clair.Analyze(image, manifest)

		log.Debug("Using priority filters: ", filters)

		funcMap := template.FuncMap{
			"vulns":     CountVulnerabilities,
			"colorized": colorized,
		}
		err = template.Must(template.New("analysis").Funcs(funcMap).Parse(analyzeTplt)).Execute(os.Stdout, analysis)
		if err != nil {
			fmt.Println(errInternalError)
			log.Fatalf("rendering analysis: %v", err)
		}

		if !isValid(analysis.MostRecentLayer()) && !noFail {
			os.Exit(1)
		}
	},
}

type PriorityCount struct {
	Priority database.Severity
	Count    int
}

func colorized(p PriorityCount) string {
	switch p.Priority {

	case database.UnknownSeverity:
		return color.WhiteString("%v: %v", p.Priority, p.Count)
	case database.NegligibleSeverity:
		return color.HiWhiteString("%v: %v", p.Priority, p.Count)
	case database.LowSeverity:
		return color.YellowString("%v: %v", p.Priority, p.Count)
	case database.MediumSeverity:
		return color.HiYellowString("%v: %v", p.Priority, p.Count)
	case database.HighSeverity:
		return color.MagentaString("%v: %v", p.Priority, p.Count)
	case database.CriticalSeverity:
		return color.RedString("%v: %v", p.Priority, p.Count)
	case database.Defcon1Severity:
		return color.HiRedString("%v: %v", p.Priority, p.Count)
	default:
		return color.WhiteString("%v: %v", p.Priority, p.Count)
	}
}

func isValid(l v1.LayerEnvelope) bool {
	for _, v := range CountVulnerabilities(l) {
		if v.Count != 0 {
			return false
		}
	}

	return true
}

func getPrioritiesFromArgs() []database.Severity {
	f := []database.Severity{}
	for _, aa := range strings.Split(filters, ",") {
		s, err := database.NewSeverity(aa)
		if err == nil {
			f = append(f, s)
		}
	}
	return f
}
func CountVulnerabilities(l v1.LayerEnvelope) []PriorityCount {

	filtersS := getPrioritiesFromArgs()

	if len(filtersS) == 0 {
		filtersS = database.Severities
	}
	r := make(map[database.Severity]int)
	for _, v := range filtersS {
		r[v] = 0
	}

	for _, f := range l.Layer.Features {
		for _, v := range f.Vulnerabilities {
			if _, ok := r[database.Severity(v.Severity)]; ok {
				r[database.Severity(v.Severity)]++
			}
		}
	}

	result := []PriorityCount{}
	for _, p := range database.Severities {
		if pp, ok := r[p]; ok {
			result = append(result, PriorityCount{p, pp})
		}
	}

	return result
}

func init() {
	RootCmd.AddCommand(analyzeCmd)
	analyzeCmd.Flags().StringVarP(&filters, "filters", "f", "", "Filters Severity, comma separated (eg. High,Critical)")
	analyzeCmd.Flags().BoolVarP(&config.IsLocal, "local", "l", false, "Use local images")
	analyzeCmd.Flags().BoolVarP(&noFail, "noFail", "n", false, "Not exiting with non-zero even with vulnerabilities found")
}
