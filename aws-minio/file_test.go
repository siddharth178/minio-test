package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

const (
	TMP_DIR_ROOT     = "tmp"
	TMP_FILE         = "temp.txt"
	NON_EXISTING_DIR = "non existing dir"
)

var files = []string{
	filepath.Join(TMP_DIR_ROOT, "a", "0", "a0.txt"),
	filepath.Join(TMP_DIR_ROOT, "a", "0", "a1.txt"),
	filepath.Join(TMP_DIR_ROOT, "b", "0", "b0.txt"),
	filepath.Join(TMP_DIR_ROOT, "c0.txt"),
	filepath.Join(TMP_DIR_ROOT, "c1.txt"),
	filepath.Join(TMP_DIR_ROOT, "c2.txt"),
}

func exitWithError(e error) {
	fmt.Println(e)
	fmt.Println("Note: Make sure you manually clean things up before running the tests again.")
	os.Exit(1)
}

func createDirTree() {
	err := os.Mkdir(TMP_DIR_ROOT, 0755)
	if err != nil {
		exitWithError(err)
	}

	f, err := os.Create(filepath.Join(TMP_DIR_ROOT, TMP_FILE))
	if err != nil {
		exitWithError(err)
	}
	f.Close()

	// create complex dir tree
	for _, fName := range files {
		dirName := filepath.Dir(fName)
		err = os.MkdirAll(dirName, 0755)
		if err != nil {
			exitWithError(err)
		}
		f, err = os.Create(fName)
		if err != nil {
			exitWithError(err)
		}
		f.Close()
	}
}

func delDirTree() {
	err := os.RemoveAll(TMP_DIR_ROOT)
	if err != nil {
		fmt.Println("Could not exit with complete cleanup")
		fmt.Println(err)
		os.Exit(1)
	}
}

func TestMain(m *testing.M) {
	// setup
	fmt.Println("Test setup running")
	createDirTree()
	fmt.Println("Test setup done")

	// run tests
	testRes := m.Run()

	//tear down
	if testRes == 0 {
		fmt.Println("Test tear down running")
		delDirTree()
		fmt.Println("Test tear down done")
	} else {
		fmt.Println("Please cleanup manually before running tests again.")
	}
	// exit
	os.Exit(testRes)

}

func TestIsDir(t *testing.T) {
	ok, err := isDir(TMP_DIR_ROOT)
	if err != nil || !ok {
		t.Error(err)
		t.Error("Expecting a directory, detected non-directory.")
	}
}

func TestIsDirDoesntExists(t *testing.T) {
	ok, err := isDir(NON_EXISTING_DIR)
	if err == nil || ok {
		t.Error("Detected non existing directory as directory.")
	}
}

func TestIsDirWithFile(t *testing.T) {
	ok, err := isDir(filepath.Join(TMP_DIR_ROOT, TMP_FILE))
	if err != nil {
		t.Error(err)
	}
	if ok {
		t.Error("Detected file as directory.")
	}
}

func TestProcessDirP(t *testing.T) {
	fileNameChan := make(chan string, 100)

	quitChan := make(chan int)
	fileCountFromChan := 0
	go func() {
		for {
			select {
			case _ = <-fileNameChan:
				fileCountFromChan++
			case <-quitChan:
				break
			}
		}
		return
	}()

	fileCount := processDirP(TMP_DIR_ROOT, fileNameChan)

	// stop the listening goroutine
	quitChan <- 1

	// check if files sent for processing are all passed to chan
	if fileCount != fileCountFromChan {
		t.Error("Files processed:", fileCount, "doesn't match files sent over chan:", fileCountFromChan)
	} else {
		fmt.Println("Files processed:", fileCountFromChan)
	}
}
