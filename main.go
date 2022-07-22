package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// I use the stack of constants for clarity
const (
	JSON_PARAM_LENGTH = 2
	LEFT_BOX          = "╟"
	RIGHT_BOX         = "╢"
	ASCII_RED         = "\x1b[31m"
	ASCII_LIGHT_BLUE  = "\x1b[94m"
	ASCII_RESET       = "\x1b[0m"
	ASCII_CYAN        = "\x1b[36m"
	EMOJI             = "┌( ಠ‿ಠ)┘"
	TAG_FETCH_URL     = "https://api.github.com/repos/michalspano/wlpr/tags"
)

/* TODO: fade in/out the wallpaper
- currently not supported by osascript;
- linked article: https://discussions.apple.com/thread/250539801 */

func main() {
	var SCRIPT_SRC, IMG_DIR string // initial path buffers
	var conf map[string]string     // buffer to store the map
	var showFooter bool = true     // show the message by default

	/* Optional flags
	-nm: display no ending message
	-v: display the current version at remote GitHub reposiroty
	 We parse the optional flags before the execution of the program */

	args := os.Args[1:]
	if len(args) == 1 {
		switch args[0] {
		case "-nm", "--no-message":
			showFooter = false
		case "-v", "--version":
			fmt.Println(fetchCurrentVersion())
			return
		default:
			raiseError("Invalid flag")
		}
	}

	HOME, _ := os.UserHomeDir() // get the $HOME directory
	confPath := HOME + "/.wlpr.json"

	// open the config file
	file, err := openFile(confPath)
	if err != nil {
		raiseError("No config file found, run './init'")
	}

	// read the config file
	data, err := readFile(file)
	if err != nil {
		raiseError("Error reading config file")
	}

	// unmarshal the json
	json.Unmarshal(data, &conf)

	if len(conf) != JSON_PARAM_LENGTH {
		raiseError("Config file is invalid, run ./init")
	}

	// evaluate the JSON parsed paths
	IMG_DIR, SCRIPT_SRC = conf["src_path"], conf["root"]

	/* We check whether `IMG_DIR` exits, if not, we warnt the user.
	   We dont' need to chcek if `SCRIPT_SCR` exits, because it's
	   been created by the init script. */

	if _, err := os.Stat(IMG_DIR); os.IsNotExist(err) {
		raiseError("Directory " + IMG_DIR + " does not exist")
	}

	// get the current wallpaper using osascript
	currentWallpaper, err := getCurrentWallpaper(SCRIPT_SRC + "current_wallpaper.scpt")
	if err != nil {
		raiseError("Error getting current wallpaper")
	}

	buff := getFiles(IMG_DIR)
	rand.Seed(time.Now().UnixNano()) // seed the random generator

	// append a random picture from the selected directory
	imgPath := IMG_DIR + buff[rand.Intn(len(buff))]

	// we would like to change the wallpaper, such that it won't be the current one
	for imgPath == currentWallpaper {
		imgPath = IMG_DIR + buff[rand.Intn(len(buff))]
	}

	// set the wallpaper
	exec.Command("/bin/bash", "-c", SCRIPT_SRC+"setter.scpt"+" "+imgPath).Run()

	// display the footer (indicates success) [default behavior]
	if showFooter {
		displayFooter()
	}
}

// raise a formatted error to the console and abort the program
func raiseError(errorMessage string) {
	fmt.Printf("%s%s%s\n", ASCII_RED, errorMessage, ASCII_RESET)
	os.Exit(1)
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

func getTerminalWidth() (int, error) {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	out, _ := cmd.Output()

	// obtain an array with: [width height]
	dims := strings.Split(string(out[:len(out)-1]), " ")

	// return the height (second element) parsed as an int
	return strconv.Atoi(dims[1])
}

func displayFooter() {
	// predefined footer buffers
	startMsg, endMsg := "Changed successfully!", LEFT_BOX+" wlpr "+RIGHT_BOX

	// print the first part of the line
	fmt.Printf("%s%s %s%s", ASCII_CYAN, EMOJI, ASCII_RESET, startMsg)

	width, err := getTerminalWidth()
	if err != nil {
		raiseError("Error getting terminal size")
	}

	// the number of (visual) pixels used by the predefined buffers
	stdinLen := len(startMsg) + len(EMOJI) - 10

	/* The number of pixels filled with empty spaces, such that the
	ending part of the footer will be displayed to the complete right
	of the width of the terminal session */

	relativeLen := width - stdinLen - len(endMsg) + 2 // +2 extra padding
	for i := 0; i < relativeLen; i++ {
		fmt.Printf(" ")
	}

	// display the ending part; add new line
	fmt.Printf("%s%s%s\n", ASCII_LIGHT_BLUE, endMsg, ASCII_RESET)
}

// fetch the applicaitons current tag via GitHub
func fetchCurrentVersion() string {
	var tags []map[string]string // form: [{string: string}]
	response, _ := http.Get(TAG_FETCH_URL)
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		raiseError("Error fetching current version")
	}
	json.Unmarshal(body, &tags) // unmarshal the json

	/* we assume that the latest tag is defined at the first index,
	given by the key-value "name" in the dictionary */

	return tags[0]["name"]
}
