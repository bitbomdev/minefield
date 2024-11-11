package ingest

import (
	"encoding/json"
	"fmt"
	"github.com/bitbomdev/minefield/pkg/tools"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/bitbomdev/minefield/pkg/graph"
)

type Vulnerability struct {
	SchemaVersion    string                 `json:"schema_version"`
	ID               string                 `json:"id"`
	Modified         string                 `json:"modified"`
	Published        string                 `json:"published"`
	Withdrawn        string                 `json:"withdrawn"`
	Aliases          []string               `json:"aliases"`
	Related          []string               `json:"related"`
	Summary          string                 `json:"summary"`
	Details          string                 `json:"details"`
	Severity         []Severity             `json:"severity"`
	Affected         []Affected             `json:"affected"`
	References       []Reference            `json:"references"`
	Credits          []Credit               `json:"credits"`
	DatabaseSpecific map[string]interface{} `json:"database_specific"`
}

type Severity struct {
	Type  string `json:"type"`
	Score string `json:"score"`
}

type Affected struct {
	Package           Package                `json:"package"`
	Severity          []Severity             `json:"severity"`
	Ranges            []Range                `json:"ranges"`
	Versions          []string               `json:"versions"`
	EcosystemSpecific map[string]interface{} `json:"ecosystem_specific"`
	DatabaseSpecific  map[string]interface{} `json:"database_specific"`
}

type Package struct {
	Ecosystem string `json:"ecosystem"`
	Name      string `json:"name"`
	Purl      string `json:"purl"`
}

type Range struct {
	Type             string                 `json:"type"`
	Repo             string                 `json:"repo"`
	Events           []Event                `json:"events"`
	DatabaseSpecific map[string]interface{} `json:"database_specific"`
}

type Event struct {
	Introduced   string `json:"introduced"`
	Fixed        string `json:"fixed"`
	LastAffected string `json:"last_affected"`
	Limit        string `json:"limit"`
}

type Reference struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

type Credit struct {
	Name    string   `json:"name"`
	Contact []string `json:"contact"`
	Type    string   `json:"type"`
}

// Vulnerabilities ingests vulnerabilities from redis into the graph
func Vulnerabilities(storage graph.Storage, progress func(count int, id string)) error {
	keys, err := storage.GetAllKeys()
	if err != nil {
		return err
	}

	nodes, err := storage.GetNodes(keys)
	if err != nil {
		return fmt.Errorf("failed to get nodes from storage: %w", err)
	}
	count := 0

	for _, node := range nodes {
		if node.Type == tools.LibraryType && strings.HasPrefix(node.Name, pkg) {
			pkgInfo, err := PURLToPackage(node.Name)
			if err != nil {
				continue
			}
			vulnsData, err := storage.GetCustomData(tools.VulnerabilityType, pkgInfo.Name)
			if err != nil {
				return fmt.Errorf("failed to get vulnerabilityType data from storage: %w", err)
			}
			if len(vulnsData) == 0 {
				continue
			}
			for vulnID, vulndata := range vulnsData {
				// We are using the vuln ID from the map instead of the vulnID from the vuln data, since the map key could be an alias of a vulnerabilityType ID
				var vuln Vulnerability

				if err := json.Unmarshal(vulndata, &vuln); err != nil {
					return fmt.Errorf("failed to unmarshal vulnerabilityType data: %w", err)
				}

				if isPackageAffected(vuln, pkgInfo) {
					vulnNode, err := graph.AddNode(storage, tools.VulnerabilityType, vuln, vulnID)
					if err != nil {
						return fmt.Errorf("failed to add vulnerabilityType node to storage: %w", err)
					}

					if err := node.SetDependency(storage, vulnNode); err != nil {
						return fmt.Errorf("failed to add dependency edge to vulnerabilityType node: %w", err)
					}

					count++
					if progress != nil {
						progress(count, vulnID)
					}
				}
			}
		}
	}
	return nil
}

