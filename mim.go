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

// ToDo
//
//コミットしていないとき = 参照すべき以前のINDEXファイルがない
// → すべてのファイルをcreateステータスで書き込む
//
//コミットしているとき = 参照すべき以前のINDEXファイルが有る
// → 以前のINDEXファイルと比較し、INDEXファイルをすべて書き換える
// このとき参照していた以前のINDEXの情報を書き込みたい
//
// MimAdd の中で旧INDEXの情報と、新INDEXの情報を管理する。
// makeBlobでfilenameとhashを返し、filenameとhashを旧indexと比較して、ステータス情報を確定
// それを新INDEX(slice)でappend
// Addの最後にINDEXに書き込み

//Init コマンド .mim配下のフォルダを作成する。
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

//Add コマンド
func MimAdd() {
	log.Println("==== Add Start ====")
	initIndex()
	indexMap := readIndex()

	var newIndexData []byte

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
			hash := makeBlob(path)
			status := getBlobStatus(indexMap, path, hash)
			c := bytes.Join(
				[][]byte{[]byte(status), []byte(" "), //status
					[]byte(path), []byte(" "), //path
					[]byte(fmt.Sprintf("%x\n", hash))}, //hash
				[]byte(""))
			newIndexData = append(newIndexData, c...)
		}
		return err
	})

	writeIndex(newIndexData)
	log.Printf("==== Add Finish ====\n\n")
}

//Blob を作る
func makeBlob(filename string) (hash [28]byte) {
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
	hash = sha256.Sum224(blobContent)

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

	return hash
}

//INDEXファイルを初期化する。INDEXファイルが存在しなければ作成、存在すればログを吐くだけ。
func initIndex() {
	indexFilename := ".mim/INDEX"
	if f, err := os.Stat(indexFilename); os.IsNotExist(err) || f.IsDir() {
		indexFile, err := os.OpenFile(indexFilename, os.O_CREATE, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		} else {
			log.Fatal(err)
			log.Println("INIT INDEX")
		}
		defer indexFile.Close()
	} else {
		log.Println("Has Been Exist Index")
	}
}

//現在あるINDEXファイルを読み込み、mapで返す。
func readIndex() (indexMap map[string]string) {
	indexMap = make(map[string]string)
	indexFilename := ".mim/INDEX"

	indexFile, err := os.Open(indexFilename)
	if err != nil {
		log.Fatal(err)
	}
	defer indexFile.Close()

	scanner := bufio.NewScanner(indexFile)
	for scanner.Scan() {
		fileContent := scanner.Text()
		if fileContent != "" {
			dataArr := strings.Split(fileContent, " ")
			// status path hashで構成されているはず。statusを消すなら要変更
			indexMap[dataArr[1]] = dataArr[2]
		}
	}
	return
}

//INDEXファイルを上書きする。
func writeIndex(indexData []byte) {
	indexFilename := ".mim/INDEX"
	indexFile, err := os.OpenFile(indexFilename, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	defer indexFile.Close()

	writer := bufio.NewWriter(indexFile)
	if _, err := writer.Write(indexData); err != nil {
		log.Fatal(err)
	} else {
		log.Printf("Make INDEX")
	}
	writer.Flush()
}

//indexのfilenameとhashを比較して、更新されているのかどうかを返す
func getBlobStatus(indexMap map[string]string, filename string, hash [28]byte) (status string) {

	if h, ok := indexMap[filename]; ok {
		if h == fmt.Sprintf("%x", hash) {
			status = "notChange"
		} else {
			status = "Change"
		}
	} else {
		status = "Create"
	}

	return
}

func main() {
	MimInit()
	MimAdd()
}
