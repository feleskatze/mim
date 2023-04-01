package main

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"crypto/sha256"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func MimInit() {
	inf := []string{".mim", ".mim/refs", ".mim/object"}

	log.Println("==== INIT Start ====")

	for _, v := range inf {
		if d, err := os.Stat(v); os.IsNotExist(err) || !d.IsDir() {
			if err := os.MkdirAll(v, os.ModePerm); err != nil {
				log.Fatal(err)
			} else {
				log.Println("INIT: make ", v)
			}
		} else {
			log.Println("INIT: Has Been Exist ", v)
		}
	}
	log.Printf("==== INIT Finish ====\n\n")
}

func MimAdd() {
	log.Println("==== Add Start ====")

	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() && d.Name() == ".mim" {
			return filepath.SkipDir
		}
		if d.IsDir() && d.Name() == ".git" {
			return filepath.SkipDir
		}

		if !d.IsDir() {
			path, _ := filepath.Rel(dir, path)
			status, path, hash := makeBlob(path)
			writeIndex(status, path, hash)
		}
		return err
	})

	log.Printf("==== Add Finish ====\n\n")
}

func makeBlob(filename string) (string, string, [28]byte) {
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	//read target file
	var data []byte
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		data = append(data, scanner.Bytes()...)
	}

	blobContent := bytes.Join([][]byte{[]byte("blob "), []byte("\x00"), data}, []byte(""))

	//compress zlib
	var com bytes.Buffer
	compressor, _ := zlib.NewWriterLevel(&com, zlib.BestCompression)
	compressor.Write(blobContent)
	compressor.Close()

	//sha224
	hash := sha256.Sum224(blobContent)

	//open write file
	blobFile, err := os.Create(fmt.Sprintf(".mim/object/%x", hash))
	if err != nil {
		log.Fatal(err)
	}
	defer blobFile.Close()

	//make blob file
	writer := bufio.NewWriter(blobFile)
	if _, err := writer.Write(com.Bytes()); err != nil {
		log.Fatal(err)
	} else {
		log.Printf("Make Blob: %s : %x\n", filename, hash)
	}
	writer.Flush()

	status := "create"
	path := filename
	return status, path, hash

}

func writeIndex(status string, path string, hash [28]byte) {
	//Write IndexFile
	indexFile, err := os.OpenFile(".mim/index", os.O_WRONLY|os.O_APPEND|os.O_CREATE, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	defer indexFile.Close()

	hash16x := fmt.Sprintf("%x\n", hash)
	indexContent := bytes.Join([][]byte{[]byte(status), []byte(" "), []byte(path), []byte(" "), []byte(hash16x)}, []byte(""))

	indexWrite := bufio.NewWriter(indexFile)
	if _, err := indexWrite.Write(indexContent); err != nil {
		log.Fatal(err)
	} else {
		log.Printf("Add INDEX: %s : %s\n", path, status)
	}
	indexWrite.Flush()
}

func MimCommit() {

}

func test(filename string, hash string) (status string) {

	var indexData [][]string

	f, err := os.Open(".mim/Index")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fileContent := scanner.Text()
		dataArr := strings.Split(fileContent, " ")
		indexData = append(indexData, dataArr)
	}
	for i := 0; i < len(indexData); i++ {
		fmt.Println(indexData[i][1])
		if filename == indexData[i][1] {
			if hash == indexData[i][2] {
				status = "noChange"
			} else {
				status = "Change"
			}
		} else {
			status = "create"
		}
	}
	return
}

func main() {
	status := test("testfile.txt", "186326af9ecd40de5cb11b6655c00ffb5696677335e0bd23cc888c4b")
	fmt.Println(status)
	// MimInit()
	// MimAdd()
}
