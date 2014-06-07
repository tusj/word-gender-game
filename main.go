package main

import (
	"flag"
	"fmt"
	"path"
	"strings"
	// "github.com/ajstarks/svgo"
	"io/ioutil"
	"os"
)

const (
	width  = 70  // px
	height = 100 // px
)

var inputDir = flag.String("input-dir", "", "Input directory for image files.")
var outputDir = flag.String("output-dir", "out", "Output directory for image files.")

func main() {
	flag.Parse()
	// Check for compulsory input arguments
	switch {
	case *outputDir == "":
		fallthrough
	case *inputDir == "":
		fmt.Fprintln(os.Stderr, "Input and output directory must be given")
		os.Exit(1)
	}

	// Sanity check
	s, err := os.Stat(*inputDir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintln(os.Stderr, "Input directory does not exist")
			os.Exit(2)
		}
		fmt.Fprintln(os.Stderr, "Could not stat input directory:", err)
		os.Exit(3)
	}

	if !s.IsDir() {
		fmt.Fprintln(os.Stderr, "Input directory is not a directory")
		os.Exit(4)
	}

	// Make the output directory if none exists
	s, err = os.Stat(*outputDir)
	if os.IsNotExist(err) {
		err := os.Mkdir(*outputDir, 0755)
		checkErr(err)
	} else if err != nil {
		fmt.Fprintln(os.Stderr, "Could not stat output directory:", err)
		os.Exit(5)
	} else if !s.IsDir() {
		fmt.Fprintln(os.Stderr, "Output directory is not a directory")
		os.Exit(6)
	}

	dirFilter := func(f os.FileInfo) bool {
		return f.IsDir()
	}

	lsDirs := func(dir string) []string {
		return ls(dir, dirFilter)
	}

	fileFilter := func(f os.FileInfo) bool {
		return !f.IsDir()
	}

	lsFiles := func(dir string) []string {
		return ls(dir, fileFilter)
	}

	makeCardsWrapper := func(ending string, files []string) {
		makeCards(*inputDir, *outputDir, ending, files)
	}

	// For every directory in input directory
	for _, d := range lsDirs(*inputDir) {
		makeCardsWrapper(d, lsFiles(path.Join(*inputDir, d)))
	}
}

func transformStrings(files []string, transformer func(string) string) []string {
	var f = make([]string, len(files))
	for i, v := range files {
		f[i] = transformer(v)
	}

	return f
}

func makeCards(inDirectory string, outDirectory string, ending string, files []string) {
	err := os.Mkdir(path.Join(outDirectory, ending), 0755)
	checkErr(err)

	removeFileType := func(s string) string {
		return strings.Split(s, ".")[0]
	}

	words := transformStrings(files, removeFileType)
	for _, v := range words {
	}
}

func ls(directory string, filter func(os.FileInfo) bool) []string {
	dirs, err := ioutil.ReadDir(directory)
	checkErr(err)
	endings := make([]string, 0)

	for _, v := range dirs {
		if filter(v) {
			endings = append(endings, v.Name())
		}
	}

	return endings

}

func checkErr(err error) {
	if err != nil {
		panic("err: " + err.Error())
	}
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
