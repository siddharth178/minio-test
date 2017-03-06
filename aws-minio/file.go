package main

import (
	"log"
	"os"
	"path/filepath"
)

// Recursively iterate through all the files in given directory
func processDirP(dirName string, fileNameChan chan string, errorChan chan error) (fileCount int, err error) {
	d, err := os.Open(dirName)
	if err != nil {
		return 0, err
	}
	defer d.Close()

	fNames, err := d.Readdirnames(0)
	if err != nil {
		return 0, err
	}
	for _, name := range fNames {
		fName := filepath.Join(dirName, name)
		fc, err := processFileP(fName, fileNameChan, errorChan)
		if err != nil {
			errorChan <- err
			// continue with other files and directories in this
			// directory
		} else {
			fileCount += fc
		}
	}
	return fileCount, nil
}

// If current object is file, send it to uploadWorker through file name chan, for directory, just go inside that
// directory and look for more files.
func processFileP(fName string, fileNameChan chan string, errorChan chan error) (fileCount int, err error) {
	f, err := os.Open(fName)
	if err != nil {
		log.Println("error in open:", err)
		return 0, err
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		log.Println("error in stat:", err)
		return 0, err
	} else {
		if stat.IsDir() {
			fc, err := processDirP(fName, fileNameChan, errorChan)
			if err != nil {
				log.Println("error occured while processing dir:", err)
				return 0, err
			} else {
				fileCount += fc
			}
		} else {
			// send absolute path on channel so that s3 doesn't complain about
			// files with keys containing . or .. or something like that
			absPath, err := filepath.Abs(fName)
			if err != nil {
				log.Println(err)
				return 0, err
			} else {
				log.Println("Sending:", absPath)
				fileNameChan <- absPath // this will block if fileNameChan is full
				fileCount++
			}
		}
	}
	return fileCount, nil
}

// check if object pointed by given name is directory
func isDir(fileName string) (bool, error) {
	dir, err := os.Open(fileName)
	if err != nil {
		return false, err
	}
	defer dir.Close()

	fInfo, err := dir.Stat()
	if err != nil {
		return false, err
	}
	return fInfo.IsDir(), nil
}
