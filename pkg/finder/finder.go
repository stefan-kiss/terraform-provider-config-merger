package finder

import (
	"github.com/stefan-kiss/terraform-provider-config-merger/pkg/envfacts"
	"path/filepath"
)

// FindConfigFiles finds all files named config.yaml that are found in the root.
func FindConfigFiles(p envfacts.ProjectStructure, fileGlobs []string) (fileList []string, err error) {
	fileList = make([]string, 0)

	for _, v := range append([]envfacts.VarMapping{p.Root}, p.Vars...) {
		dirList, err := MatchGlobs(fileGlobs, v.RealPath)
		if err != nil {
			return nil, err
		}
		fileList = append(fileList, dirList...)
	}
	return fileList, nil
}

// MatchGlobs finds any file paths that matches any of the list of globs in `dirPath`.
// The glob patterns are formed by joining `dirPath` with each of the globs in `fileGlobs`.
func MatchGlobs(fileGlobs []string, dirPath string) (matches []string, err error) {
	matches = make([]string, 0)
	for _, fileGlob := range fileGlobs {
		base := filepath.Base(fileGlob)
		results, err := filepath.Glob(filepath.Join(dirPath, base))
		if err != nil {
			return matches, err
		}
		matches = append(matches, results...)
	}
	return matches, nil
}
