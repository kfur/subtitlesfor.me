package controller

import (
	"fmt"
	"github.com/kfur/subtitler/app"
	"github.com/kfur/subtitler/app/shared/view"
	"github.com/kfur/subtitler/app/srt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"
)


// IndexGET displays the home page
func IndexGET(w http.ResponseWriter, r *http.Request) {
	// Display the view
	v := view.New(r)
	v.Name = "index/anon"
	v.Render(w)
	return
}


// NotepadCreatePOST handles the note creation form submission
func IndexPOST(w http.ResponseWriter, r *http.Request) {

	// Get form values
	var token string
	audioUrl := strings.TrimSpace(r.FormValue("url"))
	if audioUrl != "" {
		fmt.Println("Converting url: " + audioUrl)

		if !isValidUrl(audioUrl) {
			return
		}

		token = app.TokenGenerator()

		//// Open file with mp3 to recognize
		//audio, audioErr := http.Get(audioUrl)
		//if audioErr != nil {
		//	fmt.Println(audioErr)
		//	return
		//}

		waitConverted := sync.WaitGroup{}
		var subtitles *app.SubtitlesRecognitionResultArray

		go func() {
			//defer audio.Body.Close()
			//mimeType := audio.Header.Get("Content-Type")
			//splitedMimeType := strings.Split(mimeType, "/")
			//if len(splitedMimeType) != 2 {
			//	//error
			//}


			//subtitles = app.Recogniser.UploadToRecogniseWithAudioSplitting(audio.Body, splitedMimeType[1], token, &waitConverted)
			subtitles = app.Recogniser.LoadFromURL(audioUrl, token, &waitConverted)
		}()

		go func() {
			time.Sleep(time.Second * 10)
			waitConverted.Wait()
			// TODO fix "subtitles suddenly is nil"
			sort.Sort(subtitles)

			subtilesStrs := srt.MakeSRTFromJSONChunks(*subtitles)

			app.TokenMap[token].Subtitles <- app.Subtitles{strings.Join(subtilesStrs, "\n"), "", nil}
		}()
	} else {
		token = r.FormValue("token")
		fmt.Println("Received token: " + token)
	}


	// Display the view
	v := view.New(r)
	v.Name = "index/convert"
	v.Vars["url"] = audioUrl
	v.Vars["websocketRequestToken"] = token
	v.Render(w)
}

func isValidUrl(toTest string) bool {
	_, err := url.ParseRequestURI(toTest)
	if err != nil {
		return false
	} else {
		return true
	}
}