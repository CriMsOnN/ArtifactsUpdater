package main

import (
	"encoding/json"
	"fmt"
	"github.com/crimsonn/artifacts_updater/cmds"
	"os"
	"strings"
)

var Intro = `
 █████  ██████  ████████ ██ ███████  █████   ██████ ████████ ███████     ██    ██ ██████  ██████   █████  ████████ ███████ ██████  
██   ██ ██   ██    ██    ██ ██      ██   ██ ██         ██    ██          ██    ██ ██   ██ ██   ██ ██   ██    ██    ██      ██   ██ 
███████ ██████     ██    ██ █████   ███████ ██         ██    ███████     ██    ██ ██████  ██   ██ ███████    ██    █████   ██████  
██   ██ ██   ██    ██    ██ ██      ██   ██ ██         ██         ██     ██    ██ ██      ██   ██ ██   ██    ██    ██      ██   ██ 
██   ██ ██   ██    ██    ██ ██      ██   ██  ██████    ██    ███████      ██████  ██      ██████  ██   ██    ██    ███████ ██   ██
`

func main() {
	fmt.Println(Intro)
	file, err := os.Open("config.json")

	if err != nil {
		fmt.Println("There was an error: ", err)
	}

	decoder := json.NewDecoder(file)
	configuration := cmds.Configuration{}
	err = decoder.Decode(&configuration)
	if err != nil {
		fmt.Println("There was an error: ", err)
	}

	var input string

	for {
		fmt.Println("Select the os for the artifacts ( W = win32, L = linux): ")
		_, err = fmt.Scanln(&input)

		if err != nil {
			fmt.Println("There was an error: ", err)
			break
		}
		lowerString := strings.ToLower(input)
		if lowerString == "w" || lowerString == "l" {
			break
		}
	}

	var downloadUrl string
	var input2 string

	for {
		fmt.Println("Select the artifact you want to download ( R = Recommended, O = Optional, L = Latest, C = Critical): ")
		_, err = fmt.Scan(&input2)

		if err != nil {
			fmt.Println("There was an error: ", err)
			break
		}

		lowerString := strings.ToLower(input2)
		if lowerString == "r" || lowerString == "o" || lowerString == "l" || lowerString == "c" {
			break
		}
	}

	var downloadType string

	if input2 == "" {
		panic("You have to select an artifact!")
	}

	switch strings.ToLower(input2) {
	case "r":
		downloadType = "Recommended_Download"
	case "o":
		downloadType = "Optional_Download"
	case "l":
		downloadType = "Latest_Download"
	case "c":
		downloadType = "Critical_Download"
	}

	if strings.ToLower(input) == "l" {
		downloadUrl = cmds.GetArtifactUrl(&configuration, "linux", downloadType)
		cmds.DownloadFile(downloadUrl, "linux")
	} else {
		downloadUrl = cmds.GetArtifactUrl(&configuration, "win32", downloadType)
		cmds.DownloadFile(downloadUrl, "win32")
	}
}
