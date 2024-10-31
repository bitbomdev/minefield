package ingest

import "github.com/package-url/packageurl-go"

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

	pkg = "pkg:"
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

// PackageInfo is the specific package information
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

// PURLToPackage converts a Package URL string to PackageInfo
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
