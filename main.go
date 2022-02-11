package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/briandowns/spinner"
)

func runCommand(paths []string) string {
	wg := &sync.WaitGroup{}

	flag := 0

	for _, path := range paths {
		wg.Add(1)

		go func(path string) {
			//log.Println(runtime.NumGoroutine()) // Number of goroutine
			defer wg.Done()
			ext := strings.LastIndex(path, ".") // extension (.pdf)

			if path[ext:] == ".pdf" {
				flag = 1
				pdfDir := filepath.Dir(path)            // Directory of pdf file
				filename := getFileNameWithoutExt(path)

				// Run qpdf command
				cmd := exec.Command("qpdf", "--qdf", path, pdfDir+"/"+"copy_"+filename+".pdf")
				err := cmd.Run()
				if err != nil {
					panic(err)
				}

				fmt.Println("\n", path)

				// Output command status
				state := cmd.ProcessState
				fmt.Printf("  %s\n", state.String())               // Exit code and state
				fmt.Printf("    Pid: %d\n", state.Pid())           // Process ID
				fmt.Printf("    System: %v\n", state.SystemTime()) // System time (the time of processing done in the kernel)
				fmt.Printf("    User: %v\n", state.UserTime())     // User time (time consumed in the process)

				// Delete the original file
				if err := os.Remove(path); err != nil {
					fmt.Println(err)
				}

				// Rename
				if err := os.Rename(pdfDir+"/"+"copy_"+filename+".pdf", path); err != nil {
					fmt.Println(err)
				}
			}
		}(path)
	}
	wg.Wait()

	// If there is no PDF, exit.
	if flag == 0 {
		fmt.Println("PDF file is missing.")
		os.Exit(1)
	}

	return "\nDone."
}

func getFileNameWithoutExt(path string) string {
	return filepath.Base(path[:len(path)-len(filepath.Ext(path))])
}

func dirwalk(dir string) []string {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	var paths []string
	for _, file := range files {
		if file.IsDir() {
			paths = append(paths, dirwalk(filepath.Join(dir, file.Name()))...)
			continue
		}
		paths = append(paths, filepath.Join(dir, file.Name()))
	}

	return paths
}

func main() {
	// If the first argument is -h or --help, output Usage, etc. and exit.
	if len(os.Args) == 2 {
		help := os.Args[1]
		if help == "-h" || help == "--help" {
			fmt.Println(`USAGE
  $> pdf2images_concurrency.exe <DIR>
DESCRIPTION
  Remove the PDF text copy guard.
  It is recursively processed with the directory specified as the first argument as the root.
OPTION
  -h or --help
REQUIREMENTS
  Windows
INSTALLATION
  Copy the pdf_text_copy_guard_canceller folder to any local location.
AUTHOR
  Kenta Goto`)
			os.Exit(1)
		}
	}

	// Exit for all but one argument
	if len(os.Args) != 2 {
		fmt.Println("The number of arguments specified is incorrect. Only one argument is allowed.")
		os.Exit(1)
	}

	dir := os.Args[1]     // First argument (root directory to be processed)
	paths := dirwalk(dir) // Go to read the root directory recursively

	// If there are no files, exit.
	if paths == nil {
		fmt.Println("File is missing.")
		os.Exit(1)
	}

	fmt.Println("Processing...")

	// Progressbar
	s := spinner.New(spinner.CharSets[36], 100*time.Millisecond)
	s.Color("green")
	s.Start()

	// Run command
	result := runCommand(paths)

	s.Stop() // End of progress bar

	fmt.Println(result)
}
