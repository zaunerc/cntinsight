package main

import (
	"bufio"
        "bytes"
	"fmt"
	"html/template"
        "io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"

        "github.com/shurcooL/github_flavored_markdown"
	"github.com/shurcooL/github_flavored_markdown/gfmstyle"
        //"github.com/russross/blackfriday"
)

type Page struct {
	Path     string
	Markdown []byte
	Html     []byte
}

type Log struct {
	Path       string
	ParentPath string
	Text       string
}

func convertMdToHtml(pathToFile string) (*Page, error) {

	pageAsMarkdown, error := ioutil.ReadFile(pathToFile)


        appendix := enrichMd(getReadmeAsMarkdown())
        // Make sure that there is at least one newline before our heading
        pageAsMarkdown = append(pageAsMarkdown, "\n"...)
        for _, appendixLine := range appendix {
            pageAsMarkdown = append(pageAsMarkdown, appendixLine + "\n"...)
        }
        
        pageAsHtml := github_flavored_markdown.Markdown(pageAsMarkdown)

	if error != nil {
		return nil, error
	}
        
	return &Page{Path: pathToFile, Markdown: pageAsMarkdown, Html: pageAsHtml}, nil
}

/**
 * Adds various informations to the README.md of the container.
 */
func enrichMd(rdmeAsMdArr []string) (enrichedRdmeAsMdArr []string) { 
    
	appendixTemplate, error := template.ParseFiles("appendix_template.md")
        var doc bytes.Buffer 
        error = appendixTemplate.Execute(&doc, nil)
        s := doc.String()
        
	if error != nil {
                fmt.Printf("Error while processing template: >%s<.", error)
		return rdmeAsMdArr
	}

	return append(enrichedRdmeAsMdArr, strings.Split(s, "\n")...)
}

// `/var/log/AdminServer` - sshd log file
func parseLogFilePathsFromMd(pathToFile string) (*[]string, error) {

	var foundPath []string

	pageAsMarkdown, error := os.Open(pathToFile)

	if error != nil {
		return nil, error
	}

	matcher := regexp.MustCompile("`(.*)`.*-.*log.*")
	scanner := bufio.NewScanner(pageAsMarkdown)
	for scanner.Scan() {
		line := scanner.Text()
		match := matcher.FindStringSubmatch(line)
		if len(match) > 0 {
			foundPath = append(foundPath, match[1])
			fmt.Printf("Found path >%s< in file >%s<.\n", match[1], pathToFile)
		}
	}

	fmt.Printf("Found %d paths in file >%s<\n", len(foundPath), pathToFile)
	return &foundPath, nil
}

func readLogFiles(pathToLogFiles *[]string) (*[]Log, error) {

	var logs []Log

	for _, pathToLogFile := range *pathToLogFiles {
		file, err := os.Open(pathToLogFile)

		if err != nil {
			fmt.Printf("Error while reading log file >%s<: >%s<\n", pathToLogFile, err)
		} else {
			fmt.Printf("Successfully read log file >%s<.\n", pathToLogFile)
		}

		defer file.Close()

		var lines []string
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}

		log := Log{Path: pathToLogFile, ParentPath: pathToLogFile, Text: strings.Join(lines, "\n")}
		logs = append(logs, log)
	}

	return &logs, nil
}

func init() {
	// By default logger is set to write to stderr device.
	//log.SetOutput(os.Stdout)
}

func main() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/log", logHandler)
        
        // Serve the "/assets/gfm.css" file.
        http.Handle("/assets/", http.StripPrefix("/assets", http.FileServer(gfmstyle.Assets)))
        
	http.ListenAndServe(":8080", nil)
}

func getReadmeAsMarkdown() []string {

	var possibleLocations [2]string = [2]string{"/README.md",
		"testdata/README.md"}

	var fd *os.File
	var err error
	for _, possibleLocation := range possibleLocations {
		fd, err = os.Open(possibleLocation)
		if err != nil {
			break
		}
	}
        defer fd.Close()
	
	if err == nil {
		fmt.Printf("Could not find README.md file.")
		os.Exit(-1)
	}

	var lines []string
	scanner := bufio.NewScanner(fd)
        for scanner.Scan() {
            lines = append(lines, scanner.Text())
        }
	
	return lines
}

func handler(w http.ResponseWriter, r *http.Request) {
	page, _ := convertMdToHtml("testdata/README.md")
        
        io.WriteString(w, `<html><head><meta charset="utf-8"><link href="/assets/gfm.css" media="all" rel="stylesheet" type="text/css" /><link href="//cdnjs.cloudflare.com/ajax/libs/octicons/2.1.2/octicons.css" media="all" rel="stylesheet" type="text/css" /></head><body><article class="markdown-body entry-content" style="padding: 30px;">`)
        w.Write(page.Html)
        io.WriteString(w, `</article></body></html>`)
}

func logHandler(w http.ResponseWriter, r *http.Request) {

	pathToLogFiles, _ := parseLogFilePathsFromMd("testdata/README.md")
	logs, _ := readLogFiles(pathToLogFiles)

	t, error := template.ParseFiles("log.html")
	error = t.Execute(w, logs)

	if error != nil {
		fmt.Printf("Error while processing template: >%s<.", error)
	}
}
