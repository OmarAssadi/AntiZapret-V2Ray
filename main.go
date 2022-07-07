package main

import (
	"flag"
	"fmt"
	"github.com/golang/protobuf/proto"
	"golang.org/x/text/encoding/charmap"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const listNamePlaceHolder = "<list-name>.txt"

var (
	listName      = flag.String("list-name", "ZAPRETINFO", "Name of the list")
	input         = flag.String("input", filepath.Join("./", "z-i", "dump.csv"), "Path to the Zapret-Info CSV")
	geoSiteFile   = flag.String("geosite-filename", "geosite.dat", "Name of the output file")
	plainTextFile = flag.String("plaintext-filename", listNamePlaceHolder, "Name of the plaintext output file")
	outputPath    = flag.String("output-dir", "./publish", "Output path to the generated files")
)

func main() {
	flag.Parse()
	domainList, parseErr := Unmarshal(charmap.Windows1251.NewDecoder(), strings.ToUpper(*listName), *input)
	if parseErr != nil {
		panic(parseErr)
	}

	if err := domainList.Flatten(); err != nil {
		fmt.Println("Failed:", err)
		os.Exit(1)
	}

	if geoSites := domainList.ToGeoSites(); geoSites != nil {
		geoSiteData, err := proto.Marshal(geoSites)
		if err != nil {
			fmt.Println("Failed:", err)
			os.Exit(1)
		}
		if err := os.MkdirAll(*outputPath, 0755); err != nil {
			fmt.Println("Failed:", err)
			os.Exit(1)
		}
		if err := ioutil.WriteFile(filepath.Join(*outputPath, *geoSiteFile), geoSiteData, 0644); err != nil {
			fmt.Println("Failed:", err)
			os.Exit(1)
		}
		fmt.Printf("%s has been generated successfully in '%s'.\n", *geoSiteFile, *outputPath)
	}

	var outputName string
	if *plainTextFile == listNamePlaceHolder {
		outputName = *listName + ".txt"
	} else {
		outputName = *plainTextFile
	}

	if err := ioutil.WriteFile(filepath.Join(*outputPath, outputName), domainList.ToPlainText(), 0644); err != nil {
		fmt.Println("Failed:", err)
		os.Exit(1)
	}
	fmt.Printf("%s has been generated successfully in '%s'.\n", outputName, *outputPath)
}
