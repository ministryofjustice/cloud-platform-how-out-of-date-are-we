package utils

import "strings"

func SplitVersion(version string) []string {
	versionSlice := strings.Split(version, ".")

	return versionSlice
}

func CompareVersions(installedVersion string, latestVersion string) string {
	installedVersionSlice := SplitVersion(installedVersion)
	latestVersionSlice := SplitVersion(latestVersion)

	var state string

	major_diff := strings.Compare(installedVersionSlice[0], latestVersionSlice[0])
	minor_diff := strings.Compare(installedVersionSlice[1], latestVersionSlice[1])

	if major_diff == 0 && minor_diff == 0 {
		state = "SUCCESS"
	}

	if major_diff < 0 {
		state = "DANGER"
	}

	if major_diff == 0 && minor_diff < 0 {
		state = "WARNING"
	}

	return state
}
