package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/muniverse"
)

func main() {
	var dirPath string
	flag.StringVar(&dirPath, "dir", "../assets/screenshots", "destination directory")
	flag.Parse()

	stat, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		if err := os.Mkdir(dirPath, 0755); err != nil {
			essentials.Die(err)
		}
	} else if err != nil {
		essentials.Die(err)
	} else if !stat.IsDir() {
		essentials.Die("not a directory:", dirPath)
	}

	for _, spec := range muniverse.EnvSpecs {
		path := filepath.Join(dirPath, spec.Name+".png")
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			log.Println("Skipping", spec.Name)
			continue
		}
		log.Println("Capturing", spec.Name)
		env, err := muniverse.NewEnv(spec)
		if err != nil {
			essentials.Die(err)
		}
		if err := env.Reset(); err != nil {
			env.Close()
			essentials.Die(err)
		}
		if _, _, err := env.Step(time.Second); err != nil {
			env.Close()
			essentials.Die(err)
		}
		obs, err := env.Observe()
		env.Close()
		if err != nil {
			essentials.Die(err)
		}
		data, err := muniverse.ObsPNG(obs)
		if err != nil {
			essentials.Die(err)
		}
		if err := ioutil.WriteFile(path, data, 0755); err != nil {
			essentials.Die(err)
		}
	}
}
