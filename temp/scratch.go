package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
)

func processDir(dirName string) {
	log.Println("processing dir:", dirName)
	d, err := os.Open(dirName)
	if err != nil {
		log.Println(err)
		return
	}
	defer d.Close()

	fNames, err := d.Readdirnames(0)
	if err != nil {
		log.Println(err)
		return
	}
	for _, name := range fNames {
		fName := filepath.Join(dirName, name)
		f, err := os.Open(fName)
		if err != nil {
			log.Println(err)
			// continue with other files
		}
		stat, err := f.Stat()
		if err != nil {
			log.Println(err)
			// continue with other files
		} else {
			if stat.IsDir() {
				processDir(fName)
			} else {
				log.Println("processing file:", fName)
			}
		}
	}

}

func main() {
	dirName := flag.String("dir", ".", "directory to traverse")
	flag.Parse()

	processDir(*dirName)
}
