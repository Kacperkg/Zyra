package services

import (
	"strings"
	"testing"
	"zyra-api/internal/models"
)

const linuxReport = `ReportOn: 23-sep-2026:20:30:01, PkgVersion: 2.5
Database: exampledb, Os:Linux
Database checks
Backups=OK!
FailedJobs=\
Please check the following failed jobs: \
JobID Last Success Broken Failures \
343 Y 16 \
243 Y 16 \
403 Y 16 \
344 Y 16 \
Database Stats
database.uptime=125.35Days
=@=
Filesystem                   Mounted             Use%
tmpfs                        /dev/shm            1%
tmpfs                        /run                3%
/dev/mapper/data              /data               90%
example-server:/backups       /mnt/backups        92%
----------------------
Script : /home/oracle/checks/dcx.ksh
Version : 2.1
Run by : oracle@EXAMPLE-HOST
- Schedule -
30 5,20 * * * /home/oracle/checks/dcx.ksh > /dev/null
----------------------
AS
`

func TestLinuxFilesystemReport(t *testing.T) {
	threshold := 90.0
	settings := models.CheckSettings{
		Selected: map[string]bool{"backups": true, "failed_jobs": true, "filesystem": true, "missing_email": true},
		Resources: map[string]map[string]models.ResourceRule{"filesystem": {
			"/data":        {MaxUsedPercent: &threshold},
			"/mnt/backups": {MaxUsedPercent: &threshold},
			"/run":         {Ignore: true},
		}},
	}
	spaced := strings.ReplaceAll(linuxReport, "                   ", "\u00a0\t")
	for _, body := range []string{linuxReport, strings.ReplaceAll(spaced, "\n", "\r\n")} {
		p := ParseReport(body, settings)
		if !p.Complete || p.Hostname != "EXAMPLE-HOST" {
			t.Fatalf("metadata: %+v", p)
		}
		want := map[string]string{"/dev/shm": "not_evaluated", "/run": "ignored", "/data": "passed", "/mnt/backups": "failed"}
		failures := 0
		for _, r := range p.Results {
			if r.CheckType == "missing_email" {
				t.Fatal("valid Linux report classified as missing")
			}
			if r.CheckType == "failed_jobs" && r.Status != "unknown" {
				t.Fatal("unsupported failed-job rules must remain unknown")
			}
			if r.CheckType != "filesystem" {
				continue
			}
			if r.Status != want[r.Resource] {
				t.Fatalf("unexpected result: %+v", r)
			}
			delete(want, r.Resource)
			if r.Status == "failed" {
				failures++
			}
			if strings.Contains(r.Evidence, "dcx.ksh") {
				t.Fatal("footer included in filesystem evidence")
			}
		}
		if len(want) != 0 || failures != 1 {
			t.Fatalf("missing=%v failures=%d", want, failures)
		}
	}
	settings.Selected["filesystem"] = false
	p := ParseReport(linuxReport, settings)
	for _, r := range p.Results {
		if r.CheckType == "filesystem" && r.Status != "not_evaluated" {
			t.Fatal(r)
		}
	}
}

func TestLinuxIncompleteReportStillRejected(t *testing.T) {
	settings := models.CheckSettings{Selected: map[string]bool{"backups": true, "missing_email": true}}
	for _, body := range []string{
		linuxReport[:strings.Index(linuxReport, "Script :")],
		strings.Replace(linuxReport, "Backups=OK!", "", 1),
		"ORA-06550\nPLS-00201\nFilesystem Mounted Use%\ntmpfs /run 3%\nScript : /checks/dcx.ksh\nRun by : oracle@EXAMPLE-HOST",
	} {
		p := ParseReport(body, settings)
		if p.Complete || len(p.Results) != 1 || p.Results[0].CheckType != "missing_email" || p.Results[0].Status != "failed" {
			t.Fatalf("%+v", p)
		}
	}
}

func TestWindowsFilesystemStillSupported(t *testing.T) {
	threshold := 90.0
	p := ParseReport(sample("Backups=NOT_US\nFilesystem Usage\nC: 90% - 10Gb free\nD: 92% - 8Gb free"), models.CheckSettings{
		Selected:  map[string]bool{"backups": true, "filesystem": true, "missing_email": true},
		Resources: map[string]map[string]models.ResourceRule{"filesystem": {"C:": {MaxUsedPercent: &threshold}, "D:": {MaxUsedPercent: &threshold}}},
	})
	if !p.Complete || len(p.Results) != 3 || p.Results[0].Status != "not_managed" || p.Results[1].Status != "passed" || p.Results[2].Status != "failed" {
		t.Fatalf("%+v", p)
	}
}
