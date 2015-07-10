// package main for jigsaw app.
// App reads and images and configuration and creates jigsaw type files and saves to disk
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/davidp15/jigsaw/jigsaw"
	"io/ioutil"
)

//main handles command line flags and begins the process
func main() {
	//read in initialization
	var init string
	var imageRootDirectory string
	flag.StringVar(&init, "init", "", "Initialization file location")
	flag.StringVar(&imageRootDirectory, "images", "", "Root Directory for images")
	flag.Parse()
	if init == "" {
		fmt.Println("An initialization file must be specified")
		return
	}
	fmt.Println("Opening", init)
	initFile, err := ioutil.ReadFile(init)
	if err != nil {
		fmt.Println("Could not open file", err.Error())
		return
	}
	//create the jigsaw object from the json init file
	fileContents := string(initFile)
	byt := []byte(fileContents)
	jigsaw := &Jigsaw.Jigsaw{}
	fmt.Println("Reading init")
	err = json.Unmarshal(byt, jigsaw)
	if err != nil {
		// An error occurred while converting our JSON to an object
		fmt.Println("Could not convert init file" + err.Error())
		return
	}
	if !jigsaw.Init(imageRootDirectory) {
		fmt.Println("Could not initialize jigsaw")
		return
	}
	//this channel will be used to communicate when images are created and ready for copying
	c1 := make(chan int)
	//channel to be used to communicate when copying is completed anf files are ready to be saved
	c2 := make(chan int)
	//channel to communicate when process is complete
	done := make(chan bool)
	//concurrently run these three functions
	go jigsaw.InitPieces(c1)
	go jigsaw.CreateImage(c1, c2)
	go jigsaw.SaveImage(c2, done)
	//wait until we get a done message
	<-done

	fmt.Println("Done")
}
