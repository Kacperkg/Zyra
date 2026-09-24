package services

import (
	"strings"
	"testing"
	"zyra-api/internal/models"
)

func sample(sections string) string {
	return "ReportOn: 23-sep-2026:11:27:59, PkgVersion: 2.5\nDatabase: ifsprd\nDatabase checks\n" + sections + "\nScript Info\nRun by : user@server01\n"
}
func TestReportSelectedChecks(t *testing.T) {
	free := 5.0
	settings := models.CheckSettings{Selected: map[string]bool{"backups": true, "tablespace": true, "missing_email": true}, Resources: map[string]map[string]models.ResourceRule{
		"backups":    {`E:\ORADATA\IFSPRD\A.DBF`: {}, `E:\ORADATA\IFSPRD\B.DBF`: {}},
		"tablespace": {"UNDOTBS1": {MinFreePercent: &free}},
	}}
	body := sample("Backups=\\\nDatafiles needing backup:\nRMAN E:\\ORADATA\\IFSPRD\\A.DBF 22-SEP-2026 00:42:16\nRMAN E:\\ORADATA\\IFSPRD\\B.DBF 22-SEP-2026 00:42:16\nTablespaces=\\\nName Total MB Max MB Used MB %Free\nUNDOTBS1 31744 31744 30900 2.7%\nIndexes=unusable indexes\nFilesystem Usage\nE: 91% - 1Gb free")
	p := ParseReport(body, settings)
	if !p.Complete || p.Hostname != "server01" {
		t.Fatalf("metadata: %+v", p)
	}
	failed := []models.CheckResult{}
	for _, r := range p.Results {
		if r.Status == "failed" {
			failed = append(failed, r)
		}
	}
	if len(failed) != 2 || failed[0].CheckType != "backups" || failed[1].CheckType != "tablespace" {
		t.Fatalf("failures: %+v", failed)
	}
	if !strings.Contains(failed[0].Evidence, `E:\ORADATA\IFSPRD\B.DBF`) {
		t.Fatal("backup grouping lost row")
	}
}
func TestNotUSAndIncomplete(t *testing.T) {
	settings := models.CheckSettings{Selected: map[string]bool{"backups": true, "missing_email": true}}
	p := ParseReport(sample("Backups=NOT_US"), settings)
	if !p.Complete || p.Results[0].Status != "not_managed" {
		t.Fatalf("%+v", p)
	}
	p = ParseReport("ORA-06550\nPLS-00201: identifier must be declared\nFilesystem Usage\nE: 91% - 1Gb free\nScript Info\nRun by : user@server01", settings)
	if p.Complete || len(p.Results) != 1 || p.Results[0].CheckType != "missing_email" || p.Results[0].Status != "failed" {
		t.Fatalf("%+v", p)
	}
}

func TestWindowsFooterRequiresRunBy(t *testing.T) {
	settings := models.CheckSettings{Selected: map[string]bool{"backups": true, "missing_email": true}}
	for name, body := range map[string]string{
		"heading only":               "ReportOn: 23-sep-2026:11:27:59, PkgVersion: 2.5\nDatabase: ifsprd\nDatabase checks\nBackups=OK!\nScript Info\n",
		"script path without run by": "ReportOn: 23-sep-2026:11:27:59, PkgVersion: 2.5\nDatabase: ifsprd\nDatabase checks\nBackups=OK!\nScript Info\nScript : E:\\checks\\dcx.cmd\n",
	} {
		t.Run(name, func(t *testing.T) {
			p := ParseReport(body, settings)
			if p.Complete || len(p.Results) != 1 || p.Results[0].CheckType != "missing_email" || p.Results[0].Status != "failed" {
				t.Fatalf("%+v", p)
			}
		})
	}
}
func TestUnknownIsNotPassed(t *testing.T) {
	p := ParseReport(sample("Backups=UNRECOGNISED"), models.CheckSettings{Selected: map[string]bool{"backups": true}})
	if p.Results[0].Status != "unknown" {
		t.Fatal(p.Results)
	}
}
