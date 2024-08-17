package config

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"time"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var (
	cfgFile  string
	dataFile string
)

func ConfigFilePath() string {
	return cfgFile
}

func DataFilePath() string {
	return dataFile
}

func SetConfigFilePath(p string) {
	cfgFile = p
}

func SetDataFilePath(p string) {
	dataFile = p
}

const (
	DefaultDateTimePattern = "02 Jan 2006 15:04"
	DefaultDatePattern     = "02 Jan 2006"
	DataFileHeader         = "id,date,activity,hours,jira"
)

func Init() {
	cfgFile = viper.GetString("config")
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		configPath := path.Join(home, ".jtl")
		configName := "config"
		configType := "yaml"
		configFullPath := fmt.Sprintf("%v.%v", path.Join(configPath, configName), configType)

		viper.AddConfigPath(configPath)
		viper.SetConfigName(configName)
		viper.SetConfigType(configType)

		viper.Set("dataFileHeader", "id,date,activity,hours,jira")
		viper.SetDefault("host", "")
		viper.SetDefault("credentials", map[string]string{
			"username": "",
			"password": "",
		})
		viper.SetDefault("DateTimePattern", DefaultDateTimePattern)

		if !fileExists(configFullPath) {
			fmt.Println("Config file not found. Initializing default config:", configFullPath)
			err := viper.SafeWriteConfig()
			if err != nil {
				fmt.Println("Error writing config file", err)
			}
		}
	}
	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Error reading config file", err)

	}
}

func InitDataFile() {
	dataFileHeader := viper.GetString("dataFileHeader")
	createNewDataFile := func() {
		dataDir := path.Join(homeDir(), ".jtl", "data")
		f := path.Join(dataDir, GenerateDataFileName())
		createDirIfNotExists(dataDir)
		createFileIfNotExists(f, dataFileHeader)
		//upgrade file to full path in context
		dataFile = f
	}

	dataFile = viper.GetString("data")
	if dataFile == "" {
		createNewDataFile()
	} else {
		if !fileExists(dataFile) {
			fmt.Println("Provided data file doesn't exist. Default one will be used.")
			createNewDataFile()
		}
	}
}

// GenerateDataFileName returns today's default datafile name without path.
func GenerateDataFileName() string {
	now := time.Now()
	return fmt.Sprintf("%v.csv", now.Format("Jan-2006"))
}

// GetCurrentDataFileName returns current datafile name without path.
func GetCurrentDataFileName() string {
	return dataFile[strings.LastIndex(dataFile, "/")+1:]
}

func createDirIfNotExists(path string) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(path, 0777)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func createFileIfNotExists(filename string, dataFileHeader string) {
	if !fileExists(filename) {
		f, err := os.Create(filename)
		defer func() {
			if err := f.Close(); err != nil {
				log.Fatal(err)
			}
		}()
		if err != nil {
			log.Fatal(err)
		}
		f.WriteString(dataFileHeader)
	}
}

func homeDir() string {
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return home
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
