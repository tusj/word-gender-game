package main

import (
	"flag"
	// "fmt"
	"github.com/ajstarks/svgo"
	"io/ioutil"
	"os"
)

const (
	width  = 70  // px
	height = 100 // px
)

var outputFile = flag.String("output-file", "", "Output filename for svg file. If none is specifed, outputs to stdout")

func main() {
	flag.Parse()
	var out *os.File
	if *outputFile != "" {
		newFile, err := os.Create(*outputFile)
		checkErr(err)
		out = newFile
	} else {
		out = os.Stdout
	}
	canvas := svg.New(out)
	canvas.Start(width, height)
	canvas.Circle(width/2, height/2, 100)
	canvas.Text(width/2, height/2, "Hello, SVG", "text-anchor:middle;font-size:30px;fill:white")
	canvas.End()
}

func getEndings() []string {
	dirs, err := ioutil.ReadDir(".")
	checkErr(err)
	endings := make([]string, 0)

	for _, v := range dirs {
		if v.IsDir() {
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
