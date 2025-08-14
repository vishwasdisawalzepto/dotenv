package dotenv

import "github.com/fsnotify/fsnotify"

var d = new()

func Load() error {
	return d.Load()
}

func Reset() {
	d = new()
}

func Overload() error {
	return d.Overload()
}

func OverloadWatchFiles() error {
	return d.OverloadWatchFiles()
}

func OptLookupGit() {
	d.opts.lookupGit = true
}

func OptLookupMod() {
	d.opts.lookupMod = true
}

func OptDisableFileExpand() {
	d.opts.disableFileExpand = true
}

func OptDisablePathExpand() {
	d.opts.disablePathExpand = true
}

func OptDebug() {
	d.opts.debug = true
}

func OptLookupFile(file string) {
	d.opts.lookupFile = append(d.opts.lookupFile, file)
}

func OptLookupWatchFile(file string) {
	d.opts.lookupWatchFile = append(d.opts.lookupWatchFile, file)
}

func WatchConfig() {
	d.opts.watchConfig = true
	files := d.opts.ParseWatchFilePaths()
	for _, file := range files {
		go d.WatchConfig(file)
	}
}

func OnConfigChange(fn func(fsnotify.Event)) {
	d.opts.onConfigChange = fn
}
