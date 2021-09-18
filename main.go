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

func main() {
	flag.Parse()
	root := flag.Arg(0)
	absRoot, err := filepath.Abs(root)

	// Get all .po translationFiles in repo
	translationFiles, err := walkMatch(absRoot, "*.po")
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	// Parse repo's gitignore
	gitignorePath := filepath.Join(absRoot, ".gitignore")

	// Include venv in case it's not ignored
	ignore, err := ignore.CompileIgnoreFileAndLines(gitignorePath, ".venv")
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(2)
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
			fmt.Fprint(os.Stderr, err)
			os.Exit(3)
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			if translationRegex.MatchString(scanner.Text()) {
				translation := strings.Split(scanner.Text(), "msgid ")[1]

				translation = strings.Trim(translation, `"`)
				used, err := isUsedInPaths(translation, pythonFiles, htmlFiles)
				if err != nil {
					fmt.Fprint(os.Stderr, err)
					os.Exit(4)
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
	// TODO: check file if --check flag

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
