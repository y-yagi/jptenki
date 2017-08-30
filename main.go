package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"github.com/PuerkitoBio/goquery"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
)

type config struct {
	Home    string            `yaml:home`
	Aliases map[string]string `yaml:aliases`
}

var (
	red   = color.New(color.FgRed, color.Bold).SprintFunc()
	blue  = color.New(color.FgBlue, color.Bold).SprintFunc()
	white = color.New(color.FgWhite, color.Bold).SprintFunc()
)

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
	} else if class == ".wind-speed" {
		*values = append(*values, "")
	}
}

func convertWeatherToEmoji(weather string) string {
	switch weather {
	case "晴れ":
		return red("☀")
	case "曇り":
		return white("☁")
	case "雨", "強雨":
		return blue("☔")
	case "小雨", "弱雨":
		return blue("🌂")
	case "みぞれ":
		return white("❅")
	default:
		return weather
	}
}

func main() {
	var url string

	config, err := loadConfig()
	if err != nil {
		fmt.Printf("config file load Error: %v\nPlease create a config file.\n", err)
		os.Exit(1)
	}

	if len(os.Args) > 1 {
		url = config.Aliases[os.Args[1]]
		if len(url) == 0 {
			fmt.Printf("'%s' is not defined. Please add alias to config file.\n", os.Args[1])
			os.Exit(1)
		}
	} else {
		url = config.Home
	}

	doc, err := goquery.NewDocument(url)
	if err != nil {
		fmt.Printf("Document get error: %v.\n", err)
		os.Exit(1)
	}

	var targetClasses = []string{".hour", ".weather", ".temperature", ".prob_precip", ".wind-blow", ".wind-speed"}
	var tableIDs = []string{"#forecast-point-1h-today", "#forecast-point-1h-tomorrow", "#forecast-point-1h-dayaftertomorrow"}
	var values = []string{}
	w := os.Stdout
	table := tablewriter.NewWriter(w)

	for _, id := range tableIDs {
		doc.Find(id).Each(func(i int, s *goquery.Selection) {
			header := s.Find(".head td").Text()
			for _, class := range targetClasses {
				setTitle(&values, class)
				for _, value := range strings.Split(s.Find(class).Text(), "\n") {
					value := strings.TrimSpace(value)
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
}
