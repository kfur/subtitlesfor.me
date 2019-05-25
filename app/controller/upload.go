package controller

import (
	"encoding/json"
	"fmt"
	"github.com/kfur/subtitler/app"
	"github.com/kfur/subtitler/app/srt"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)




func UploadAudio(w http.ResponseWriter, r *http.Request) {

	waitConverted := sync.WaitGroup{}

	token := app.TokenGenerator()

	err := r.ParseForm()
	if err != nil {
		fmt.Println(err)
	}

	// Retrieve file content due to mimetype is multipart/form-data
	file, fileHeader, err := r.FormFile("audio")
	if err != nil {
		fmt.Println(err)
	}

	_, params, err := mime.ParseMediaType(fileHeader.Header.Get("Content-Disposition"))
	if err != nil {
		fmt.Println(err)
	}
	filename := params["filename"]
	ext := filepath.Ext(filename)

	subtitles := app.Recogniser.UploadToRecogniseWithAudioSplitting(file, ext, token, &waitConverted)

	go func() {

		waitConverted.Wait()
		sort.Sort(subtitles)

		subtilesStrs := srt.MakeSRTFromJSONChunks(*subtitles)

		app.TokenMap[token].Subtitles <-
			app.Subtitles{
			strings.Join(subtilesStrs, "\n"),
			strings.TrimRight(filename, ext) + ".srt",
			nil,
		}
	}()

	// Return token to client
	var responseWriter io.Writer = w
	respData, err := json.Marshal(struct {
		Token string `json:"token"`
	}{token})
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	_, err = responseWriter.Write(respData)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	return
}
