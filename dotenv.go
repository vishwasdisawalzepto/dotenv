package dotenv

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/joho/godotenv"
)

var NoFailsToLoadErr = errors.New("[dotenv] No files to load")

type dotenv struct {
	files []string
	opts  *options
}

func (d *dotenv) Load() error {
	var loadErr error
	parsedFiles := d.opts.ParseFilePaths()
	if len(parsedFiles) == 0 {
		d.logf("[dotenv] No files found to load")
		return NoFailsToLoadErr
	}
	d.logf("[dotenv] Loading parsedFiles %s", strings.Join(parsedFiles, ", "))

	for _, parsedFile := range parsedFiles {
		if err := loadFile(parsedFile, false); err != nil {
			loadErr = wrapError(loadErr, err)
			d.logf("[dotenv] Loading parsedFile %s failed with error %s", parsedFile, err.Error())
			continue
		}
		d.logf("[dotenv] Loaded parsedFile %s", parsedFile)
		d.files = append(d.files, parsedFile)
	}
	d.logf("[dotenv] Loaded parsedFiles %s", strings.Join(d.files, ", "))
	return loadErr
}

func (d *dotenv) logf(args ...interface{}) {
	if d.opts.debug {
		log.Println(fmt.Sprintf(args[0].(string), args[1:]...))
	}
}

func wrapError(loadErr error, err error) error {
	if loadErr != nil {
		loadErr = fmt.Errorf("%s\n%w", loadErr.Error(), err)
	} else {
		loadErr = err
	}
	return loadErr
}

func loadFile(file string, overload bool) error {
	fileEnv, err := godotenv.Read(file)
	if err != nil {
		d.logf("[dotenv] Loading file %s failed with error %s", file, err.Error())
		return err
	}
	d.logf("[dotenv] Loading file %s", file)

	osEnv := map[string]bool{}
	for _, rawEnvLine := range os.Environ() {
		key := strings.Split(rawEnvLine, "=")[0]
		osEnv[key] = true
	}

	for key, value := range fileEnv {
		if !osEnv[key] || overload {
			_ = os.Setenv(key, value)
		}
	}

	return nil
}

func (d *dotenv) overloadFromFiles(parsedFiles []string) error {
	for _, parsedFile := range parsedFiles {
		if err := loadFile(parsedFile, true); err != nil && d.opts.debug {
			log.Printf("[dotenv] Overloading parsedFile %s failed with error %s", parsedFile, err.Error())
			continue
		}
		d.files = append(d.files, parsedFile)
	}
	return nil
}

func (d *dotenv) Overload() error {
	return d.overloadFromFiles(d.opts.ParseFilePaths())
}

func (d *dotenv) OverloadWatchFiles() error {
	return d.overloadFromFiles(d.opts.ParseWatchFilePaths())
}

func new() *dotenv {
	return &dotenv{opts: newOpts()}
}

func (d *dotenv) WatchConfig(filename string) {
	initWG := sync.WaitGroup{}
	initWG.Add(1)
	go func() {
		watcher, err := newWatcher()
		if err != nil {
			log.Println(fmt.Sprintf("failed to create watcher: %s", err))
			os.Exit(1)
		}
		defer watcher.Close()

		configFile := filepath.Clean(filename)
		configDir, _ := filepath.Split(configFile)
		realConfigFile, _ := filepath.EvalSymlinks(filename)

		eventsWG := sync.WaitGroup{}
		eventsWG.Add(1)
		go func() {
			for {
				select {
				case event, ok := <-watcher.Events:
					if !ok { // 'Events' channel is closed
						eventsWG.Done()
						return
					}
					currentConfigFile, _ := filepath.EvalSymlinks(filename)
					// we only care about the config file with the following cases:
					// 1 - if the config file was modified or created
					// 2 - if the real path to the config file changed (eg: k8s ConfigMap replacement)
					if (filepath.Clean(event.Name) == configFile &&
						(event.Has(fsnotify.Write) || event.Has(fsnotify.Create))) ||
						(currentConfigFile != "" && currentConfigFile != realConfigFile) {
						realConfigFile = currentConfigFile
						if d.opts.onConfigChange != nil {
							d.opts.onConfigChange(event)
						}
					} else if filepath.Clean(event.Name) == configFile && event.Has(fsnotify.Remove) {
						eventsWG.Done()
						return
					}

				case err, ok := <-watcher.Errors:
					if ok { // 'Errors' channel is not closed
						log.Println(fmt.Sprintf("watcher error: %s", err))
					}
					eventsWG.Done()
					return
				}
			}
		}()
		watcher.Add(configDir)
		initWG.Done()   // done initializing the watch in this go routine, so the parent routine can move on...
		eventsWG.Wait() // now, wait for event loop to end in this go-routine...
	}()
	initWG.Wait() // make sure that the go routine above fully ended before returning
}

func newWatcher() (*fsnotify.Watcher, error) {
	return fsnotify.NewWatcher()
}
