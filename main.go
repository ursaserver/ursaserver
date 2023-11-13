package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ursaserver/ursa"
)

const MissingConfFileMessage = `missing configuration file name`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, MissingConfFileMessage)
		return
	}
	confFile, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}
	defer confFile.Close()
	var conf Conf
	if err := json.NewDecoder(confFile).Decode(&conf); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}
	if err := CheckConf(&conf); err != nil {
		fmt.Fprintf(os.Stderr, "error detected in configuration\n%v\n", err)
		return
	}
}

func ConfToUrsaConf(c Conf) ursa.Conf {
	var ursaConf ursa.Conf
	return ursaConf
}
