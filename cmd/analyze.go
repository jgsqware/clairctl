package cmd

import (
	"fmt"
	"os"
	"text/template"

	"strings"

	"github.com/coreos/clair/api/v1"
	"github.com/coreos/clair/utils/types"
	"github.com/jgsqware/clairctl/clair"
	"github.com/jgsqware/clairctl/config"
	"github.com/jgsqware/clairctl/docker"
	"github.com/spf13/cobra"
)

const analyzeTplt = `
Image: {{.String}}
 {{range $v := vulns .MostRecentLayer}}
 {{$v.Priority}}: {{$v.Count}}{{end}}
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
			"vulns": CountVulnerabilities,
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
	Priority types.Priority
	Count    int
}

func isValid(l v1.LayerEnvelope) bool {
	for _, v := range CountVulnerabilities(l) {
		if v.Count != 0 {
			return false
		}
	}

	return true
}

func getPrioritiesFromArgs() []types.Priority {
	f := []types.Priority{}
	for _, aa := range strings.Split(filters, ",") {
		if types.Priority(aa).IsValid() {
			f = append(f, types.Priority(aa))
		}
	}
	return f
}
func CountVulnerabilities(l v1.LayerEnvelope) []PriorityCount {

	filtersS := getPrioritiesFromArgs()

	if len(filtersS) == 0 {
		filtersS = types.Priorities
	}
	r := make(map[types.Priority]int)
	for _, v := range filtersS {
		r[v] = 0
	}

	for _, f := range l.Layer.Features {
		for _, v := range f.Vulnerabilities {
			if _, ok := r[types.Priority(v.Severity)]; ok {
				r[types.Priority(v.Severity)]++
			}
		}
	}

	result := []PriorityCount{}
	for _, p := range types.Priorities {
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
