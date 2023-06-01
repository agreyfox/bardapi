package bingimg

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// NewBingImage is a function that creates and returns a new BingImage with the given parameters
func NewBingImage(cookie string) *BingImgWork {
	work := createSession(cookie)
	work.AuthCookie = cookie
	return work
}

func createSession(authCookie string) *BingImgWork {
	// create a new http client
	client := &http.Client{}
	// create a default request
	req, err := http.NewRequest("GET", "https://www.bing.com/images/create/", nil)
	if err != nil {
		return nil
	}
	// set the headers for the request
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9,zh-CN;q=0.8,zh-TW;q=0.7,zh;q=0.6")
	req.Header.Set("Cache-Control", "max-age=0")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referrer-Policy", "origin-when-cross-origin")
	req.Header.Set("Referrer", "https://www.bing.com/images/create/")
	req.Header.Set("Origin", "https://www.bing.com")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/111.0.0.0 Safari/537.36 Edg/111.0.1661.54")
	req.Header.Set("Cookie", `_U=${authCookie}`)
	req.Header.Set("Sec-Ch-UA", `"Microsoft Edge";v="111", "Not(A:Brand";v="8", "Chromium";v="111"`)
	req.Header.Set("Sec-Ch-UA-Mobile", "?0")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	// set the default request for the client
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		// copy the headers from the default request to the redirected request
		for key, values := range via[0].Header {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}
		return nil
	}
	//return client

	return &BingImgWork{
		Client: client,
	}
}

func (c *BingImgWork) getImages(prompt string) ([]string, error) {
	fmt.Println("Sending request...")
	urlEncodedPrompt := url.QueryEscape(prompt)

	url := fmt.Sprintf("%s/images/create?q=%s&rt=3&FORM=GENCRE", BING_URL, urlEncodedPrompt) // force use rt=3
	fmt.Println(url)
	// create a post request
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, err
	}
	// send the request and get the response
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var redirectURL string
	// check the status code
	if resp.StatusCode == 200 {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		fmt.Println(string(bodyBytes))
		redirectURL = resp.Request.URL.String()
		redirectURL = strings.Replace(redirectURL, "&nfy=1", "", 1)
	} else if resp.StatusCode != 302 {
		fmt.Errorf("ERROR: the status is %d instead of 302 or 200", resp.StatusCode)
		return nil, errors.New("Redirect failed")
	}

	fmt.Println("Redirected to", redirectURL)

	// create a get request for the redirect URL
	req, err = http.NewRequest("GET", redirectURL, nil)
	if err != nil {
		return nil, err
	}
	// send the request and get the response
	resp, err = c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	requestID := strings.Split(redirectURL, "id=")[1]

	pollingURL := fmt.Sprintf("%s/images/create/async/results/%s?q=%s", BING_URL, requestID, urlEncodedPrompt)

	fmt.Println("Waiting for results...")
	startWait := time.Now()
	var imagesResponse *http.Response

	for {
		if time.Since(startWait) > 5*time.Minute {
			return nil, errors.New("Timeout error")
		}
		fmt.Print(".")
		// create a get request for the polling URL
		req, err = http.NewRequest("GET", pollingURL, nil)
		if err != nil {
			return nil, err
		}
		// send the request and get the response
		imagesResponse, err = c.Client.Do(req)
		if err != nil {
			return nil, err
		}
		defer imagesResponse.Body.Close()
		if imagesResponse.StatusCode != 200 {
			return nil, errors.New("Could not get results")
		}
		bodyBytes, err := ioutil.ReadAll(imagesResponse.Body)
		if err != nil {
			return nil, err
		}
		bodyString := string(bodyBytes)
		if bodyString == "" {
			time.Sleep(time.Second)
			continue
		} else {
			break
		}
	}
	bodyBytes, err := ioutil.ReadAll(imagesResponse.Body)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &data); err != nil {
		return nil, err
	}

	if data["errorMessage"] == "Pending" {
		return nil, errors.New(
			"This prompt has been blocked by Bing. Bing's system flagged this prompt because it may conflict with their content policy. More policy violations may lead to automatic suspension of your access.",
		)
	} else if data["errorMessage"] != "" {
		return nil, errors.New(
			fmt.Sprintf("Bing returned an error: %s", data["errorMessage"]),
		)
	}

	imageLinks := regexp.MustCompile(`src="([^"]+)"`).FindAllStringSubmatch(data["data"].(string), -1)
	var normalImageLinks []string
	for _, link := range imageLinks {
		normalImageLinks = append(normalImageLinks, strings.Split(link[1], "?w=")[0])
	}

	normalImageLinks = removeDuplicates(normalImageLinks)

	badImages := BadImagesLinks

	for _, im := range normalImageLinks {
		if contains(badImages, im) {
			return nil, errors.New("Bad images")
		}
	}

	if len(normalImageLinks) == 0 {
		return nil, errors.New("No images")
	}

	return normalImageLinks, nil
}

