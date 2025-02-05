package utils

import (
	"strings"
)

func SplitVersion(version string) []string {
	versionSlice := strings.Split(version, ".")

	return versionSlice
}

func CompareVersions(installedVersion string, latestVersion string) string {
	if strings.Contains(installedVersion, ".") && strings.Contains(latestVersion, ".") {
		installedVersionSlice := SplitVersion(installedVersion)
		latestVersionSlice := SplitVersion(latestVersion)

		major_diff := strings.Compare(installedVersionSlice[0], latestVersionSlice[0])
		minor_diff := strings.Compare(installedVersionSlice[1], latestVersionSlice[1])

		if major_diff == 0 && minor_diff == 0 {
			return "success"
		}

		if major_diff == 0 && minor_diff < 0 {
			return "warning"
		}

		return "danger"
	}

	if installedVersion == latestVersion {
		return "success"
	}

	return "danger"
}
