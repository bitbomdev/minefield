package ingest

// Based on the osv-scanner's query code to OSV.dev

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/bit-bom/minefield/pkg"
	"github.com/package-url/packageurl-go"
)

type Vulnerability struct {
	ID string `json:"id"`
}

type Package struct {
	PURL      string `json:"purl,omitempty"`
	Name      string `json:"name,omitempty"`
	Ecosystem string `json:"ecosystem,omitempty"`
}

type Query struct {
	Commit  string  `json:"commit,omitempty"`
	Package Package `json:"package,omitempty"`
	Version string  `json:"version,omitempty"`
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
var ErrBadPurl = fmt.Errorf("bad purl")

func Vulnerabilities(storage pkg.Storage) error {
	keys, err := storage.GetAllKeys()
	if err != nil {
		return err
	}

	for _, key := range keys {
		node, err := storage.GetNode(key)
		if err != nil {
			return err
		}

		if node.Type == "PACKAGE" {
			if node.Name == "" {
				continue
			}
			vulns, err := queryOSV(node.Name)
			if err != nil {
				return err
			}

			for _, vuln := range vulns {
				vulnNode, err := pkg.AddNode(storage, "VULNERABILITY", any(vuln), vuln.ID)
				if err != nil {
					return err
				}

				if err := node.SetDependency(storage, vulnNode); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func getPURLEcosystem(pkgURL packageurl.PackageURL) (Ecosystem, error) {
	ecoMap, ok := purlEcosystems[pkgURL.Type]
	if !ok {
		return "", ErrBadPurl
	}

	wildcardRes, hasWildcard := ecoMap["*"]
	if hasWildcard {
		return wildcardRes, nil
	}

	ecosystem, ok := ecoMap[pkgURL.Namespace]
	if !ok {
		return "", ErrBadPurl
	}

	return ecosystem, nil
}

// PURLToPackageQuery converts a Package URL string to an OSV.dev query
func PURLToPackageQuery(purl string) (Query, error) {
	parsedPURL, err := packageurl.FromString(purl)
	if err != nil {
		return Query{}, err
	}
	ecosystem, err := getPURLEcosystem(parsedPURL)
	if err != nil {
		return Query{}, err
	}

	// Ensure the ecosystem is correctly set

	ecosystemStr := Ecosystem(strings.Trim(string(ecosystem), ":"))
	if ecosystemStr == "alpine" {
		ecosystemStr = "Alpine"
	}

	// PackageInfo expects the full namespace in the name for ecosystems that specify it.
	name := parsedPURL.Name
	if parsedPURL.Namespace != "" {
		switch ecosystemStr { //nolint:exhaustive
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

	return Query{
		Version: parsedPURL.Version,
		Package: Package{
			Name:      name,
			Ecosystem: string(ecosystemStr),
		},
	}, nil
}

func queryOSV(purl string) ([]Vulnerability, error) {
	query, err := PURLToPackageQuery(purl)
	if errors.Is(err, ErrBadPurl) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	queryBytes, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	requestBuf := bytes.NewBuffer(queryBytes)
	req, err := http.NewRequest(http.MethodPost, "https://api.osv.dev/v1/query", requestBuf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to query OSV, query = %s, purl = %s : %w", string(queryBytes), purl, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OSV query failed with status, query = %s, purl = %s : %s", string(queryBytes), purl, resp.Status)
	}

	var result struct {
		Vulns []Vulnerability `json:"vulns"`
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to decode OSV response: %w", err)
	}

	return result.Vulns, nil
}
