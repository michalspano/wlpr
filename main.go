package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"time"
)

/* TODO: fade in/out the wallpaper
- currently not supported by osascript;
- linked article: https://discussions.apple.com/thread/250539801 */

func main() {
	HOME, _ := os.UserHomeDir() // get the $HOME directory
	DIR, _ := os.Getwd()        // get the current directory

	// Set up the paths
	confPath := HOME + "/.wlpr.json"
	SCRIPT_SRC := DIR + "/scripts/"

	file, err := openFile(confPath)
	if err != nil {
		fmt.Println("No config file found, run './init'")
		os.Exit(1)
	}

	var conf map[string]string // buffer to store the map

	// read and unmarshal JSON config file
	data, err := readFile(file)
	json.Unmarshal(data, &conf)

	// assume that there is only one configuration option
	if len(conf) != 1 {
		fmt.Println("Config file is invalid, run ./init")
		os.Exit(1)
	}

	// append the directory and check if it exists
	IMG_DIR := conf["src_path"]
	if _, err := os.Stat(IMG_DIR); os.IsNotExist(err) {
		fmt.Printf("Directory %s does not exist\n", IMG_DIR)
		os.Exit(1)
	}

	// get the current wallpaper using osascript
	currentWallpaper, err := getCurrentWallpaper(SCRIPT_SRC + "current_wallpaper.scpt")
	if err != nil {
		fmt.Println("Error getting current wallpaper")
		os.Exit(1)
	}

	buff := getFiles(IMG_DIR)
	rand.Seed(time.Now().UnixNano()) // seed the random generator

	// append a random picture from the selected directory
	imgPath := IMG_DIR + "/" + buff[rand.Intn(len(buff))]

	// we would like to change the wallpaper, such that it won't be the current one
	for imgPath == currentWallpaper {
		imgPath = IMG_DIR + "/" + buff[rand.Intn(len(buff))]
	}

	// set the wallpaper
	exec.Command("/bin/bash", "-c", SCRIPT_SRC+"setter.scpt"+" "+imgPath).Run()
}

// open a file and catch errors
func openFile(path string) (*os.File, error) {
	return os.Open(path)
}

func readFile(file *os.File) ([]byte, error) {
	return ioutil.ReadAll(file)
}

// get the possible pictures in the directory
func getFiles(path string) []string {
	var buff []string
	files, _ := ioutil.ReadDir(path)
	for _, f := range files {
		if f.IsDir() || f.Name()[0] == '.' { // omit dirs and dotfiles
			continue
		}
		buff = append(buff, f.Name())
	}
	return buff
}

func getCurrentWallpaper(script string) (string, error) {
	cmd := exec.Command("/bin/bash", "-c", script)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out[:len(out)-1]), nil
}
