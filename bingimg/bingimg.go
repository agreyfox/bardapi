package bingimg

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"image/jpeg"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/nfnt/resize"
)

func init() {
	// Get the Bing URL from the environment variable, or use the default value.
	bingURL = os.Getenv("BING_URL")
	if bingURL == "" {
		bingURL = "https://www.bing.com"
	}

	// Generate a random IP address in the range 13.104.0.0/14.
	forwardedIP = fmt.Sprintf("13.%d.%d.%d", rand.Intn(3)+101, rand.Intn(255), rand.Intn(255))

	// Create a map of HTTP headers.
	headers = map[string]string{
		"accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
		"accept-language": "en-US,en;q=0.9",
		"cache-control":   "max-age=0",
		"content-type":    "application/x-www-form-urlencoded",
		"referrer":        "https://www.bing.com/images/create/",
		"origin":          "https://www.bing.com",
		"user-agent":      "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36 Edg/110.0.1587.63",
		"x-forwarded-for": forwardedIP,
	}

	// Do something...
}

// Debug function.
func debug(debugFile string, textVar interface{}) {
	
	// helper function for debug
	f, err := os.OpenFile("bingimg.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	_, err = f.WriteString(fmt.Sprintf("%v", textVar))
	if err != nil {
		panic(err)
	}
}

func NewBingImageClient(options IOptions) *BingImageClient {

	pwd, _ := os.Getwd()
	options = IOptions{
		Dir:    pwd,
		Notify: false,
		Token:  options.Token,
	}
	v := url.Values{}
	// 遍历 map 中的键值对，添加到 url.Values 中
	for key, value := range headers {
		v.Add(key, value)
	}
	client := http.Client{}

	return &BingImageClient{
		Headers: headers,
		Data:    v,
		Options: options,
		Client:  &client,
	}
}

func NewImageGen(authCookie string, quiet bool, debugFile string) (*ImageGen, error) {
	bingURL := os.Getenv("BING_URL")
	if bingURL == "" {
		bingURL = "https://www.bing.com"
	}

	// Generate a random IP address in the range 13.104.0.0/14.
	forwardedIP = fmt.Sprintf("13.%d.%d.%d", rand.Intn(3)+101, rand.Intn(255), rand.Intn(255))

	// Create a map of HTTP headers.
	headers = map[string]string{
		"accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
		"accept-language": "en-US,en;q=0.9",
		"cache-control":   "max-age=0",
		"content-type":    "application/x-www-form-urlencoded",
		"referrer":        "https://www.bing.com/images/create/",
		"origin":          "https://www.bing.com",
		"user-agent":      "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36 Edg/110.0.1587.63",
		"x-forwarded-for": forwardedIP,
	}
	jar, err := cookiejar.New(nil)
	if err != nil {
		// handle error
		fmt.Println(err)
		return nil, err
	}
	session := &http.Client{
		Jar:     jar,
		Timeout: time.Duration(200 * time.Second),
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if req.URL.Host == bingURL {
				return nil
			}

			fmt.Println("redirect...", via[len(via)-1].URL)
			return http.ErrUseLastResponse

		},
	}

	//session.Headers = headers
	authc := http.Cookie{
		Name:  "_U",
		Value: authCookie,
	}

	return &ImageGen{
		session:   session,
		headers:   headers,
		cookies:   []*http.Cookie{&authc},
		quiet:     quiet,
		debugFile: debugFile,
	}, nil
}

func (c *ImageGen) MakeThumbnail(sign bool) {
	c.Thumbnail = sign
}

func (c *ImageGen) UpdateHeader() {
	// Generate a random IP address in the range 13.104.0.0/14.
	forwardedIP = fmt.Sprintf("13.%d.%d.%d", rand.Intn(3)+101, rand.Intn(255), rand.Intn(255))

	// Create a map of HTTP headers.
	headers = map[string]string{
		"accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
		"accept-language": "en-US,en;q=0.9",
		"cache-control":   "max-age=0",
		"content-type":    "application/x-www-form-urlencoded",
		"referrer":        "https://www.bing.com/images/create/",
		"origin":          "https://www.bing.com",
		"user-agent":      "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36 Edg/110.0.1587.63",
		"x-forwarded-for": forwardedIP,
	}
	c.headers = headers
}

