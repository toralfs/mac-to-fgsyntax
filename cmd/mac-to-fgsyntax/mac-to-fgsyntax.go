package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

const inputpath string = "../../input/inputMacList.txt"
const outputpath string = "../../output/outputMacList.txt"

func main() {

	macList := readTextFile(inputpath)
	macFGList := convertToFGsyntax(macList)
	writeTextFile(outputpath, macFGList)
	fmt.Println()

	/*for _, l := range macFGList {
		fmt.Println(l)
	}
	fmt.Println(appendList)*/
}

func convertToFGsyntax(macList []string) []string {
	var macFGList []string
	appendList := "append member "
	for _, mac := range macList {
		mac = strings.TrimSpace(mac)
		macFGList = append(macFGList, fmt.Sprintf("edit \"%s\"", mac))
		macFGList = append(macFGList, fmt.Sprintf("set type mac"))
		macFGList = append(macFGList, fmt.Sprintf("set start-mac %s", mac))
		macFGList = append(macFGList, fmt.Sprintf("set end-mac %s", mac))
		macFGList = append(macFGList, fmt.Sprintf("next"))
		appendList = fmt.Sprintf("%s \"%s\"", appendList, mac)
	}
	macFGList = append(macFGList, fmt.Sprintf("\n%s", appendList))
	return macFGList
}

func readTextFile(path string) []string {
	var fileContent []string

	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err = file.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	s := bufio.NewScanner(file)

	for s.Scan() {
		fileContent = append(fileContent, s.Text())
	}

	return fileContent
}

func writeTextFile(path string, text []string) {
	f, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err = f.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	for _, l := range text {
		_, err := f.WriteString(fmt.Sprintf("%s\n", l))
		if err != nil {
			log.Fatal(err)
		}
	}
}
