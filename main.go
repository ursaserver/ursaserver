package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/ursaserver/ursa"
)

func main() {
	portPtr := flag.Int("port", 3333, "server port")
	filePtr := flag.String("file", "conf.json", "configuration json file")
	flag.Parse()

	confFile, err := os.Open(*filePtr)
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
	// Check if the configuration is valid
	if err := CheckConf(&conf); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}
	// Convert the configuration got into ursa.Conf
	ursaConf := confToUrsaConf(conf)

	// Create a ursa reverse proxy based on the provided configuration
	rp := ursa.New(ursaConf)
	// Create HTTP server
	hostPort := fmt.Sprintf(":%v", *portPtr)
	fmt.Println("listening at", hostPort)
	http.ListenAndServe(hostPort, rp)
}
