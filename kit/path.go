package kit

import (
	"path/filepath"
	"strings"
)

var (
	assetLocations = []string{
		"templates/customers",
		"assets",
		"config",
		"layout",
		"snippets",
		"templates",
		"locales",
		"sections",
	}
)

func pathInProject(root, filename string) bool {
	return pathToProject(root, filename) != "" || isProjectDirectory(root, filename)
}

func isProjectDirectory(root, filename string) bool {
	filename = strings.TrimPrefix(
		filepath.ToSlash(filepath.Clean(filename)),
		filepath.ToSlash(filepath.Clean(root)+"/"),
	)

	for _, dir := range assetLocations {
		if dir == filename {
			return true
		}
	}

	return false
}

func pathToProject(root, filename string) string {
	filename = strings.TrimPrefix(
		filepath.ToSlash(filepath.Clean(filename)),
		filepath.ToSlash(filepath.Clean(root)+"/"),
	)

	for _, dir := range assetLocations {
		split := strings.SplitAfterN(filename, dir+"/", 2)
		if len(split) > 1 && strings.HasPrefix(filename, dir+"/") {
			return filepath.ToSlash(filepath.Join(dir, split[len(split)-1]))
		}
	}

	return ""
}