// removeDuplicates is a helper function that removes duplicate elements from a slice of strings
func removeDuplicates_old(slice []string) []string {
	set := make(map[string]bool)
	var result []string
	for _, s := range slice {
		if !set[s] {
			set[s] = true
			result = append(result, s)
		}
	}
	return result
}

// contains is a helper function that checks if a slice of strings contains a given string
func contains_old(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

// save is a function that downloads images from links and saves them to outputDir
func (c *BingImgWork) saveImages(links []string, outputDir string) error {
	fmt.Println("\nDownloading images...")
	// create outputDir if it does not exist
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		return err
	}
	imageNum := 0
	for _, link := range links {
		// create a get request for the image link
		req, err := http.NewRequest("GET", link, nil)
		if err != nil {
			return err
		}
		// send the request and get the response
		resp, err := c.Client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		// create a file to write the image data
		outputPath := filepath.Join(outputDir, fmt.Sprintf("%d.jpeg", imageNum))
		writer, err := os.Create(outputPath)
		if err != nil {
			return err
		}
		defer writer.Close()
		// copy the image data to the file
		if _, err := io.Copy(writer, resp.Body); err != nil {
			return err
		}
		imageNum++
	}
	return nil
}

// generateImageFiles is a function that generates images from a prompt and returns them as base64 encoded strings
func (c *BingImgWork) generateImageFiles(prompt string) ([]BingImage, error) {

	authCookie := config.BingImageCookie

	outputDir := filepath.Join(config.TempDir, generateMD5(prompt))
	if authCookie == "" || prompt == "" {
		return nil, errors.New("需要image的cookies及相关设置")
	}
	// Create image generator session
	//session := c.createSession(authCookie)
	imageLinks, err := c.getImages(prompt)
	if err != nil {
		return nil, err
	}
	if err := c.saveImages(imageLinks, outputDir); err != nil {
		return nil, err
	}
	// Read saved images from the output directory
	imageFiles, err := os.ReadDir(outputDir)
	if err != nil {
		return nil, err
	}
	var images []BingImage
	for _, file := range imageFiles {
		filePath := filepath.Join(outputDir, file.Name())
		fileData, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}
		images = append(images, BingImage{
			Filename: file.Name(),
			Data:     base64.StdEncoding.EncodeToString(fileData),
		})
	}
	return images, nil
}

// generateMD5 is a function that takes a string and returns its MD5 hash
func generateMD5(s string) string {
	// convert the string to a byte slice
	data := []byte(s)
	// create a new md5 hash object
	hash := md5.New()
	// write the data to the hash object
	hash.Write(data)
	// get the hash sum as a byte slice
	sum := hash.Sum(nil)
	// encode the byte slice to a hexadecimal string
	hexString := hex.EncodeToString(sum)
	// return the hexadecimal string
	return hexString
}
