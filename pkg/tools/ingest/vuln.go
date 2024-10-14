package ingest

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/bit-bom/minefield/pkg/graph"
	"github.com/package-url/packageurl-go"
)

const OSVTagName = "osv-vulns"

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
type Ecosystem string

const (
	EcosystemGo        Ecosystem = "Go"
	EcosystemNPM       Ecosystem = "npm"
	EcosystemOSSFuzz   Ecosystem = "OSS-Fuzz"
	EcosystemPyPI      Ecosystem = "PyPI"
	EcosystemRubyGems  Ecosystem = "RubyGems"
	EcosystemCratesIO  Ecosystem = "crates.io"
	EcosystemPackagist Ecosystem = "Packagist"
	EcosystemMaven     Ecosystem = "Maven"
	EcosystemNuGet     Ecosystem = "NuGet"
	EcosystemDebian    Ecosystem = "Debian"
	EcosystemAlpine    Ecosystem = "Alpine"
	EcosystemHex       Ecosystem = "Hex"
)

// used like so: purlEcosystems[PkgURL.Type][PkgURL.Namespace]
// * means it should match any namespace string
var purlEcosystems = map[string]map[string]Ecosystem{
	"apk":      {"alpine": EcosystemAlpine},
	"cargo":    {"*": EcosystemCratesIO},
	"deb":      {"debian": EcosystemDebian},
	"hex":      {"*": EcosystemHex},
	"golang":   {"*": EcosystemGo},
	"maven":    {"*": EcosystemMaven},
	"nuget":    {"*": EcosystemNuGet},
	"npm":      {"*": EcosystemNPM},
	"composer": {"*": EcosystemPackagist},
	"generic":  {"*": EcosystemOSSFuzz},
	"pypi":     {"*": EcosystemPyPI},
	"gem":      {"*": EcosystemRubyGems},
}

// Specific package information
type PackageInfo struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	Ecosystem string `json:"ecosystem"`
	Commit    string `json:"commit,omitempty"`
}

func getPURLEcosystem(pkgURL packageurl.PackageURL) Ecosystem {
	ecoMap, ok := purlEcosystems[pkgURL.Type]
	if !ok {
		return Ecosystem(pkgURL.Type + ":" + pkgURL.Namespace)
	}

	wildcardRes, hasWildcard := ecoMap["*"]
	if hasWildcard {
		return wildcardRes
	}

	ecosystem, ok := ecoMap[pkgURL.Namespace]
	if !ok {
		return Ecosystem(pkgURL.Type + ":" + pkgURL.Namespace)
	}

	return ecosystem
}

// PURLToPackage converts a Package URL string to models.PackageInfo
func PURLToPackage(purl string) (PackageInfo, error) {
	parsedPURL, err := packageurl.FromString(purl)
	if err != nil {
		return PackageInfo{}, err
	}
	ecosystem := getPURLEcosystem(parsedPURL)

	// PackageInfo expects the full namespace in the name for ecosystems that specify it.
	name := parsedPURL.Name
	if parsedPURL.Namespace != "" {
		switch ecosystem { //nolint:exhaustive
		case EcosystemMaven:
			// Maven uses : to separate namespace and package
			name = parsedPURL.Namespace + ":" + parsedPURL.Name
		case EcosystemDebian, EcosystemAlpine:
			// Debian and Alpine repeats their namespace in PURL, so don't add it to the name
			name = parsedPURL.Name
		default:
			name = parsedPURL.Namespace + "/" + parsedPURL.Name
		}
	}

	return PackageInfo{
		Name:      name,
		Ecosystem: string(ecosystem),
		Version:   parsedPURL.Version,
	}, nil
}

// Vulnerabilities ingests vulnerabilities from redis into the graph
func Vulnerabilities(storage graph.Storage, progress func(count int, id string)) error {
	const pkg = "pkg:"
	const library = "library"
	const vulnerability = "vuln"
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
		if node.Type == library && strings.HasPrefix(node.Name, pkg) {
			purl, err := PURLToPackage(node.Name)
			if err != nil {
				continue
			}
			vulnsData, err := storage.GetCustomData(OSVTagName, purl.Name)
			if err != nil {
				return fmt.Errorf("failed to get vulnerability data from storage: %w", err)
			}
			if len(vulnsData) == 0 {
				continue
			}
			for vulnID, vulndata := range vulnsData {
				// We are using the vuln ID from the map instead of the vulnID from the vuln data, since the map key could be an alias of a vulnerability ID
				var vuln Vulnerability

				if err := json.Unmarshal(vulndata, &vuln); err != nil {
					return fmt.Errorf("failed to unmarshal vulnerability data: %w", err)
				}

				pkgInfo, err := PURLToPackage(node.Name)
				if err != nil {
					return fmt.Errorf("failed to convert PURL to package info: %w", err)
				}

				if isPackageAffected(vuln, pkgInfo) {
					vulnNode, err := graph.AddNode(storage, vulnerability, vuln, vulnID)
					if err != nil {
						return fmt.Errorf("failed to add vulnerability node to storage: %w", err)
					}

					if err := node.SetDependency(storage, vulnNode); err != nil {
						return fmt.Errorf("failed to add dependency edge to vulnerability node: %w", err)
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

// LoadVulnerabilities processes the vulnerability data and adds it to the storage.
func LoadVulnerabilities(storage graph.Storage, data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("data is empty")
	}

	vuln := Vulnerability{}
	if err := json.Unmarshal(data, &vuln); err != nil {
		return fmt.Errorf("failed to unmarshal vulnerability data: %w", err)
	}

	errors := []error{}

	vulnData, err := json.Marshal(vuln)
	if err != nil {
		return fmt.Errorf("failed to marshal vulnerability data: %w", err)
	}

	for _, affected := range vuln.Affected {
		if err := storage.AddOrUpdateCustomData(OSVTagName, affected.Package.Name, vuln.ID, vulnData); err != nil {
			errors = append(errors, fmt.Errorf("failed to add vulnerability to storage: %w", err))
		}

		for _, alias := range vuln.Aliases {
			if err := storage.AddOrUpdateCustomData(OSVTagName, alias, vuln.ID, vulnData); err != nil {
				errors = append(errors, fmt.Errorf("failed to add vulnerability alias to storage: %w", err))
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors occurred during vulnerabilities ingestion: %v", errors)
	}

	return nil
}

// isPackageAffected checks if the package is affected by the vulnerability.
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
