package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	Mim_add()
}

func Mim_init() {
	if d, err := os.Stat(".mim"); os.IsNotExist(err) || !d.IsDir() {
		//target
		inf := []string{".mim/index", ".mim/refs", ".mim/object"}
		for _, v := range inf {
			if err := os.MkdirAll(v, 0755); err != nil {
				log.Fatal(err)
			} else {
				log.Println("Make: ", v)
			}
		}
	} else {
		log.Println("Init Has Been")
	}

}

func Mim_add() {
	f, err := os.Open("./testfile.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	data := make([]byte, 1024)
	count, err := f.Read(data)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(count)
	fmt.Println(string(data[:count]))

}
