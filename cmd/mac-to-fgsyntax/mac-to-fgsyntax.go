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

	fmt.Println("Enter address group name that mac address objects should be added to:")
	addrGrp := readUserInputSingle()

	macList := readTextFile(inputpath)
	macFGList := convertToFGsyntax(macList, addrGrp)

	writeTextFile(outputpath, macFGList)
}

func convertToFGsyntax(macList []string, addGrp string) []string {
	var macFGList []string
	appendList := "append member "
	macFGList = append(macFGList, fmt.Sprintf("config firewall address"))
	for _, mac := range macList {
		mac = strings.ToLower(strings.TrimSpace(mac))
		macFGList = append(macFGList, fmt.Sprintf("    edit \"%s\"", mac))
		macFGList = append(macFGList, fmt.Sprintf("        set type mac"))
		macFGList = append(macFGList, fmt.Sprintf("        set start-mac %s", mac))
		macFGList = append(macFGList, fmt.Sprintf("        set end-mac %s", mac))
		macFGList = append(macFGList, fmt.Sprintf("    next"))
		appendList = fmt.Sprintf("%s \"%s\"", appendList, mac)
	}
	macFGList = append(macFGList, fmt.Sprintf("end"))
	macFGList = append(macFGList, fmt.Sprintf("\nconfig firewall addrgrp"))
	macFGList = append(macFGList, fmt.Sprintf("    edit %s", addGrp))
	macFGList = append(macFGList, fmt.Sprintf("        %s", appendList))
	macFGList = append(macFGList, fmt.Sprintf("    next"))
	macFGList = append(macFGList, fmt.Sprintf("end"))
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

func readUserInputSingle() string {
	s := bufio.NewScanner(os.Stdin)
	s.Scan()
	ln := s.Text()
	if err := s.Err(); err != nil {
		log.Fatal(err)
	}
	return ln
}
