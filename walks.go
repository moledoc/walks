/*
Package walks provides functions to walk directory structure and perform user-defined actions on files and directories.
*/

package walks

import (
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"
)

// WaitGroup is a variable to easily handle goroutine waiting.
var WaitGroup sync.WaitGroup

// Search is a variable to hold expressions of directories and files to search.
var Search *regexp.Regexp = regexp.MustCompile("")

// Ignore is a variable to hold regexp expression of directories and files to ignore.
var Ignore *regexp.Regexp = regexp.MustCompile("")

// SetIgnore sets global Ignore with the contents of ignore file,
// where each line represents one file or directory to ignore.
func SetIgnore(ignFilePath string) {
	if ignFilePath == "" {
		return
	}
	tempIgn := false
	contents, err := os.ReadFile(ignFilePath)
	if err != nil {
		// create temp ignore file, if does not exist, do not get an error (so that we could have default ignore file in program flag, see [ado](https://github.com/moledoc/directory/tree/main/ado)).
		err = os.WriteFile(ignFilePath, []byte(""), 0755)
		if err != nil {
			log.Fatal(err)
		}
		tempIgn = true
	}
	var ign string
	for i, line := range strings.Split(string(contents), "\n") {
		if line == "" {
			break
		}
		if i != 0 {
			ign += "|"
		}
		if line == "." || line == ".." {
			line = "^" + line + "$"
		}
		ign += strings.Replace(line, ".", "\\.", -1)
	}
	Ignore = regexp.MustCompile(ign)
	if tempIgn {
		os.RemoveAll(ignFilePath)
	}
}

// Walk is a concurrent function that walks recursively given directory structure, performing given actions on files and directories.
// Actions on files and directories are expected to take the corresponding file/dir path as an argument and not return anything.
// Directories and files can also be ignored by setting Ignore value with SetIgnore function or manually before Walk call.
// Depth of directory structure can be controlled with variables depth.
func Walk(root string, fileAction func(string), dirAction func(string), depth int) {
	WaitGroup.Add(2)
	go func() { defer WaitGroup.Done(); walk(root, fileAction, dirAction, depth, 0) }()
	WaitGroup.Wait()
}

// walk is Walk's inner function, that actually walks the directory structure.
// walk is concurrent.
func walk(root string, fileAction func(string), dirAction func(string), depth int, level int) {
	defer WaitGroup.Done()
	if depth != -1 && level > depth {
		return
	}
	if pathType, err := os.Stat(root); err != nil {
		log.Fatal(err)
	} else if !pathType.IsDir() {
		log.Fatal("Argument `root` must be path to a directory")
	}
	subpaths, err := ioutil.ReadDir(root)
	if err != nil {
		log.Fatal(err)
	}
	for _, path := range subpaths {
		pathName := root + "/" + path.Name()
		if Ignore.MatchString(pathName) && Ignore.String() != "" {
			continue
		}
		switch pathType := path.Mode(); {
		case pathType.IsDir():
			dirAction(pathName)
			WaitGroup.Add(1)
			go walk(pathName, fileAction, dirAction, depth, level+1)
		case pathType.IsRegular():
			fileAction(pathName)
		default:
			log.Fatal("Unreachable: invalid path type.")
		}
	}
}

// WalkLinear walks recursively given directory structure, performing given actions on files and directories.
// Actions on files and directories are expected to take the corresponding file/dir path as an argument and not return anything.
// Directories and files can also be ignored by setting Ignore value with SetIgnore function or setting it manually.
// Depth of directory structure can be controlled with variables depth (and level).
func WalkLinear(root string, fileAction func(string), dirAction func(string), depth int, level int) {
	if level == depth {
		return
	}
	if pathType, err := os.Stat(root); err != nil {
		log.Fatal(err)
	} else if !pathType.IsDir() {
		log.Fatal("Argument `root` must be path to a directory")
	}
	subpaths, err := ioutil.ReadDir(root)
	if err != nil {
		log.Fatal(err)
	}
	for _, path := range subpaths {
		pathName := root + "/" + path.Name()
		if Ignore.MatchString(pathName) && Ignore.String() != "" {
			continue
		}
		switch pathType := path.Mode(); {
		case pathType.IsDir():
			dirAction(pathName)
			WalkLinear(pathName, fileAction, dirAction, depth, level+1)
		case pathType.IsRegular():
			fileAction(pathName)
		default:
			log.Fatal("Unreachable: invalid path type.")
		}
	}
}
