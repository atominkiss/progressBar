package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Download struct {
	File          *os.File
	Response      *http.Response
	ContentLength int
	Done          bool
}

type ProgressBar struct {
	Download *Download
}

func main() {
	download := NewDownload("https://download.mozilla.org/?product=firefox-28.0-SSL&os=osx&lang=en-US",
		"firefox.dmg")
	progressBar := ProgressBar{download}
	progressBar.Start()
}

func (download *Download) StartDownload() {
	_, err := io.Copy(download.File, download.Response.Body)
	if err != nil {
		fmt.Println("Error:", err)
		download.Done = true
		return
	}
	download.Done = true
}

func NewDownload(url string, fileName string) *Download {
	// Выделяем память для нового объекта download
	download := new(Download)
	
	out, err := os.Create(fileName)
	if err != nil {
		fmt.Println("Error:", err)
	}
	download.File = out
	
	response, err := http.Get(url)
	if err != nil {
		fmt.Println("Error:", err)
	}
	
	download.Response = response
	download.ContentLength, _ = strconv.Atoi(response.Header.Get("content-length"))
	download.Done = false
	
	return download
}

func (download *Download) BytesDownloaded() int {
	info, err := download.File.Stat()
	if err != nil {
		fmt.Println("Error:", err)
		return 0
	}
	return int(info.Size())
}

func (progressBar *ProgressBar) Start() {
	// Запускаем загрузку в горутине
	go progressBar.Download.StartDownload()
	progressBar.Show()
	
	// Закрываем файл и соединение
	_ = progressBar.Download.Response.Body.Close()
	_ = progressBar.Download.File.Close()
}

func (progressBar *ProgressBar) Show() {
	var progress int
	totalBytes := int(progressBar.Download.ContentLength)
	lastTime := false
	
	// перерисовываем индикатор выполнения пока идет загрузка
	for !progressBar.Download.Done || lastTime {
		// Запускаем progressBar - с новой строки
		fmt.Print("\r[")
		bytesDone := progressBar.Download.BytesDownloaded()
		progress = 40 * bytesDone / totalBytes
		
		// рисуем progressBar
		for i := 0; i < 40; i++ {
			if i < progress {
				fmt.Print("=")
			} else if i == progress {
				fmt.Print(">")
			} else {
				fmt.Print(" ")
			}
		}
		fmt.Print("] ")
		//  сколько загрузили кВ
		fmt.Printf("%d/%dkB", bytesDone/1000, totalBytes/1000)
		// подождем 100 миллисекунд
		time.Sleep(100 * time.Millisecond)
		//
		// После завершения загрузки нам потребуется еще одна итерация цикла
		if progressBar.Download.Done && !lastTime {
			lastTime = true
		} else {
			lastTime = false
		}
	}
	
	fmt.Println()
}
