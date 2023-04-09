package envfacts

import (
	"fmt"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

type ProjectStructure struct {
	Root VarMapping
	//Vars []string
	Vars []VarMapping
}

type VarMapping struct {
	VariableName  string
	VariableValue string
	RealPath      string
}

// ExtractVar extracts the string between double brackets from the given string, trimming whitespaces
func ExtractVar(s string) (string, error) {
	vars := strings.Split(s, "{{")
	if len(vars) < 2 || vars[0] != "" {
		return "", fmt.Errorf("variable needs to be wrapped in double brackets, with no leading characters: %q", s)
	}
	vars = strings.Split(vars[1], "}}")
	if len(vars) < 2 || vars[1] != "" {
		return "", fmt.Errorf("variable needs to be wrapped in double brackets, with no leading characters: %q", s)
	}
	return strings.TrimSpace(vars[0]), nil
}

// ParseProjectStructure parses the project structure from the given string.
func ParseProjectStructure(s string) (p ProjectStructure, err error) {
	vars := strings.Split(s, string(filepath.Separator))

	if len(vars) < 1 || vars[0] == "" {
		return p, fmt.Errorf("project structure needs to have at least the root directory: %q", s)
	}
	p.Root = VarMapping{
		VariableValue: vars[0],
		RealPath:      "",
	}
	p.Vars = make([]VarMapping, len(vars)-1)
	for idx, v := range vars[1:] {
		extracted, err := ExtractVar(v)
		if err != nil {
			return p, err
		}
		p.Vars[idx] = VarMapping{
			VariableName: extracted,
		}
	}
	return p, nil
}

// GetAbsPath returns the absolute path, while also doing home directory replacement
func GetAbsPath(inputPath string, homeDirFunc func() (string, error)) (absPath string, err error) {
	cleanPath := filepath.Clean(inputPath)
	switch cleanPath[0] {
	case '~':
		homeDir, err := homeDirFunc()
		if err != nil {
			return "", err
		}
		absPath = strings.Replace(cleanPath, "~", homeDir, 1)
	case '/':
		return cleanPath, err
	default:
		absPath, err = filepath.Abs(cleanPath)
		if err != nil {
			return "", err
		}
	}
	return absPath, nil
}

// MapPathToProject maps the given path to the project structure.
func (p *ProjectStructure) MapPathToProject(projectPath string, homeDirFunc func() (string, error)) (err error) {
	absPath, err := GetAbsPath(projectPath, homeDirFunc)
	dirs := strings.Split(absPath, string(filepath.Separator))
	dirs[0] = string(filepath.Separator) + dirs[0]
	rootIdx := 0
	found := false

	for i := len(dirs) - 1; i >= 0; i-- {
		if dirs[i] == p.Root.VariableValue {
			if len(dirs[i:]) != len(p.Vars)+1 {
				return fmt.Errorf("projectPath %q does not match project structure %q", projectPath, p)
			}
			found = true
			rootIdx = i
			break
		}
	}
	if !found {
		return fmt.Errorf("projectPath %q does not match project structure %q", projectPath, p)
	}
	p.Root.RealPath = path.Join(dirs[:rootIdx+1]...)
	varStartIdx := rootIdx + 1
	for i := varStartIdx; i < len(dirs); i++ {
		p.Vars[i-varStartIdx].RealPath = path.Join(dirs[:i+1]...)
		p.Vars[i-varStartIdx].VariableValue = dirs[i]
	}
	return nil
}

// GetFileDir returns the directory for the current source file
func GetFileDir() string {
	_, filename, _, _ := runtime.Caller(1)
	return path.Dir(filename)
}
