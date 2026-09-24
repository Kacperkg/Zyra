package services

import (
	"strings"
	"testing"
	"zyra-api/internal/models"
)

func TestFindingsCommentIncludesAssessmentTimeDetails(t *testing.T) {
	usage := 92.0
	threshold := 90.0
	results := []models.CheckResult{
		{CheckType: "filesystem", Resource: "/mnt/oracle", Status: "failed", Summary: "breached threshold", Evidence: "/dev/mapper/oracle /mnt/oracle 92%", Value: &usage, Threshold: &threshold},
		{CheckType: "filesystem", Resource: "/archive", Status: "failed", Summary: "breached threshold", Evidence: "/dev/mapper/archive /archive 95%", Value: floatPointer(95), Threshold: &threshold},
	}

	groups := groupedFailures(results)
	if len(groups) != 1 || len(groups[0]) != 2 {
		t.Fatalf("expected one grouped filesystem ticket, got %#v", groups)
	}
	comment := findingsComment(groups[0])
	for _, want := range []string{"Filesystem check failed", "/mnt/oracle", "Reported usage: 92%", "Maximum configured usage: 90%", "/archive"} {
		if !strings.Contains(comment, want) {
			t.Fatalf("comment %q does not contain %q", comment, want)
		}
	}
}

func floatPointer(value float64) *float64 { return &value }
