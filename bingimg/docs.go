package bingimg

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"sync"
)

type BingImgWork struct {
	Cookies    []*http.Cookie
	Client     *http.Client
	AuthCookie string
	Name       string
}

// Image is a struct that holds the filename and data of an image
type BingImage struct {
	Filename string
	Data     string
}

const BING_URL = "https://www.bing.com"
const BASE_ENDPOINT = "https://www.bing.com/images/create"

var BadImagesLinks = []string{
	"https://r.bing.com/rp/in-2zU3AJUdkgFe7ZKv19yPBHVs.png",
	"https://r.bing.com/rp/TX9QuO3WzcCJz1uaaSwQAz39Kb0.jpg",
}

const Cookie_U = "1uAhsjMwUpA_R39yotapzfmdbVrgtSKlA3nbu4_K66p0htEIrblPaE-kqJt0mkxBueIjpSiu6X_Nw4SxFdr6F9bmHNP1xyvpvI1QHC3WBnp5VGOZhTrDbX_uaIa_ZdK2T0htJBoup2KErYZ7-0otR7qjGKWZ80l2C05PGhD2ABgrTF9oq73J9N-3w6R96WAeaYg4FFxECg4GKtiVsAs40ZQ"
const PromptTest2 = "Imagine a smart software that can analyze and edit photos with ease. It can detect faces, objects, colors, and emotions in your images and apply filters, effects, and enhancements accordingly. It can also help you create stunning collages, slideshows, and animations with your photos. Whether you want to beautify your selfies, improve your landscapes, or unleash your creativity, this software can do it all for you. Try it now and see the difference!"
const PromptTest1 = "A simple, yet elegant logo that features a speech bubble with a question mark inside. The speech bubble is made of different colors, representing the diversity of people who use the app."
const Auth_Kiev = "FABaBBRaTOJILtFsMkpLVWSG6AN6C/svRwNmAAAEgAAACOUkG6JL+vxXGARWw490H53rlXTmBK6XkX1mFRSLAstP7Z8783MFUGypRvUzFeTVriQZlOgDfkxFi/+NecnLYgUsaxE3FIVhYQaE7lBM4LZ9UOiz5Vhu8Lxfe77fQL46xpykFsK+pPQuKSs+kooO3u6mRn9/A2a56+tEV+XHqaGCGEZwwBtjVsg7EdMPuwOEr4yvVJmJsTiPxhyLl8e4Hf+UtA7qpgKf3bOFIFYDoq5uNTzj3cBTnlYez8t43UINzEgBCShzwlqiMEZhOomefNFgt2VlrQbmzU/KfAXtCelrz13wR0FlQzIGRNHkE7NiZFi2FGnBJEt7H1Q3cZ029nRJ414t1qIyT0GvCjs+8TU7Yf5cVShBdwNWhCVNPy2zwcPq+/c7ic12A+er1cVjt+Ju7608Cv9f6zg+F1jCoYD3jTzXXVqLSYLP1HmI3yBAk18oWovjvLb13o+bSsMEoUwyIriLnAgfPXEKrRwGr/lk+Wo5BHISWt1lmGDYxHshIwyIb1J84lMlRL/0UX7Si1CJ9/pYUwOcF0FCi0c1hnv77GZcsM3V4L4AuC/vwSeSh6Ww5gGDwi49yPQk8dukQ7hs2SMYxbkixfA3uUIXHHUlyT3QJCSSCaFGbJ+PvN69vLKbrBOBfpAeIYRDh8i8vC8kn10pMAhR8mIbO0pkoCNUMt0QZWjklO053thF6XQHeSbXm53E8ts3sDmC4OODtDOdnZ9RKEhypbK1bAr9prN7e/+qbjhWeZmb8tD0oORRvHdufVrA9pmOA5iUSw7qoxo/QDd5EgPPmENc6qV1A+v1sOsQdvVQ6SqCcwkhuWpZkN3FQRGsp6fy5BZ9CiZ2LCS4dtqAAtC9jhDkPghtmlt2epL+b93hxKqzI/qDl1Ac09zCqyGeZPsNn4m6Jp2s5Qf96YLUNZxidwfHD26U8jceO9YbHCz8VNifOREO4QPsYhLDMv8aBfDNUWoKgSDejAlMur9bVgdwkxhcFwJ7xtWkYOUp2mwDxRrQpHDWUmInpqI9ybwWxazUz5YtKwpIn5DvSxqpdP8UEcv2zCsSodxnnGxQgl3F60wzOGm8tVrCWxpnxAcym3fR0arhOFFX0d2m5+hN+pIHfK4jx9FBIgfHEGw1odGZ+im+O+2iLVl3SSQgJMnqbNTq8+gHYvi66rVYn4tlGmYZrldz2pDivxSTaKJHpxbLp8Hhtci66n9Ob1aFaZzhvC3q7CXM3utHlXp21QOeGQj6OXoMDpFUogFTj/NmbsxI8YvRzMsSP634DhmhtTrawnqiTTwfcd18a5tpyd87cCMcCoOqqnEBqshxQMOGtyvxXxJjpk+dbs1UA63NhQExAL3eNM6gHiZIUMHVXYHCax/I4KX6jXfdrWaiZzlQCb0U6rZdlO4TKJQwbaUxFAD94wTnuaHhtGeAjU5ZdqwWQeD7gQ=="

const sc = "\u001B7" // save cursor
const rc = "\u001B8" // restore cursor

var (

	// Error messages.
	errorTimeout         = "Your request has timed out."
	errorRedirect        = "Redirect failed"
	errorBlockedPrompt   = "Your prompt has been blocked by Bing. Try to change any bad words and try again."
	errorNoResults       = "Could not get results"
	errorUnsupportedLang = "\nthis language is currently not supported by bing"
	errorBadImages       = "Bad images"
	errorNoImages        = "No images"

	// Sending message.
	sendingMessage = "Sending request..."

	// Waiting message.
	waitMessage = "Waiting for results..."

	// Downloading message.
	downloadMessage  = "\nDownloading images..."
	headers          = map[string]string{}
	forwardedIP      = fmt.Sprintf("13.%d.%d.%d", rand.Intn(3)+101, rand.Intn(255), rand.Intn(255))
	bingURL          = ""
	thumbnail_width  = 512
	thumbnail_height = 512
)

// Config is a struct that holds the configuration options
type Config struct {
	BingImageCookie string
	TempDir         string
}

type BingImageClient struct {
	Headers map[string]string
	Data    url.Values
	Cookies []*http.Cookie
	Options IOptions
	Client  *http.Client
}

type ImageGen struct {
	session   *http.Client
	headers   map[string]string
	cookies   []*http.Cookie
	quiet     bool
	debugFile string
	Thumbnail bool
}

type IOptions struct {
	Dir    string
	Notify bool
	Token  string
}

// config is a pointer to Config that will be initialized only once
var config *Config

// once is a sync.Once object that ensures the initialization happens only once
var once sync.Once

// initConfig is a function that initializes the config pointer with values from environment variables or defaults
func initConfig() {
	once.Do(func() {
		config = &Config{
			BingImageCookie: os.Getenv("BING_IMAGE_COOKIE"),
			TempDir:         os.Getenv("TEMP_DIR"),
		}
		// if tempDir is empty, use "/tmp" as default
		if config.TempDir == "" {
			config.TempDir = "/tmp"
		}
	})
}
