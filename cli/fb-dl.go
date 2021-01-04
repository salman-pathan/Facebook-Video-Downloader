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

	DownloadVideo(url)
}

func GetLink(query, url string) ([]byte, error) {

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

func DownloadVideo(videoUrl string) {
	videoIdRegEx := regexp.MustCompile("[^videos/]*$")
	videoId := string(videoIdRegEx.Find([]byte(videoUrl)))

	fmt.Println(string(videoId))

	query := url.Values{}
	responseByte, err := GetLink(query.Encode(), videoUrl)
	if err != nil {
		fmt.Println("Something went wrong!")
	}

	regEx := regexp.MustCompile(`(?P<hdSrc>hd_src):"(?P<hdLink>.*?)",(?P<sdSrc>sd_src):"(?P<sdLink>.*?)"`)
	if strings.Contains(string(responseByte), "hd_src:null") {
		regEx = regexp.MustCompile(`(?P<hdSrc>hd_src):(?P<hdLink>.*?),(?P<sdSrc>sd_src):"(?P<sdLink>.*?)"`)
	}

	// fmt.Println(string(responseByte))

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

	videoLink := VideoLink{
		FileName: videoId,
		HDLink:   linkMap["hdLink"],
		SDLink:   linkMap["sdLink"],
	}

	ValidateURL(videoLink)

}

func ValidateURL(videoLink VideoLink) {

	if res == "hd" {
		if videoLink.HDLink == "null" {
			fmt.Println("Cannot find HD file, please try another resolution")
			return
		}
		SaveFile(videoLink.HDLink, videoLink.FileName)
	}

	if res == "sd" {
		if videoLink.SDLink == "null" {
			fmt.Println("Cannot find SD file, please try another resolution")
			return
		}
		SaveFile(videoLink.SDLink, videoLink.FileName)
	}
}

func SaveFile(videoUrl string, fileName string) {
	fmt.Println("Downloading...")

	videoFile, err := http.Get(videoUrl)
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

type VideoLink struct {
	FileName string
	HDLink   string
	SDLink   string
}
