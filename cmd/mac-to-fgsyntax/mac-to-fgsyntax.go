package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"text/template"
	"time"
)

const path string = "tmp"

var tpl *template.Template
var convertedFiles = sync.Map{}

func init() {
	tpl = template.Must(template.ParseGlob("../../templates/*"))
}

func main() {
	port := flag.String("p", "8080", "server port")
	flag.Parse()

	http.HandleFunc("/", index)
	http.HandleFunc("/mac-to-fgsyntax", macFGsyntax)
	http.HandleFunc("/mac-to-fgsyntax/upload", uploadFile)
	http.HandleFunc("/mac-to-fgsyntax/download", downloadFile)
	http.Handle("/favicon.ico", http.NotFoundHandler())

	log.Printf("Serving on HTTP port: %s\n", *port)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}

func index(w http.ResponseWriter, r *http.Request) {
	tpl.ExecuteTemplate(w, "index.gohtml", nil)
}

func macFGsyntax(w http.ResponseWriter, r *http.Request) {
	tpl.ExecuteTemplate(w, "mac-to-fgsyntax.gohtml", nil)
}

func uploadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(10 << 20) // 10 MB limt
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	// retrieve address group name from form
	addrGrp := r.FormValue("addrGrp")
	if addrGrp == "" {
		http.Error(w, "Address group name is required", http.StatusBadRequest)
		return
	}

	// retrieve the file from form data
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving the file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Create temporary file in tmp folder, but first check if tmp folder eixsts and create if not.
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			log.Println(err)
		}
	}
	tempFile, err := os.CreateTemp(path, "upload-*.txt")
	if err != nil {
		http.Error(w, "Error creating temporary file", http.StatusInternalServerError)
		return
	}
	defer func() {
		tempFile.Close()
		os.Remove(tempFile.Name())
	}()

	// Copy uploaded file content to temp file
	_, err = io.Copy(tempFile, file)
	if err != nil {
		http.Error(w, "Error writing to temporary file", http.StatusInternalServerError)
		return
	}

	inTempFileName := tempFile.Name()
	outTempFileName := doConversion(inTempFileName, addrGrp)

	// Generate uuid for converted file
	uuid := fmt.Sprintf("%d", time.Now().UnixNano())
	convertedFiles.Store(uuid, outTempFileName)

	http.Redirect(w, r, "/mac-to-fgsyntax/download?fileID="+uuid, http.StatusSeeOther)
}

func downloadFile(w http.ResponseWriter, r *http.Request) {
	// get file uuid from query
	fileID := r.URL.Query().Get("fileID")
	if fileID == "" {
		http.Error(w, "Missing file ID", http.StatusBadRequest)
		return
	}

	// retrieve file from uuid map
	value, ok := convertedFiles.Load(fileID)
	if !ok {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	filePath := value.(string)

	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "Converted file not found", http.StatusInternalServerError)
		return
	}
	defer func() {
		file.Close()
		os.Remove(filePath)
		convertedFiles.Delete(fileID)
	}()

	// Get file info
	fileInfo, err := file.Stat()
	if err != nil {
		http.Error(w, "Could not get file info", http.StatusInternalServerError)
		return
	}

	// Set headers to prompt download
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileInfo.Name()))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))

	// serve file
	file.Seek(0, io.SeekStart)
	io.Copy(w, file)
}

func doConversion(inFileName string, addrGrpName string) string {
	macList := readTextFile(inFileName)
	macFGList := convertToFGsyntax(macList, addrGrpName)
	return writeTextFile("tmp", macFGList)
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

func writeTextFile(path string, text []string) string {
	f, err := os.CreateTemp(path, "converted-*.txt")
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

	return f.Name()
}