func (self *ImageGen) GetImages(prompt string) ([]string, error) {
	// Fetches image links from Bing
	// Parameters:
	// prompt: string
	if !self.quiet {
		fmt.Println(sendingMessage)
	}
	if self.debugFile != "" {
		debug(self.debugFile, sendingMessage)
	}
	urlEncodedPrompt := url.QueryEscape(prompt)
	payload := fmt.Sprintf("q=%s&qs=ds", urlEncodedPrompt)
	// https://www.bing.com/images/create?q=<PROMPT>&rt=3&FORM=GENCRE
	url := fmt.Sprintf("%s/images/create?q=%s&rt=4&FORM=GENCRE", bingURL, urlEncodedPrompt)
	req, _ := http.NewRequest("Post", url, strings.NewReader(payload))
	for k, v := range self.headers {
		req.Header.Add(k, v)
	}

	self.session.Jar.SetCookies(req.URL, self.cookies)
	//response, err := self.session.Post(url, "", strings.NewReader(payload))
	response, err := self.session.Do(req)

	if err != nil {
		fmt.Println(err)
		return []string{}, err
	}
	response.Close = true
	defer response.Body.Close()
	// check for content waring message
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
		return []string{}, err
	}

	bodyStr := string(body)

	if strings.Contains(strings.ToLower(bodyStr), "this prompt has been blocked") {
		if self.debugFile != "" {
			debug(self.debugFile, fmt.Sprintf("ERROR: %s", errorBlockedPrompt))
		}
		fmt.Println(errorBlockedPrompt)
		return []string{}, errors.New(errorBlockedPrompt)
	}
	if strings.Contains(strings.ToLower(bodyStr), "we're working hard to offer image creator in more languages") {
		if self.debugFile != "" {
			debug(self.debugFile, fmt.Sprintf("ERROR: %s", errorUnsupportedLang))
		}
		fmt.Println(errorUnsupportedLang)
		return []string{}, errors.New(errorUnsupportedLang)
	}
	if response.StatusCode != 302 {
		// if rt4 fails, try rt3
		url = fmt.Sprintf("%s/images/create?q=%s&rt=3&FORM=GENCRE", BING_URL, urlEncodedPrompt)
		response3, err := self.session.Post(url, "", nil)
		if err != nil {
			fmt.Println(err)
			return []string{}, err
		}
		response3.Close = true
		defer response3.Body.Close()
		if response3.StatusCode != 302 {
			if self.debugFile == "" {
				debug(self.debugFile, fmt.Sprintf("ERROR: %s", errorRedirect))
			}
			fmt.Printf("ERROR: %s\n", bodyStr)
			fmt.Println(errorRedirect)
			return []string{}, errors.New(errorRedirect)
		}
		response = response3
	}
	// Get redirect URL
	redirectURL := strings.Replace(response.Header.Get("Location"), "&nfy=1", "", -1)
	requestID := strings.Split(redirectURL, "id=")[1]
	self.session.Get(fmt.Sprintf("%s%s", BING_URL, redirectURL))
	// https://www.bing.com/images/create/async/results/{ID}?q={PROMPT}
	pollingURL := fmt.Sprintf("%s/images/create/async/results/%s?q=%s", BING_URL, requestID, urlEncodedPrompt)
	// Poll for results
	if self.debugFile == "" {
		debug(self.debugFile, "Polling and waiting for result")
	}
	if !self.quiet {
		fmt.Println("Waiting for results...")
	}
	startWait := time.Now()
	fmt.Print(sc)
	resultstr := ""
	for {
		if int(time.Since(startWait).Seconds()) > 200 {
			if self.debugFile == "" {
				debug(self.debugFile, fmt.Sprintf("ERROR: %s", errorTimeout))
			}
			fmt.Println(errorTimeout)
			return []string{}, errors.New(errorTimeout)
		}
		if !self.quiet {
			fmt.Print(rc + sc)
			fmt.Print(".")

		}
		response, err := self.session.Get(pollingURL)
		if err != nil {
			fmt.Println(err)
			return []string{}, err
		}
		response.Close = true
		defer response.Body.Close()
		if response.StatusCode != 200 {
			if self.debugFile == "" {
				debug(self.debugFile, fmt.Sprintf("ERROR: %s", errorNoResults))
			}
			fmt.Println(errorNoResults)
			return []string{}, errors.New(errorNoResults)
		}
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			fmt.Println(err)
			return []string{}, err
		}
		bodyStr := string(body)
		resultstr = fmt.Sprintf("%s", bodyStr)
		fmt.Println(resultstr, response.Status)
		if bodyStr == "" || strings.Contains(bodyStr, "errorMessage") {
			time.Sleep(1 * time.Second)
			continue
		} else {
			break
		}
	}
	// Use regex to search for src="" imageLinks
	re := regexp.MustCompile(`src="([^"]+)"`)
	imageLinks := re.FindAllStringSubmatch(resultstr, -1)

	// Remove size limit normalImageLinks
	normalImageLinks := make([]string, 0, len(imageLinks))
	for _, link := range imageLinks {
		normalImageLinks = append(normalImageLinks, strings.Split(link[1], "?w=")[0])
	}

	// Remove duplicates normalImageLinks
	normalImageLinks = removeDuplicates(normalImageLinks)

	// Bad images badImages
	badImages := []string{
		"https://r.bing.com/rp/in-2zU3AJUdkgFe7ZKv19yPBHVs.png",
		"https://r.bing.com/rp/TX9QuO3WzcCJz1uaaSwQAz39Kb0.jpg",
	}
	for _, img := range normalImageLinks {
		if contains(badImages, img) {
			fmt.Println(errorBadImages)
			return []string{}, errors.New(errorBadImages)
		}
	}

	// No images
	if len(normalImageLinks) == 0 {
		panic(errorNoImages)
	}

	return normalImageLinks, nil
}

