// Copyright (C) 2021  Johanan Idicula
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

// Command django-translation-cleaner is a tool for cleaning unused
// translations from .po files in a Django project.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	ignore "github.com/sabhiram/go-gitignore"
)

var check = flag.Bool("check", false, "")
var usage = `Usage: django-translation-cleaner [options...] <path to repo>

django-translation-cleaner is a tool for cleaning unused translations from .po
files in a Django project.

Example:

- Clean unused translations from a Django project:
    $ django-translation-cleaner /path/to/repo

- Check for unused translations in a Django project. If there are any unused
translations, prints them to stdout and returns non-zero exit code:
    $ django-translation-cleaner --check /path/to/repo

Options:
 --check (-c) Checks for unused translations without removing them. Returns
non-zero exit code if there are any unused translations.
 --help (-h) Prints this message.
`

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s\n", usage)
	}
	flag.Parse()
	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(2)
	}
	root := flag.Arg(0)
	absRoot, err := filepath.Abs(root)

	// Get all .po translationFiles in repo
	translationFiles, err := walkMatch(absRoot, "*.po")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Parse repo's gitignore
	// TODO: make this conditional on whether path has a gitignore in its lineage?
	gitignorePath := filepath.Join(absRoot, ".gitignore")

	// Include venv in case it's not ignored
	ignore, err := ignore.CompileIgnoreFileAndLines(gitignorePath, ".venv")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(3)
	}

	// Clean ignored files from list of .po files
	translationFiles = cleanIgnoredPaths(translationFiles, ignore)
	// fmt.Println(translationFiles)

	pythonFiles, err := walkMatch(absRoot, "*.py")
	pythonFiles = cleanIgnoredPaths(pythonFiles, ignore)
	htmlFiles, err := walkMatch(absRoot, "*.html")
	htmlFiles = cleanIgnoredPaths(htmlFiles, ignore)

	var unusedCount int
	var unused string

	translationRegex := regexp.MustCompile(`(?m)^msgid.*$`)
	for _, path := range translationFiles {
		f, err := os.Open(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(4)
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			if translationRegex.MatchString(scanner.Text()) {
				translation := strings.Split(scanner.Text(), "msgid ")[1]

				translation = strings.Trim(translation, `"`)
				used, err := isUsedInPaths(translation, pythonFiles, htmlFiles)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(5)
				}
				if !used {
					unusedCount++
					if *check {
						unused += translation + "\n"
					}
				}
			}
		}

	}

	// TODO: write out new file by default

	if *check {
		if unusedCount > 0 {
			fmt.Fprintf(os.Stdout, "%s\n", unused)
			fmt.Fprintln(os.Stdout, `💥 💔 💥`)
			fmt.Fprintf(os.Stdout, "\033[1m%v unused translations\033[0m\n", unusedCount)
			os.Exit(1)
		} else {
			fmt.Fprintln(os.Stdout, `All done! ✨ 🍰 ✨\n`)
			os.Exit(0)
		}
	}

}

// walkMatch walks a provided path and returns all paths matching a provided
// shell pattern.
func walkMatch(root, pattern string) ([]string, error) {
	var matches []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if matched, err := filepath.Match(pattern, filepath.Base(path)); err != nil {
			return err
		} else if matched {
			matches = append(matches, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return matches, nil
}

// cleanIgnoredFiles cleans gitignored paths from a list of filepaths.
func cleanIgnoredPaths(files []string, ignore *ignore.GitIgnore) []string {
	var trackedFiles []string
	for _, name := range files {
		if !ignore.MatchesPath(name) {
			trackedFiles = append(trackedFiles, name)
		}
	}
	return trackedFiles
}

// isUsedInPaths searches for a string in the filepaths provided.
func isUsedInPaths(msg string, pathSlices ...[]string) (bool, error) {
	var allPaths []string

	for _, paths := range pathSlices {
		allPaths = append(allPaths, paths...)
	}

	for _, path := range allPaths {
		f, err := os.Open(path)
		if err != nil {
			return true, err
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			if strings.Contains(scanner.Text(), msg) {
				return true, nil
			}
		}
	}
	return false, nil
}
