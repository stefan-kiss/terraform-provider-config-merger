package finder

import (
	"github.com/stefan-kiss/terraform-provider-config-merger/pkg/envfacts"
	"os"
	"path/filepath"
)

// FindConfigFiles finds all files named config.yaml that are found in the root.
func FindConfigFiles(p envfacts.ProjectStructure) (fileList []string, err error) {
	fileList = make([]string, 0)

	for _, v := range append([]envfacts.VarMapping{p.Root}, p.Vars...) {
		checkPath := filepath.Join(v.RealPath, "config.yaml")
		if FileExists(checkPath) {
			fileList = append(fileList, checkPath)
		}
	}
	return fileList, nil
}

// FileExists checks if the given file exists.
func FileExists(filePath string) bool {
	if fileInfo, err := os.Stat(filePath); err != nil || fileInfo.IsDir() {
		return false
	}
	return true
}
