package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

var res string

func main() {
	var url string

	fmt.Println("FB Video Downloader")

	flag.StringVar(&url, "url", "", "Specify Facebook Video URL")
	flag.StringVar(&res, "r", "hd", "Specify video resolution")

	flag.Parse()

	if url == "" {
		fmt.Println("Invalid URL")
		os.Exit(1)
	}

	downloadVideo(url)
}

func getLink(query, url string) ([]byte, error) {

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Cache-Control", "no-cache")
	req.URL.RawQuery = query

	transport := &http.Transport{
		TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}

	client := &http.Client{Transport: transport}
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	return body, nil
}

func downloadVideo(videoURL string) {
	videoIDRegEx := regexp.MustCompile("[^videos/]*$")
	videoID := string(videoIDRegEx.Find([]byte(videoURL)))

	fmt.Println(string(videoID))

	query := url.Values{}
	responseByte, err := getLink(query.Encode(), videoURL)
	if err != nil {
		fmt.Println("Something went wrong!")
	}

	regEx := regexp.MustCompile(`(?P<hdSrc>hd_src):"(?P<hdLink>.*?)",(?P<sdSrc>sd_src):"(?P<sdLink>.*?)"`)
	if strings.Contains(string(responseByte), "hd_src:null") {
		regEx = regexp.MustCompile(`(?P<hdSrc>hd_src):(?P<hdLink>.*?),(?P<sdSrc>sd_src):"(?P<sdLink>.*?)"`)
	}

	names := regEx.SubexpNames()
	result := regEx.FindStringSubmatch(string(responseByte))

	linkMap := map[string]string{}

	if len(result) == 0 {
		fmt.Println("Download link not found!")
		return
	}

	for i, n := range result {
		linkMap[names[i]] = n
	}

	videoLink := videoLink{
		FileName: videoID,
		HDLink:   linkMap["hdLink"],
		SDLink:   linkMap["sdLink"],
	}

	validateURL(videoLink)

}

func validateURL(videoLink videoLink) {

	if res == "hd" {
		if videoLink.HDLink == "null" {
			fmt.Println("Cannot find HD file, please try another resolution")
			return
		}
		saveFile(videoLink.HDLink, videoLink.FileName)
	}

	if res == "sd" {
		if videoLink.SDLink == "null" {
			fmt.Println("Cannot find SD file, please try another resolution")
			return
		}
		saveFile(videoLink.SDLink, videoLink.FileName)
	}
}

func saveFile(videoURL string, fileName string) {
	fmt.Println("Downloading...")

	videoFile, err := http.Get(videoURL)
	if err != nil {
		fmt.Println("Can't download the video.")
		return
	}
	defer videoFile.Body.Close()

	file, err := os.Create(fileName + ".mp4")
	if err != nil {
		fmt.Println("Cant't save file")
	}
	defer file.Close()

	fmt.Println("Saving...")
	io.Copy(file, videoFile.Body)
	fmt.Println("Done!")

}

type videoLink struct {
	FileName string
	HDLink   string
	SDLink   string
}
