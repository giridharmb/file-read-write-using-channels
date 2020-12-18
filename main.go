package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"
)

func getFileSize(filePath string) (int64, error) {
	fi, err := os.Stat(filePath)
	if err != nil {
		return -1, err
	}
	// get the size
	size := fi.Size()
	return size, nil
}

func readFile(name string) (byteCount int, buffer *bytes.Buffer) {

	const chunksize int = 400

	var count int
	data, err := os.Open(name)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = data.Close()
	}()

	reader := bufio.NewReader(data)
	buffer = bytes.NewBuffer(make([]byte, 0))
	part := make([]byte, chunksize)

	for {
		fmt.Printf("\n part => %v \n", part)
		count, err = reader.Read(part)
		if err != nil {
			break
		}
		buffer.Write(part[:count])
	}

	if err != io.EOF {
		log.Fatal("Error Reading ", name, ": ", err)
	} else {
		err = nil
	}
	byteCount = buffer.Len()
	return
}

func writeFile(name string, buffer *bytes.Buffer) {
	err := ioutil.WriteFile(name, buffer.Bytes(), 0755)
	if err != nil {
		fmt.Printf("\n(ERROR) : could not write file : %v", err.Error())
	}
}

type readBytesData struct {
	readBytes      []byte
	totalBytesRead int64
	totalFileSize  int64
}

func readFileViaChannel(name string, readBytes chan readBytesData) {
	fmt.Printf("\nreadFileViaChannel() ...")

	const chunksize int = 500

	size, _ := getFileSize(name)

	fmt.Printf("\nname => %v , size => %v", name, size)

	var count int
	data, err := os.Open(name)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = data.Close()
	}()

	reader := bufio.NewReader(data)
	part := make([]byte, chunksize)

	var counter int64
	counter = 0
	for {
		//fmt.Printf("\ncounter => %v", counter)
		time.Sleep(10 * time.Millisecond)
		count, err = reader.Read(part)
		if err != nil {
			close(readBytes)
			break
		}
		counter = counter + int64(count)
		myData := readBytesData{
			readBytes:      part[:count],
			totalBytesRead: counter,
			totalFileSize:  size,
		}
		readBytes <- myData
	}
}

func writeFileFromChannel(name string, readBytes chan readBytesData) {
	fmt.Printf("\nwriteFileFromChannel() ...")
	completeData := make([]byte, 0)

	for bytesData := range readBytes {
		for _, data := range bytesData.readBytes {
			completeData = append(completeData, data)
			fmt.Printf("\nwriteFileFromChannel() : totalFileSizeBytes => %v , bytesReadSoFar : %v", bytesData.totalFileSize, len(completeData))
		}
	}
	fmt.Printf("\nwriteFileFromChannel() : done collecting all bytes...")
	err := ioutil.WriteFile(name, completeData, 0755)
	if err != nil {
		fmt.Printf("\n(ERROR) : could not write to file via channel : %v", err.Error())
	}
}

func main() {
	/*
		fmt.Printf("\nReading file...")
		length, buffer := readFile("data.txt")
		fmt.Printf("\nFile Length:")
		fmt.Println(length)
		fmt.Printf("\nWriting File...")
		writeFile("data-write.txt", buffer)
	*/

	readBytesChan := make(chan readBytesData)

	fileName := "data.txt"
	writeFileNameViaChannel := "data-write-channel.txt"

	var wg sync.WaitGroup

	wg.Add(1)
	go func(wg *sync.WaitGroup, fileName string, readBytes chan readBytesData) {
		defer wg.Done()
		readFileViaChannel(fileName, readBytesChan)
	}(&wg, fileName, readBytesChan)

	wg.Add(1)
	go func(wg *sync.WaitGroup, fileName string, readBytes chan readBytesData) {
		defer wg.Done()
		writeFileFromChannel(fileName, readBytes)
	}(&wg, writeFileNameViaChannel, readBytesChan)

	wg.Wait()

}
