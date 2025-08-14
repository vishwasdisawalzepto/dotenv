# .env
dotenv is the golang library to load the .env for golang project the golang way.

## Usage

```go
package main

import (
    "fmt"
    "github.com/gwthm-in/dotenv"
    "os"
)

func main() {
	dotenv.OptLookupMod() // to load the .env file from the module root
	dotenv.OptLookupGit() // to load the .env file from the git root
	dotenv.OptLookupFile("application.env") // to look for a specific file instead of .env file
	dotenv.OptLookupFile("application.env.$ENV") // to look for a specific file instead of .env file
	dotenv.OptLookupWatchFile("dynamic.env") //watch for dynamic file
	dotenv.OptLookupWatchFile("dynamic_2.env") //you could configure multiple dynamic files
	dotenv.OnConfigChange(func(event fsnotify.Event) {  //callback for changes in dynamic env
		dotenv.OverloadWatchFiles()
	})

	dotenv.WatchConfig() //watcher activated

	err := dotenv.Load() //loads both static and dynamic files
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(os.Getenv("DB_HOST"))
}```
