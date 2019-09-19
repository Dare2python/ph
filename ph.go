package main

import (
	"archive/zip"
	"fmt"
	"github.com/h2non/filetype"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/mknote"
	"io"
	"log"
	"os"
	"path/filepath"
	"unicode/utf8"
)

func check(e error) {
	if e != nil {
		log.Print(e)
	}
}

func visit() filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		check(err)
		fmt.Print(info.Name())
		if info.IsDir() {
			fmt.Println("\t\t\t\tDIR")
			return nil
		}
		r, _ := utf8.DecodeRuneInString(info.Name())
		if r == '.' {
			fmt.Println("\t\t\thidden")
			return nil
		}
		head := make([]byte, 261)
		file, e := os.Open(path)
		check(e)
		_, e = file.Read(head)
		check(e)
		if ! filetype.IsArchive(head) {
			fmt.Println("\t\tis NOT archive. Skipping ...")
			return nil
		}
		fmt.Println("\t\tis Archive. Working ...")
		e = Unzip(path)
		check(e)

		//testExif()

		return nil
	}
}

func testExif() error {
	test := "/Users/serz/Documents/ph/img_2219_12004093673_o.jpg"
	p, err := os.Open(test) //photo
	exif.RegisterParsers(mknote.All...)
	x, err := exif.Decode(p)
	check(err)

	tm, err := x.DateTime()
	check(err)
	fmt.Println("Taken: ", tm)

	//lat, long, err := x.LatLong()
	//check(err)
	//fmt.Println("lat, long: ", lat, ", ", long)

	return nil
}

func Unzip(src string) error {
	a, err := zip.OpenReader(src) //archive
	check(err)
	defer a.Close()
	for _, f := range a.File { // file
		//fmt.Print(f.Name)
		fpath := filepath.Join(filepath.Dir(src), f.Name)
		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		//if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
		//	return filenames, fmt.Errorf("%s: illegal file path", fpath)
		//}
		if f.FileInfo().IsDir() {
			// Make Folder
			err = os.MkdirAll(fpath, os.ModePerm)
			check(err)
			continue
		}
		if f.FileInfo().Size() == 0 {
			// empty file :(
			log.Print(f.Name + " is empty")
			continue
		}
		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		check(err)

		p, err := f.Open() //photo
		check(err)
		//_, err = io.CopyN(os.Stdout, p, 10)
		//check(err)
		_, err = io.Copy(outFile, p)
		check(err)

		//_, err = outFile.Seek(0, 0)
		//check(err)
		outFile.Close()
		outFile, err = os.OpenFile(fpath, os.O_RDWR, f.Mode())
		check(err)

		exif.RegisterParsers(mknote.All...)
		x, err := exif.Decode(outFile)
		check(err)

		newName := f.Name
		monthFolder := "no_date"

		if x != nil{
			tm, err := x.DateTime()
			if err == nil {
				newName = tm.Format("2006-01-02_1504_05.000") + filepath.Ext(fpath)
				monthFolder = tm.Format("2006-01")
			}
		}


		//fmt.Println(" > ", monthFolder, newName)
		fmt.Print(".")

		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		p.Close()

		fullMonthFolder := filepath.Join(filepath.Join(filepath.Dir(fpath), ".."), monthFolder)

		err = os.MkdirAll(fullMonthFolder, os.ModePerm)
		if err != nil {
			fmt.Println(err)
		}

		os.Rename(fpath, filepath.Join(fullMonthFolder, newName))

	}
	return nil
}

func main() {
	root := "/Users/serz/Documents/ph/"
	if len(os.Args) > 1 {
		root = os.Args[1]
	}

	err := filepath.Walk(root, visit())
	check(err)
}
