package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const assetsPathEnv string = "GO_ASSETS"
const fileIn = "mac-to-fgsyntax/in.txt"
const fileOut = "mac-to-fgsyntax/out.txt"

func main() {
	// Welcome message
	fmt.Printf("------------------------------------------------------------------------------\n")
	fmt.Printf("---------------- MAC address to FortiGate syntax converter -------------------\n")
	fmt.Printf("------------------------------------------------------------------------------\n")
	fmt.Printf("Program will extract any mac addresses from the text entered.\n")
	fmt.Printf("Supported mac address formats are:\n")
	fmt.Printf("00:00:5e:00:53:01, 00-00-5e-00-53-01 and 0000.5e00.5301\n")
	fmt.Printf("------------------------------------------------------------------------------\n\n")

	// First get address group name, common for both methods.
	fmt.Println("Enter address group name that mac address objects should be added to:")
	addrGrp := readUserInputSingle()

	// Select usage method
	for {
		usrChoice := selectMethod()
		switch usrChoice {
		case 1: // cli in/out method
			fmt.Printf("Enter the list of mac addresses and then press CTRL+D (or CTRL+Z if using Windows):\n")
			userMacInput := readUserInput()
			macList := parseUserInput(userMacInput)
			macFGList := convertToFGsyntax(macList, addrGrp)

			fmt.Printf("\nList converted to Fortigate Syntax: \n")
			fmt.Printf("--------------------------------------\n")
			for _, l := range macFGList {
				fmt.Println(l)
			}
			fmt.Printf("\nPress any button to exit\n")
			readUserInputSingle()
			os.Exit(0)
		case 2: // file in/out method
			// Init file path
			assetsPath, err := initEnv(assetsPathEnv)
			if err != nil {
				fmt.Println(err, "\nExiting program...")
				os.Exit(0)
			}
			fullInPath := filepath.Join(assetsPath, fileIn)
			fullOutPath := filepath.Join(assetsPath, fileOut)

			// read input file, parse and write output file.
			userMacInput := readTextFile(fullInPath)
			macList := parseUserInput(userMacInput)
			macFGList := convertToFGsyntax(macList, addrGrp)

			writeTextFile(fullOutPath, macFGList)
			fmt.Printf("List converted to FortiGate syntax and written to %s\n", fullOutPath)
			fmt.Printf("\nPress any button to exit\n")
			readUserInputSingle()
			os.Exit(0)
		default:
			fmt.Printf("Invalid choice, try again..\n\n")
		}
	}
}

func initEnv(envName string) (string, error) {
	env := os.Getenv(assetsPathEnv)
	var err error = nil
	if env == "" {
		err = fmt.Errorf("environment variable: \"%s\" is not set", envName)
	}
	return env, err
}

func parseUserInput(userInput []string) []net.HardwareAddr {
	var macList []net.HardwareAddr
	for _, l := range userInput {
		parts := strings.Fields(l)
		for _, part := range parts {
			mac, err := net.ParseMAC(part)
			if err == nil {
				macList = append(macList, mac)
			}
		}
	}
	return macList
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

func convertToFGsyntax(macList []net.HardwareAddr, addGrp string) []string {
	var macFGList []string
	sectionOffset := "    "
	appendList := "append member "
	macFGList = append(macFGList, "config firewall address")
	for _, hwAddr := range macList {
		mac := hwAddr.String()
		macFGList = append(macFGList, fmt.Sprintf("%sedit \"%s\"", sectionOffset, mac))
		macFGList = append(macFGList, fmt.Sprintf("%[1]s%[1]sset type mac", sectionOffset))
		macFGList = append(macFGList, fmt.Sprintf("%[1]s%[1]sset start-mac %s", sectionOffset, mac))
		macFGList = append(macFGList, fmt.Sprintf("%[1]s%[1]sset end-mac %s", sectionOffset, mac))
		macFGList = append(macFGList, fmt.Sprintf("%snext", sectionOffset))
		appendList = fmt.Sprintf("%s \"%s\"", appendList, mac)
	}
	macFGList = append(macFGList, "end")
	macFGList = append(macFGList, "\nconfig firewall addrgrp")
	macFGList = append(macFGList, fmt.Sprintf("%sedit \"%s\"", sectionOffset, addGrp))
	macFGList = append(macFGList, fmt.Sprintf("%[1]s%[1]s%s", sectionOffset, appendList))
	macFGList = append(macFGList, fmt.Sprintf("%snext", sectionOffset))
	macFGList = append(macFGList, "end")
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
	if s.Scan() {
		return s.Text()
	}
	if err := s.Err(); err != nil {
		log.Fatal(err)
	}
	return ""
}
