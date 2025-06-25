package types

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/peteradeojo/gopvm/util"
)

type Config struct {
	UseCache             bool   `json:"use_cache"`
	InstallDir           string `json:"install_dir"`
	CurrentActiveVersion string `json:"current_active_version"`
	BinaryDir            string `json:"binary_dir"`
	CacheDir             string `json:"cache_dir"`
	file                 *os.File
	InstalledVersions    []string `json:"installed_versions"`
}

func (c *Config) Save() error {
	data, err := json.Marshal(c)
	if err == nil {
		c.file.Truncate(0)
		c.file.Seek(0, 0)

		_, err = c.file.Write(data)
		if err != nil {
			log.Println(err)
			panic(err)
		}
	} else {
		log.Println(err)
	}

	return err
}

func (c *Config) SetFile(file *os.File) {
	c.file = file
}

func (c *Config) MakeDirs() {
	if exists, _ := util.CheckDirExists(c.InstallDir); !exists {
		os.MkdirAll(c.InstallDir, 0755)
	}

	if exists, _ := util.CheckDirExists(c.CacheDir); !exists {
		os.MkdirAll(c.CacheDir, 0755)
	}
}

func (c *Config) ReadConfig(filename *string) {
	configFile := ".pvm/config"
	if filename != nil {
		configFile = *filename
	}

	configData, err := os.ReadFile(configFile)

	if err != nil {
		log.Fatalln(err)
	}

	err = json.Unmarshal(configData, c)
	if err != nil {
		fmt.Println(err)
		cwd, _ := os.Getwd()
		c.InstallDir = cwd + "/install/"
		c.UseCache = true
		c.CacheDir = cwd + "/cache/"
		c.InstalledVersions = make([]string, 0, 12)

		c.Save()
	}
}