// LoadVulnerabilities processes the vulnerabilityType data and adds it to the storage.
func LoadVulnerabilities(storage graph.Storage, data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("data is empty")
	}

	vuln := Vulnerability{}
	if err := json.Unmarshal(data, &vuln); err != nil {
		return fmt.Errorf("failed to unmarshal vulnerabilityType data: %w", err)
	}

	errors := []error{}

	vulnData, err := json.Marshal(vuln)
	if err != nil {
		return fmt.Errorf("failed to marshal vulnerabilityType data: %w", err)
	}

	for _, affected := range vuln.Affected {
		if err := storage.AddOrUpdateCustomData(tools.VulnerabilityType, affected.Package.Name, vuln.ID, vulnData); err != nil {
			errors = append(errors, fmt.Errorf("failed to add vulnerabilityType to storage: %w", err))
		}

		for _, alias := range vuln.Aliases {
			if err := storage.AddOrUpdateCustomData(tools.VulnerabilityType, alias, vuln.ID, vulnData); err != nil {
				errors = append(errors, fmt.Errorf("failed to add vulnerabilityType alias to storage: %w", err))
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors occurred during vulnerabilities ingestion: %v", errors)
	}

	return nil
}

// isPackageAffected checks if the package is affected by the vulnerabilityType.
func isPackageAffected(vuln Vulnerability, pkgInfo PackageInfo) bool {
	for _, affected := range vuln.Affected {
		if affected.Package.Name != pkgInfo.Name || affected.Package.Ecosystem != pkgInfo.Ecosystem {
			continue
		}

		if isVersionIncluded(pkgInfo.Version, affected.Versions) {
			return true
		}

		if isVersionInRanges(pkgInfo.Version, affected.Ranges, affected.Package.Ecosystem) {
			return true
		}
	}
	return false
}

func isVersionIncluded(version string, versions []string) bool {
	for _, v := range versions {
		if v == version {
			return true
		}
	}
	return false
}

func isVersionInRanges(version string, ranges []Range, ecosystem string) bool {
	for _, r := range ranges {
		vulnerable := false
		sortedEvents := sortRangeEvents(r.Events, r.Type, ecosystem)
		for _, evt := range sortedEvents {
			switch {
			case evt.Introduced != "" && compareVersions(version, evt.Introduced, r.Type, ecosystem) >= 0:
				vulnerable = true
			case evt.Fixed != "" && compareVersions(version, evt.Fixed, r.Type, ecosystem) >= 0:
				vulnerable = false
			case evt.LastAffected != "" && compareVersions(version, evt.LastAffected, r.Type, ecosystem) > 0:
				vulnerable = false
			}
		}

		if vulnerable {
			return true
		}
	}
	return false
}

func sortRangeEvents(events []Event, eventType string, ecosystem string) []Event {
	sortedEvents := make([]Event, len(events))
	copy(sortedEvents, events)

	lessFunc := func(i, j int) bool {
		vi := getVersionFromEvent(events[i])
		vj := getVersionFromEvent(events[j])
		return compareVersions(vi, vj, eventType, ecosystem) < 0
	}

	sort.Slice(sortedEvents, lessFunc)
	return sortedEvents
}

func getVersionFromEvent(evt Event) string {
	if evt.Introduced != "" {
		return evt.Introduced
	}
	if evt.Fixed != "" {
		return evt.Fixed
	}
	if evt.LastAffected != "" {
		return evt.LastAffected
	}
	return ""
}

func compareVersions(v1, v2, eventType, ecosystem string) int {
	const (
		EventTypeSEMVER    = "SEMVER"
		EventTypeECOSYSTEM = "ECOSYSTEM"
		EventTypeGIT       = "GIT"
	)
	switch eventType {
	case EventTypeSEMVER:
		ver1, err1 := semver.NewVersion(v1)
		ver2, err2 := semver.NewVersion(v2)
		if err1 != nil || err2 != nil {
			return strings.Compare(v1, v2)
		}
		return ver1.Compare(ver2)
	case EventTypeECOSYSTEM:
		return compareEcosystemVersions(v1, v2, ecosystem)
	case EventTypeGIT:
		return strings.Compare(v1, v2)
	default:
		return strings.Compare(v1, v2)
	}
}

func compareEcosystemVersions(v1, v2, ecosystem string) int {
	// Implement ecosystem-specific version comparison logic here.
	// Placeholder implementation:
	return strings.Compare(v1, v2)
}
