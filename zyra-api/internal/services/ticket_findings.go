package services

import (
	"fmt"
	"strings"
	"zyra-api/internal/models"
)

func findingsComment(results []models.CheckResult) string {
	if len(results) == 0 {
		return "The assessment reported an issue, but no structured findings were available."
	}
	lines := []string{fmt.Sprintf("%s check failed.", displayCheckType(results[0].CheckType))}
	for _, result := range results {
		parts := []string{}
		if result.Resource != "" {
			parts = append(parts, "Resource: "+result.Resource)
		}
		if result.Summary != "" {
			parts = append(parts, "Finding: "+result.Summary)
		}
		if result.Value != nil {
			label := "Reported value"
			if result.CheckType == "filesystem" {
				label = "Reported usage"
			} else if result.CheckType == "tablespace" || result.CheckType == "asm_space" {
				label = "Reported free space"
			}
			parts = append(parts, fmt.Sprintf("%s: %g%%", label, *result.Value))
		}
		if result.Threshold != nil {
			label := "Configured threshold"
			if result.CheckType == "filesystem" {
				label = "Maximum configured usage"
			} else if result.CheckType == "tablespace" || result.CheckType == "asm_space" {
				label = "Minimum configured free space"
			}
			parts = append(parts, fmt.Sprintf("%s: %g%%", label, *result.Threshold))
		}
		if evidence := strings.TrimSpace(result.Evidence); evidence != "" {
			if result.CheckType == "missing_email" && result.Summary == "Received but incomplete" {
				parts = append(parts, "The received report is available in the Raw Email tab")
			} else {
				const maxFindingEvidence = 4000
				if len(evidence) > maxFindingEvidence {
					evidence = evidence[:maxFindingEvidence] + "… (truncated; see Raw Email)"
				}
				parts = append(parts, "Reported evidence: "+evidence)
			}
		}
		lines = append(lines, "- "+strings.Join(parts, "; "))
	}
	return strings.Join(lines, "\n")
}

func displayCheckType(value string) string {
	labels := map[string]string{
		"archive_destinations": "Archive destinations",
		"asm_space":            "ASM space",
		"backups":              "Backups",
		"failed_jobs":          "Failed jobs",
		"filesystem":           "Filesystem",
		"fra":                  "Recovery area space",
		"missing_email":        "Missing email",
		"tablespace":           "Tablespace",
	}
	if label := labels[value]; label != "" {
		return label
	}
	return strings.ReplaceAll(value, "_", " ")
}

func groupedFailures(results []models.CheckResult) [][]models.CheckResult {
	byType := map[string][]models.CheckResult{}
	order := []string{}
	for _, result := range results {
		if result.Status != "failed" {
			continue
		}
		if _, exists := byType[result.CheckType]; !exists {
			order = append(order, result.CheckType)
		}
		byType[result.CheckType] = append(byType[result.CheckType], result)
	}
	groups := make([][]models.CheckResult, 0, len(order))
	for _, checkType := range order {
		groups = append(groups, byType[checkType])
	}
	return groups
}
