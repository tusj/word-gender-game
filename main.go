// This program takes produces a game to learn the genders of
// some word endings in french.

// It takes an input directory containing images of words which
// ends with a certain ending, for example -ette, and produces
// one card for each image. Every card will contain images of
// the other cards which have the same ending.
// The program uses the template/html package to produce the cards
// in html by linking to the pictures in the html.

// The images should be named the name of the object the image
// depicts, and the images should be organised in subfolders
// which are named after the proper ending.
// For example, the subfolder ette should contain images
// whose depiction ends with -ette, for example une cantine and une copine.
package main

import (
	"flag"
	"fmt"
	// "github.com/ajstarks/svgo"
	// svg "github.com/tusj/go-svg"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"strings"
	"text/template"
)

var inputDir = flag.String("input-dir", "", "Input directory for image files.")
var outputDir = flag.String("output-dir", "out", "Output directory for image files.")
var templateFile = flag.String("template-file", "card.html.template", "Golang template file used to make the cards")

var cardTemplate string

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

	// Read the template file
	tmpl, err := ioutil.ReadFile(*templateFile)
	checkErr(err)
	cardTemplate = string(tmpl)

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

	// Set absolute path of input directory if it is not
	if !path.IsAbs(*inputDir) {
		wd, err := os.Getwd()
		checkErr(err)
		*inputDir = path.Join(wd, *inputDir)
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

	dirFilter := func(f os.FileInfo) bool { return f.IsDir() }
	lsDirs := func(dir string) []string { return ls(dir, dirFilter) }
	fileFilter := func(f os.FileInfo) bool { return !f.IsDir() }
	lsFiles := func(dir string) []string { return ls(dir, fileFilter) }

	makeCardsWrapper := func(ending string, files []string) {
		makeCards(*inputDir, *outputDir, ending, files)
	}

	// For every directory in input directory
	// Make an output directory containing the cards
	// made with the images of the directory in the input directory
	for _, d := range lsDirs(*inputDir) {
		makeCardsWrapper(d, lsFiles(path.Join(*inputDir, d)))
	}
}

// General transformer
func transformStrings(files []string, transformer func(string) string) []string {
	var f = make([]string, len(files))
	for i, v := range files {
		f[i] = transformer(v)
	}

	return f
}

// The data struct to be used with the template
type Card struct {
	Color      string
	Ending     string
	Image      string
	Word       string
	OtherCards []Card
}

func makeCards(inDirectory string, outDirectory string, ending string, files []string) {
	// Don't do anything if there are no input files
	if len(files) == 0 {
		return
	}

	// Make directory if it does not exist
	err := os.Mkdir(path.Join(outDirectory, ending), 0755)
	if err != nil {
		if os.IsNotExist(err) {
			checkErr(err)
		}
	}

	// Extract the words from the file names
	removeFileType := func(s string) string { return strings.Split(s, ".")[0] }
	removeGender := func(s string) string { return strings.Split(s, " ")[1] }

	words := transformStrings(files, removeFileType)
	wordsWithoutGender := transformStrings(words, removeGender)

	for _, v := range wordsWithoutGender {
		fmt.Println(v)
	}
	data := make([]Card, len(files))
	otherCards := make([]Card, len(files))

	for i, v := range files {
		image, err := url.Parse(path.Join(inDirectory, ending, v))
		checkErr(err)
		data[i] = Card{"lightgreen", ending, "file://" + image.String(), words[i], nil}
		otherCards[i] = Card{Image: data[i].Image, Word: wordsWithoutGender[i]}
	}

	for i := range data {
		data[i].OtherCards = otherCards
	}
	for _, v := range data {
		fmt.Println()
		fmt.Println()
		fmt.Println(v.Image)
	}

	for i := range files {
		f, err := os.Create(path.Join(outDirectory, ending, words[i]+".html"))
		defer f.Close()
		if err != nil {
			if os.IsExist(err) {
				checkErr(err)
			}
		}
		tmpl, err := template.New(words[i]).Parse(cardTemplate)
		checkErr(err)

		err = tmpl.Execute(f, data[i])
		checkErr(err)
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
