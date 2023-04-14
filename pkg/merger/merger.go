package merger

import (
	"github.com/cppforlife/go-patch/patch"
	"github.com/geofffranks/simpleyaml"
	"github.com/geofffranks/spruce"
	"github.com/geofffranks/yaml"
	log "github.com/sirupsen/logrus"
	"github.com/starkandwayne/goutils/ansi"
	"github.com/voxelbrain/goptions"
	"io"
	"os"
)

type RootIsArrayError struct {
	msg string
}

func (r RootIsArrayError) Error() string {
	return r.msg
}

type YamlFile struct {
	Path   string
	Reader io.ReadCloser
}

//func loadYamlFile(file string) (YamlFile, error) {
//	var target YamlFile
//	if file == "-" {
//		target = YamlFile{Reader: os.Stdin, Path: "-"}
//	} else {
//		f, err := os.Open(file)
//		if err != nil {
//			return YamlFile{}, ansi.Errorf("@R{Error reading file} @m{%s}: %s", file, err.Error())
//		}
//		target = YamlFile{Path: file, Reader: f}
//	}
//	return target, nil
//}

type MergeOpts struct {
	SkipEval       bool               `goptions:"--skip-eval, description='Do not evaluate spruce logic after merging docs'"`
	Prune          []string           `goptions:"--prune, description='Specify keys to prune from final output (may be specified more than once)'"`
	CherryPick     []string           `goptions:"--cherry-pick, description='The opposite of prune, specify keys to cherry-pick from final output (may be specified more than once)'"`
	FallbackAppend bool               `goptions:"--fallback-append, description='Default merge normally tries to key merge, then inline. This flag says do an append instead of an inline.'"`
	EnableGoPatch  bool               `goptions:"--go-patch, description='Enable the use of go-patch when parsing files to be merged'"`
	MultiDoc       bool               `goptions:"--multi-doc, -m, description='Treat multi-doc yaml as multiple files.'"`
	Help           bool               `goptions:"--help, -h"`
	Files          goptions.Remainder `goptions:"description='List of files to merge. To read STDIN, specify a filename of \\'-\\'.'"`
}

func isArrayError(err error) bool {
	_, ok := err.(RootIsArrayError)
	return ok
}

func readFile(file *YamlFile) ([]byte, error) {
	var data []byte
	var err error

	if file.Path == "-" {
		file.Path = "STDIN"
		stat, err := os.Stdin.Stat()
		if err != nil {
			return nil, ansi.Errorf("@R{Error statting STDIN} - Bailing out: %s\n", err.Error())
		}
		if stat.Mode()&os.ModeCharDevice == 0 {
			data, err = io.ReadAll(os.Stdin)
			if err != nil {
				return nil, ansi.Errorf("@R{Error reading file} @m{%s}: %s\n", file.Path, err.Error())
			}
		}
	} else {
		data, err = io.ReadAll(file.Reader)
		if err != nil {
			return nil, ansi.Errorf("@R{Error reading file} @m{%s}: %s\n", file.Path, err.Error())
		}
	}
	if len(data) == 0 && file.Path == "STDIN" {
		return nil, ansi.Errorf("@R{Error reading STDIN}: no data found. Did you forget to pipe data to STDIN, or specify yaml files to merge?")
	}

	return data, nil
}

func parseYAML(data []byte) (map[interface{}]interface{}, error) {
	y, err := simpleyaml.NewYaml(data)
	if err != nil {
		return nil, err
	}

	if empty_y, _ := simpleyaml.NewYaml([]byte{}); *y == *empty_y {
		log.Debugf("YAML doc is empty, creating empty hash/map")
		return make(map[interface{}]interface{}), nil
	}

	doc, err := y.Map()

	if err != nil {
		if _, arrayErr := y.Array(); arrayErr == nil {
			return nil, RootIsArrayError{msg: ansi.Sprintf("@R{Root of YAML document is not a hash/map}: %s\n", err)}
		}
		return nil, ansi.Errorf("@R{Root of YAML document is not a hash/map}: %s\n", err.Error())
	}

	return doc, nil
}

func parseGoPatch(data []byte) (patch.Ops, error) {
	opdefs := []patch.OpDefinition{}
	err := yaml.Unmarshal(data, &opdefs)
	if err != nil {
		return nil, ansi.Errorf("@R{Root of YAML document is not a hash/map. Tried parsing it as go-patch, but got}: %s\n", err)
	}
	ops, err := patch.NewOpsFromDefinitions(opdefs)
	if err != nil {
		return nil, ansi.Errorf("@R{Unable to parse go-patch definitions: %s\n", err)
	}
	return ops, nil
}

func MergeAllDocs(files []YamlFile, options MergeOpts) (*spruce.Evaluator, error) {
	m := &spruce.Merger{AppendByDefault: options.FallbackAppend}
	root := make(map[interface{}]interface{})

	for _, file := range files {
		log.Debugf("Processing file '%s'", file.Path)

		data, err := readFile(&file)
		if err != nil {
			return nil, err
		}

		doc, err := parseYAML(data)
		if err != nil {
			if isArrayError(err) && options.EnableGoPatch {
				log.Debugf("Detected root of document as an array. Attempting go-patch parsing")
				ops, err := parseGoPatch(data)
				if err != nil {
					return nil, ansi.Errorf("@m{%s}: @R{%s}\n", file.Path, err.Error())
				}
				newObj, err := ops.Apply(root)
				if err != nil {
					return nil, ansi.Errorf("@m{%s}: @R{%s}\n", file.Path, err.Error())
				}
				if newRoot, ok := newObj.(map[interface{}]interface{}); !ok {
					return nil, ansi.Errorf("@m{%s}: @R{Unable to convert go-patch output into a hash/map for further merging|\n", file.Path)
				} else {
					root = newRoot
				}
			} else {
				return nil, ansi.Errorf("@m{%s}: @R{%s}\n", file.Path, err.Error())
			}
		} else {
			// this is ignored in spruce original code also. TBD if we need to treat it as an error.
			_ = m.Merge(root, doc)

		}
		tmpYaml, _ := yaml.Marshal(root) // we don't care about errors for debugging
		log.Debugf("Current data after processing '%s':\n%s", file.Path, tmpYaml)
	}

	if m.Error() != nil {
		return nil, m.Error()
	}

	ev := &spruce.Evaluator{Tree: root, SkipEval: options.SkipEval}
	err := ev.Run(options.Prune, options.CherryPick)
	return ev, err
}
