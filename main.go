package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"github.com/PuerkitoBio/goquery"
	"github.com/olekukonko/tablewriter"
)

type config struct {
	Home string `yaml:home"`
}

func loadConfig() (*config, error) {
	home := os.Getenv("HOME")
	if home == "" && runtime.GOOS == "windows" {
		home = os.Getenv("APPDATA")
	}

	fname := filepath.Join(home, ".config", "jptenki", "config.yml")
	buf, err := ioutil.ReadFile(fname)
	if err != nil {
		return nil, err
	}

	var cfg config
	err = yaml.Unmarshal(buf, &cfg)
	return &cfg, err
}

func showHeader(w io.Writer, header string) {
	fmt.Fprintf(w, "─────────────────────────────────────\n")
	fmt.Fprintf(w, "  %s\n", strings.TrimSpace(header))
	fmt.Fprintf(w, "─────────────────────────────────────\n")
}

func setTitle(values *[]string, class string) {
	if class == ".hour" {
		*values = append(*values, "時間")
	} else if class == ".temperature" {
		*values = append(*values, "気温")
	} else if class == ".windSpeed" {
		*values = append(*values, "")
	}
}

func convertWeatherToEmoji(weather string) string {
	switch weather {
	case "晴れ":
		return "☀"
	case "曇り":
		return "☁"
	case "雨":
		return "☔"
	default:
		return weather
	}
}

func main() {
	config, err := loadConfig()
	if err != nil {
		fmt.Printf("Config file load Error: %v\nPlease create a config file.\n", err)
		os.Exit(1)
	}

	doc, err := goquery.NewDocument(config.Home)
	if err != nil {
		log.Fatal(err)
	}

	var targetClasses = []string{".hour", ".weather", ".temperature", ".prob_precip", ".windBlow", ".windSpeed"}
	var values = []string{}
	w := os.Stdout
	table := tablewriter.NewWriter(w)

	doc.Find(".leisurePinpointWeather").Each(func(i int, s *goquery.Selection) {
		header := s.Find(".head td").Text()
		for _, class := range targetClasses {
			setTitle(&values, class)
			for _, value := range strings.Split(s.Find(class).Text(), "\n") {
				if len(value) > 0 {
					if class == ".weather" {
						values = append(values, convertWeatherToEmoji(value))
					} else {
						values = append(values, value)
					}
				}
			}
			table.Append(values)
			values = nil
		}

		showHeader(w, header)
		table.Render()
		fmt.Fprintf(w, "\n")
		table = tablewriter.NewWriter(w)
	})
}
