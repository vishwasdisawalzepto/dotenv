package dotenv

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

const defaultConfigFile = ".env"

type options struct {
	lookupMod bool // look up for go.mod file, by default false
	lookupGit bool // look up for .git directory, by default false

	lookupFile  []string // file type of .env file, by default .env, ex: .env.test
	lookupPaths []string // look up for .env file in these paths, by default the current directory

	lookupWatchFile []string

	disableFileExpand bool // disable expanding lookupFile to find .env.${ENVIRONMENT} files, by default false
	disablePathExpand bool // disable expanding lookupPaths to find .env file, by default false
	debug             bool // enable debug mode, by default false
	watchConfig       bool // enable watch config mode, by default false
	onConfigChange    func(fsnotify.Event)
}

func (o *options) FilesOrDefault() []string {
	if len(o.lookupFile) == 0 {
		return []string{defaultConfigFile}
	}
	return append(o.lookupFile, o.lookupWatchFile...)
}

// ParseFilePaths parses the given files and returns the absolute path of the files
func (o *options) ParseFilePaths() []string {
	files := o.FilesOrDefault()
	return o.extractParsedFiles(files)
}

func (o *options) ParseWatchFilePaths() []string {
	files := o.lookupWatchFile
	return o.extractParsedFiles(files)
}

func (o *options) extractParsedFiles(files []string) []string {
	var parsedFiles []string
	d.logf("[dotenv] Files to parse: %+v", files)

	for _, file := range files {
		for _, path := range o.lookupPaths {
			if filepath.IsAbs(file) {
				parsedFiles = append(parsedFiles, file)
				continue
			}

			if fullPath := filepath.Join(path, file); fileExists(fullPath) {
				parsedFiles = append(parsedFiles, fullPath)
				continue
			}
		}

		if fullPath := o.gitRepoFile(file); fileExists(fullPath) {
			parsedFiles = append(parsedFiles, fullPath)
			continue
		}

		if fullPath := o.goModFile(file); fileExists(fullPath) {
			parsedFiles = append(parsedFiles, fullPath)
			continue
		}
	}

	d.logf("[dotenv] Parsed files: %s", parsedFiles)
	return filterValidFiles(parsedFiles)
}

func (o *options) goModFile(file string) string {
	if !o.lookupMod {
		return ""
	}

	modPath := goModPath()
	if modPath == "" {
		return ""
	}

	return filepath.Join(modPath, file)
}

func (o *options) gitRepoFile(file string) string {
	if !o.lookupGit {
		return ""
	}

	repoPath := gitRepoPath()
	if repoPath == "" {
		return ""
	}

	return filepath.Join(repoPath, file)
}

func fileExists(fullPath string) bool {
	_, err := os.Stat(fullPath)
	return !os.IsNotExist(err)
}

func filterValidFiles(files []string) []string {
	var validFilePaths []string
	for _, file := range files {
		if _, err := os.Stat(file); err == nil {
			validFilePaths = append(validFilePaths, file)
		}
	}
	return validFilePaths
}

func gitRepoPath() string {
	bytes, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err == nil {
		return strings.TrimSuffix(string(bytes), "\n")
	}
	return ""
}

func goModPath() string {
	bytes, err := exec.Command("go", "env", "GOMOD").Output()
	if err == nil {
		return filepath.Dir(string(bytes))
	}
	return ""
}

func newOpts() *options {
	return &options{
		lookupFile:  []string{},
		lookupPaths: []string{"./"},
	}
}
