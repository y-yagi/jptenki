package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/y-yagi/configure"
)

const cmd = "jptenki"

type config struct {
	Home   string            `toml:"home"`
	Places map[string]string `yaml:"places"`
}

var (
	red   = color.New(color.FgRed, color.Bold).SprintFunc()
	blue  = color.New(color.FgBlue, color.Bold).SprintFunc()
	white = color.New(color.FgWhite, color.Bold).SprintFunc()
)

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
	} else if class == ".precipitation" {
		*values = append(*values, "降水量")
	}
}

func convertWeatherToEmoji(weather string) string {
	switch weather {
	case "晴れ":
		return red("☀")
	case "曇り":
		return white("☁")
	case "雨":
		return blue("☂")
	case "強雨", "豪雨":
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
	var cfg config
	var area string

	err := configure.Load(cmd, &cfg)
	if err != nil {
		fmt.Printf("config file load Error: %v\nPlease create a config file.\n", err)
		os.Exit(1)
	}

	if len(os.Args) > 1 {
		area = os.Args[1]
	} else {
		area = cfg.Home
	}

	url = cfg.Places[area]
	if len(url) == 0 {
		fmt.Printf("'%s' is not defined. Please add alias to config file.\n", area)
		os.Exit(1)
	}

	res, err := http.Get(url)
	if err != nil {
		fmt.Printf("Page get error: %v.\n", err)
		os.Exit(1)
	}

	defer res.Body.Close()
	if res.StatusCode != 200 {
		fmt.Printf("Status code error: %d %s", res.StatusCode, res.Status)
		os.Exit(1)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Printf("Document read error: %v.\n", err)
		os.Exit(1)
	}

	var targetClasses = []string{".hour", ".weather", ".temperature", ".prob_precip", ".precipitation", ".wind-blow", ".wind-speed"}
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
