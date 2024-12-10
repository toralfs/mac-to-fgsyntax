package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const inputpath string = "input/inputMacList.txt"
const outputpath string = "output/outputMacList.txt"

func main() {
	// Welcome message
	fmt.Printf("------------------------------------------------------------------------------\n")
	fmt.Printf("---------------- MAC address to FortiGate syntax converter -------------------\n")
	fmt.Printf("------------------------------------------------------------------------------\n")
	fmt.Printf("Program expects a list of mac addresses with one mac address on each line\n")
	fmt.Printf("Expected mac address format is: xx:xx:xx:xx:xx:xx\n\n\n")

	// First get address group name, common for both methods.
	fmt.Println("Enter address group name that mac address objects should be added to:")
	addrGrp := readUserInputSingle()

	// Select usage method
	for {
		usrChoice := selectMethod()
		switch usrChoice {
		case 1: // cli in/out method
			fmt.Printf("Enter the list of mac addresses and then press CTRL+D or CTRL+Z (depending on OS):\n")
			macList := readUserInput()
			macFGList := convertToFGsyntax(macList, addrGrp)

			fmt.Printf("\nList converted to Fortigate Syntax: \n")
			fmt.Printf("--------------------------------------\n")
			for _, l := range macFGList {
				fmt.Println(l)
			}
			fmt.Printf("\nPress any button to exit\n")
			readUserInput()
			os.Exit(0)
		case 2: // file in/out method
			macList := readTextFile(inputpath)
			macFGList := convertToFGsyntax(macList, addrGrp)

			writeTextFile(outputpath, macFGList)
			fmt.Printf("List converted to FortiGate syntax and writting to %s\n", outputpath)
			fmt.Printf("\nPress any button to exit\n")
			readUserInput()
			os.Exit(0)
		default:
			fmt.Printf("Invalid choice, try again..\n\n")
		}
	}
}

func selectMethod() int {
	fmt.Println("Select usage method:")
	fmt.Println("1: command line input/output")
	fmt.Println("2: input/output files")

	choice, err := strconv.Atoi(readUserInputSingle())
	if err != nil {
		fmt.Println("Invalid input, needs to be a number.")
		return -1
	}
	return choice
}

func convertToFGsyntax(macList []string, addGrp string) []string {
	var macFGList []string
	appendList := "append member "
	macFGList = append(macFGList, "config firewall address")
	for _, mac := range macList {
		if validateMac(mac) {
			mac = strings.ToLower(strings.TrimSpace(mac))
			macFGList = append(macFGList, fmt.Sprintf("    edit \"%s\"", mac))
			macFGList = append(macFGList, "        set type mac")
			macFGList = append(macFGList, fmt.Sprintf("        set start-mac %s", mac))
			macFGList = append(macFGList, fmt.Sprintf("        set end-mac %s", mac))
			macFGList = append(macFGList, "    next")
			appendList = fmt.Sprintf("%s \"%s\"", appendList, mac)
		}
	}
	macFGList = append(macFGList, "end")
	macFGList = append(macFGList, "\nconfig firewall addrgrp")
	macFGList = append(macFGList, fmt.Sprintf("    edit \"%s\"", addGrp))
	macFGList = append(macFGList, fmt.Sprintf("        %s", appendList))
	macFGList = append(macFGList, "    next")
	macFGList = append(macFGList, "end")
	return macFGList
}

func validateMac(mac string) bool {
	reMac := regexp.MustCompile("^([0-9A-Fa-f]{2}[:]){5}([0-9A-Fa-f]{2})$")
	if match := reMac.FindStringSubmatch(mac); match != nil {
		return true
	} else {
		return false
	}
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
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			log.Println(err)
		}
	}
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

func readUserInput() []string {
	s := bufio.NewScanner(os.Stdin)

	var lines []string
	for {
		if !s.Scan() {
			break
		}
		lines = append(lines, s.Text())
	}

	err := s.Err()
	if err != nil {
		log.Fatal(err)
	}

	return lines
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
