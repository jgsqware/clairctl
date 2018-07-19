package clair

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/coreos/clair/database"
)

func TestReportAsHtml(t *testing.T) {
	var analysis ImageAnalysis
	err := json.Unmarshal([]byte(getSampleAnalysis()), &analysis)

	if err != nil {
		t.Errorf("Failing with error: %v", err)
	}

	html, err := ReportAsHTML(analysis)
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile(os.TempDir()+"/clairctl-html-report.html", []byte(html), 0700)
	if err != nil {
		log.Fatal(err)
	}
}

func TestInvertedPriorities(t *testing.T) {
	expected := []database.Severity{database.Defcon1Severity, database.CriticalSeverity, database.HighSeverity, database.MediumSeverity, database.LowSeverity, database.NegligibleSeverity, database.UnknownSeverity}
	ip := invertedPriorities()
	for i, v := range ip {
		if v != expected[i] {
			t.Errorf("Expecting %v, got %v", expected, ip)
		}
	}
}
