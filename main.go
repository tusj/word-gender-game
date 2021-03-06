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
	colors "github.com/lucasb-eyer/go-colorful"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"strings"
	"text/template"
)

var inputDir = flag.String("input-dir", "asset", "Input directory for image files.")
var templateFile = flag.String("template-file", "card.html.template", "Golang template file used to make the cards")
var templateRefFile = flag.String("template-ref-file", "cardref.html.template", "Golang template file used to make the card references")
var outputDir = "cards"

var cardTemplate string
var cardRefTemplate string

func main() {
	flag.Parse()

	// Check for compulsory input arguments
	if *inputDir == "" {
		fmt.Fprintln(os.Stderr, "Input and output directory must be given")
		os.Exit(1)
	}

	// Read the template file
	tmpl, err := ioutil.ReadFile(*templateFile)
	checkErr(err)
	cardTemplate = string(tmpl)

	tmpll, err := ioutil.ReadFile(*templateRefFile)
	checkErr(err)
	cardRefTemplate = string(tmpll)

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
	s, err = os.Stat(outputDir)
	if os.IsNotExist(err) {
		err := os.Mkdir(outputDir, 0755)
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

	makeCardsWrapper := func(ending string, files []string, color string) {
		makeCards(*inputDir, outputDir, ending, files, color)
	}

	dirs := lsDirs(*inputDir)

	// Every ending should have its own color
	// We want to avoid similar colors
	randColors := colors.FastHappyPalette(len(dirs))

	// For every directory in input directory
	// Make an output directory containing the cards
	// made with the images of the directory in the input directory
	for i, d := range dirs {
		makeCardsWrapper(d, lsFiles(path.Join(*inputDir, d)), randColors[i].Hex())
	}
}

// String transformer
func transformStrings(files []string, transformer func(string) string) []string {
	var f = make([]string, len(files))
	for i, v := range files {
		f[i] = transformer(v)
	}

	return f
}

// The data struct used to hold references to the other cards
type CardRef struct {
	Image string
	Word  string
}

// The data struct to be used with the card template
type Card struct {
	Color      string
	Ending     string
	Image      string
	Word       string
	OtherCards []CardRef
}

// For every ending, there is a set of words
type Cards struct {
	Cards []Card
}

type CardRefs struct {
	Ending   string
	Color    string
	CardRefs []CardRef
}

func makeCards(inDirectory string, outDirectory string, ending string, files []string, color string) {
	// Don't do anything if there are no input files
	if len(files) == 0 {
		return
	}

	// Make output directory if it does not exist
	// err := os.Mkdir(path.Join(outDirectory, ending), 0755)
	// if err != nil {
	// 	if os.IsNotExist(err) {
	// 		checkErr(err)
	// 	}
	// }

	// Extract the words from the file names
	removeFileType := func(s string) string { return strings.Split(s, ".")[0] }
	removeGender := func(s string) string { return strings.Split(s, " ")[1] }

	// Remove the file type endings from the words
	words := transformStrings(files, removeFileType)
	// Remove the word genders
	wordsWithoutGender := transformStrings(words, removeGender)

	imagePath := func(v string) string {
		image, err := url.Parse(path.Join(inDirectory, ending, v))
		checkErr(err)
		return "file://" + image.String()
	}
	images := transformStrings(files, imagePath)
	cardRefs := CardRefs{ending, color, make([]CardRef, len(files))}

	// The template data
	data := Cards{make([]Card, len(files))}

	// The references of the other cards for each ending for each card
	// Every card contains references to the other cards with the same ending
	// except the card itself
	otherCards := make([][]CardRef, len(files))

	// Make the template data
	for i := range files {
		data.Cards[i] = Card{
			color,
			ending,
			images[i],
			words[i],
			nil}
	}

	// Make the references
	for i := range data.Cards {
		otherCards[i] = make([]CardRef, len(data.Cards)-1)
		j := 0
		for k := range data.Cards {
			if k == i {
				continue
			}

			otherCards[i][j] = CardRef{
				data.Cards[k].Image,
				wordsWithoutGender[k]}
			j++
		}
	}

	// Make the card reference
	for i := range cardRefs.CardRefs {
		cardRefs.CardRefs[i].Image = images[i]
		cardRefs.CardRefs[i].Word = wordsWithoutGender[i]
	}

	// Set the references for the template data
	for i := range data.Cards {
		data.Cards[i].OtherCards = otherCards[i]
	}

	// Produce the cards from the template with the template data
	f, err := os.Create(path.Join(outDirectory, ending+".html"))
	defer f.Close()
	if err != nil {
		if os.IsExist(err) {
			checkErr(err)
		}
	}

	tmpl, err := template.New(ending).Parse(cardTemplate)
	checkErr(err)
	err = tmpl.Execute(f, data)
	checkErr(err)

	// Make the card reference
	ff, err := os.Create(path.Join(outDirectory, ending+"-ref.html"))
	defer ff.Close()
	if err != nil {
		if os.IsExist(err) {
			checkErr(err)
		}
	}

	tmpll, err := template.New(ending + "-ref").Parse(cardRefTemplate)
	checkErr(err)
	err = tmpll.Execute(ff, cardRefs)
	checkErr(err)
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
