package envfacts

import (
	"github.com/go-test/deep"
	"github.com/k0kubun/pp"
	"reflect"
	"testing"
	_ "unsafe"
)

func TestParseProjectStructure(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		wantP   ProjectStructure
		wantErr bool
	}{
		{
			name:    "EmptyString",
			args:    args{s: ""},
			wantP:   ProjectStructure{},
			wantErr: true,
		},
		{
			name: "CorrectlyFormatted",
			args: args{s: "config/{{environment}}/{{global.region}}/{{project}}"},
			wantP: ProjectStructure{
				Root: VarMapping{
					VariableValue: "config",
					RealPath:      "",
				},
				Vars: []VarMapping{
					{
						VariableName: "environment",
					},
					{
						VariableName: "global.region",
					},
					{
						VariableName: "project",
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotP, err := ParseProjectStructure(tt.args.s)
			if (err != nil) != tt.wantErr {
				pp.Printf("ParseProjectStructure() gotP = %v", gotP)
				t.Errorf("ParseProjectStructure() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotP, tt.wantP) {
				t.Errorf("ParseProjectStructure() gotP = %v, want %v", gotP, tt.wantP)
			}
		})
	}
}

func TestExtractVar(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "emptyString",
			args: args{
				s: "",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "NoSpaces",
			args: args{
				s: "{{environment}}",
			},
			want:    "environment",
			wantErr: false,
		},
		{
			name: "Spaces",
			args: args{
				s: "{{  environment }}",
			},
			want:    "environment",
			wantErr: false,
		},
		{
			name: "WrongFormatLeadCharacter",
			args: args{
				s: " {{environment}}",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "Dots",
			args: args{
				s: "{{global.environment}}",
			},
			want:    "global.environment",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractVar(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractVar() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ExtractVar() got = %v, want %v", got, tt.want)
			}
		})
	}
}

// HomeDirTesting returns a static home directory for testing
func HomeDirTesting() (string, error) {
	return "/home/test", nil
}

func TestGetAbsPath(t *testing.T) {
	type args struct {
		inputPath   string
		homeDirFunc func() (string, error)
	}
	tests := []struct {
		name        string
		args        args
		wantAbsPath string
		wantErr     bool
	}{
		{
			name: "HomeDir",
			args: args{
				inputPath:   "~/config/development/us-east-2/s3bucket",
				homeDirFunc: HomeDirTesting,
			},
			wantAbsPath: "/home/test/config/development/us-east-2/s3bucket",
			wantErr:     false,
		},
		{
			name: "RelativePath",
			args: args{
				inputPath:   "config/development/us-east-2/s3bucket",
				homeDirFunc: HomeDirTesting,
			},
			wantAbsPath: GetFileDir() + "/config/development/us-east-2/s3bucket",
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAbsPath, err := GetAbsPath(tt.args.inputPath, tt.args.homeDirFunc)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAbsPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotAbsPath != tt.wantAbsPath {
				t.Errorf("GetAbsPath() gotAbsPath = %v, want %v", gotAbsPath, tt.wantAbsPath)
			}
		})
	}
}

func TestProjectStructure_MapPathToProject(t *testing.T) {
	type fields struct {
		Root VarMapping
		Vars []VarMapping
	}
	type args struct {
		projectPath string
		homeDirFunc func() (string, error)
	}
	tests := []struct {
		name                      string
		fields                    fields
		args                      args
		resultingProjectStructure ProjectStructure
		wantErr                   bool
	}{
		{
			name: "Base",
			fields: fields{
				Root: VarMapping{
					VariableName:  "",
					VariableValue: "config",
					RealPath:      "",
				},
				Vars: []VarMapping{
					{
						VariableName:  "environment",
						VariableValue: "",
						RealPath:      "",
					},
					{
						VariableName:  "global.region",
						VariableValue: "",
						RealPath:      "",
					},
					{
						VariableName:  "project",
						VariableValue: "",
						RealPath:      "",
					},
				},
			},
			args: args{
				projectPath: "config/development/us-east-2/s3bucket",
				homeDirFunc: HomeDirTesting,
			},
			resultingProjectStructure: ProjectStructure{
				Root: VarMapping{
					VariableName:  "",
					VariableValue: "config",
					RealPath:      GetFileDir() + "/config",
				},
				Vars: []VarMapping{
					{
						VariableName:  "environment",
						VariableValue: "development",
						RealPath:      GetFileDir() + "/config/development",
					},
					{
						VariableName:  "global.region",
						VariableValue: "us-east-2",
						RealPath:      GetFileDir() + "/config/development/us-east-2",
					},
					{
						VariableName:  "project",
						VariableValue: "s3bucket",
						RealPath:      GetFileDir() + "/config/development/us-east-2/s3bucket",
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := ProjectStructure{
				Root: tt.fields.Root,
				Vars: tt.fields.Vars,
			}
			if err := p.MapPathToProject(tt.args.projectPath, tt.args.homeDirFunc); (err != nil) != tt.wantErr {
				t.Errorf("MapPathToProject() error = %v, wantErr %v", err, tt.wantErr)
			}
			if diff := deep.Equal(p, tt.resultingProjectStructure); diff != nil {
				for _, d := range diff {
					t.Logf("MapPathToProject() differences betwween want and got: %v", d)
				}
				t.Fail()
			}
		})
	}
}
