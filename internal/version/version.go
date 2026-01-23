package version

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
	hv "github.com/hashicorp/go-version"
)

const (
	InitialVersion = "0.0.0"
)

// RemoteVersions fetches all available Go versions from golang.org/dl.
func RemoteVersions(golangUrl string) ([]*hv.Version, error) {
	res, err := http.Get(golangUrl + "/dl")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch versions from %s/dl: %w", golangUrl, err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	var versionsRaw []string
	doc.Find("a.download").Each(func(i int, s *goquery.Selection) {
		url, exists := s.Attr("href")
		if !exists || !strings.HasSuffix(url, "src.tar.gz") {
			return
		}
		reg := regexp.MustCompile(`/dl/go([0-9.]+)\.src\.tar\.gz$`)
		ver := reg.FindStringSubmatch(url)
		if len(ver) > 1 {
			versionsRaw = append(versionsRaw, ver[1])
		}
	})

	versions := make([]*hv.Version, 0, len(versionsRaw))
	for _, raw := range versionsRaw {
		v, err := hv.NewVersion(raw)
		if err != nil {
			continue
		}
		versions = append(versions, v)
	}

	sort.Sort(sort.Reverse(hv.Collection(versions)))

	return versions, nil
}

// RemoteLatestVersions returns the latest patch version for each minor version from remote.
func RemoteLatestVersions(golangUrl string) ([]*hv.Version, error) {
	versions, err := RemoteVersions(golangUrl)
	if err != nil {
		return nil, err
	}
	return LatestMinorVersions(versions), nil
}

// LatestMinorVersions returns the latest patch version for each minor version.
// Input must be sorted in descending order.
func LatestMinorVersions(versions []*hv.Version) []*hv.Version {
	latestMap := make(map[string]*hv.Version)
	for _, v := range versions {
		seg := v.Segments()
		minorKey := fmt.Sprintf("%d.%d", seg[0], seg[1])
		if _, exists := latestMap[minorKey]; exists {
			continue
		}
		latestMap[minorKey] = v
	}

	var result []*hv.Version
	for _, v := range latestMap {
		result = append(result, v)
	}

	sort.Sort(sort.Reverse(hv.Collection(result)))

	return result
}

// LocalVersions returns all installed Go versions from the goroots directory.
func LocalVersions(gorootsDir string) ([]*hv.Version, error) {
	files, err := os.ReadDir(gorootsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*hv.Version{}, nil
		}
		return nil, fmt.Errorf("failed to read directory %s: %w", gorootsDir, err)
	}

	versions := make([]*hv.Version, 0)
	for _, file := range files {
		if !file.IsDir() {
			continue
		}
		v, err := hv.NewVersion(file.Name())
		if err != nil {
			continue
		}
		versions = append(versions, v)
	}

	sort.Sort(sort.Reverse(hv.Collection(versions)))

	return versions, nil
}

// LatestVersion finds the latest patch version matching the given minor version.
func LatestVersion(ver string, latestVersions []*hv.Version) (string, error) {
	target, err := hv.NewVersion(ver)
	if err != nil {
		return "", fmt.Errorf("invalid version format: %s", ver)
	}
	seg := target.Segments()

	latest, err := hv.NewVersion(InitialVersion)
	if err != nil {
		return "", err
	}

	for _, v := range latestVersions {
		if latest.GreaterThan(v) {
			continue
		}
		segl := v.Segments()
		if seg[0] == segl[0] && seg[1] == segl[1] {
			latest = v
		}
	}

	if latest.Original() == InitialVersion {
		return "", fmt.Errorf("no matching version found for %s", ver)
	}

	return latest.Original(), nil
}