func removeDuplicates(slice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func (c *ImageGen) SaveImages(links []string, outputDir string, fileName string) ([]string,error) {
	ctx := context.Background()
	ret:=[]string{}
	// Saves images to output directory
	if c.debugFile == "" {
		debug(c.debugFile, downloadMessage)
	}
	if !c.quiet {
		fmt.Println(downloadMessage)
	}
	// Create the output directory if it does not exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return ret,err
	}
	// Use a prefix for the file name if provided
	prefix := ""
	if fileName != "" {
		prefix = fileName + "_"
	}
	jpegIndex := 0
	var wg sync.WaitGroup

	for _, link := range links {
		// Find a unique file name for the image
		afilePath := ""
		for {
			afilePath = filepath.Join(outputDir, fmt.Sprintf("%s%d.jpeg", prefix, jpegIndex))
			if _, err := os.Stat(afilePath); os.IsNotExist(err) {
				break
			}
			jpegIndex++
		}
		// Make a GET request to the link
		req, err := http.NewRequestWithContext(ctx, "GET", link, nil)
		if err != nil {
			return ret,err
		}
		resp, err := c.session.Do(req)
		if err != nil {
			return ret,err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return ret,fmt.Errorf("invalid status code: %d", resp.StatusCode)
		}
		// Create and open the output file
		outputFile, err := os.Create(afilePath)
		if err != nil {
			return ret,err
		}
		defer outputFile.Close()
		// Copy the response body to the output file
		if _, err := io.Copy(outputFile, resp.Body); err != nil {
			return ret,err
		}

		// Check for inappropriate contents in the image
		if c.isBadImage(afilePath) {
			return ret,fmt.Errorf("inappropriate contents found in the generated images. Please try again or try another prompt")
		} else {
			if c.Thumbnail {
				wg.Add(1)
				go func(pre string, jpi int, wg *sync.WaitGroup) {
					defer func() {
						wg.Done()
					}()
					afilePath :=filepath.Join(outputDir, fmt.Sprintf("%s%d.jpeg", pre, jpi))
					thumbfilePath := filepath.Join(outputDir, fmt.Sprintf("%s%d.thumb.jpeg", pre, jpi))
					imgfile, err := os.Open(afilePath)
					if err != nil {
						fmt.Println("error:", err)
						return
					}
					// 解码图像文件
					img, err := jpeg.Decode(imgfile)
					if err != nil {
						fmt.Println("error:", err)
						return
					}
					// 计算缩略图大小
					width := thumbnail_width // 缩略图宽度
					ratio := float64(width) / float64(img.Bounds().Dx())
					height := uint(float64(img.Bounds().Dy()) * ratio)

					// 调整图像大小
					resizedImg := resize.Resize(uint(width), height, img, resize.Bilinear)

					// 保存调整大小后的图像文件
					out, err := os.Create(thumbfilePath)
					if err != nil {
						fmt.Println("error:", err)
						return
					}
					jpeg.Encode(out, resizedImg, nil)
					out.Close()
					imgfile.Close()
					fmt.Println("Create Thumbnail:", thumbfilePath)
					targetfilename := filepath.Base(thumbfilePath)
					ret = append(ret, targetfilename)
				}(prefix, jpegIndex, &wg)

			}else{
				targetfilename := filepath.Base(afilePath)
				ret=append(ret, targetfilename)
				fmt.Println("use origin file as output :",targetfilename)
			}
		}

	}
	wg.Wait()
	return ret,nil
}

func (self *ImageGen) isBadImage(filePath string) bool {
	// TODO: implement a function to check if the image contains inappropriate contents
	// This could be done by using an external API or library that can detect such contents
	// For example: https://cloud.google.com/vision/docs/detecting-safe-search
	// For simplicity, we assume that this function returns false for now
	return false
}
