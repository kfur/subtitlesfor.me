package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/3d0c/gmf"
	"github.com/kfur/subtitler/app/audio"
	c "github.com/kfur/subtitler/config"
	"github.com/watson-developer-cloud/go-sdk/speechtotextv1"
	"io"
	"log"
	"sync"
)

type Subtitles struct {
	SrtText         string
	Filename        string
	ConvertingError error
}

type SubtitlesChunk struct {
	Chunk    string
	Position uint
	Duration float64
}

type SubtitlesRecognitionResultArray []SubtitlesChunk

func (s *SubtitlesRecognitionResultArray) Len() int {
	return len(*s)
}

func (s *SubtitlesRecognitionResultArray) Less(i, j int) bool {
	return (*s)[i].Position < (*s)[j].Position
}

func (s *SubtitlesRecognitionResultArray) Swap(i, j int) {
	(*s)[i], (*s)[j] = (*s)[j], (*s)[i]
}

type AudioFileReader struct {
	reader io.ReadCloser
}

func (afl *AudioFileReader) customReader() ([]byte, int) {
	b := make([]byte, gmf.IO_BUFFER_SIZE)

	n, err := afl.reader.Read(b)
	if err != nil && err == io.EOF {
		fmt.Println("closing body: ", err)
		afl.reader.Close()
		return b, n
	}
	if err != nil {
		fmt.Println(err)
		return nil, n
	}

	return b, n
}

type ReadCloserWithCallback struct {
	readCloser      io.Reader
	onCloseCallback func(readCloser *ReadCloserWithCallback)
}

func (r *ReadCloserWithCallback) Read(p []byte) (n int, err error) {
	return r.readCloser.Read(p)
}

func (r *ReadCloserWithCallback) Close() error {
	r.onCloseCallback(r)
	return nil
}

type RecogniseService struct {
	service *speechtotextv1.SpeechToTextV1
}

var Recogniser *RecogniseService

func InitRecogniseService() {
	// Instantiate the Watson Speech To Text service
	var serviceErr error
	var Recogniser = RecogniseService{}
	Recogniser.service, serviceErr = speechtotextv1.
		NewSpeechToTextV1(&speechtotextv1.SpeechToTextV1Options{
			URL:       c.Config.STTOptions.URL,
			IAMApiKey: c.Config.STTOptions.IAMApiKey,
		})

	// Check successful instantiation
	if serviceErr != nil {
		panic(serviceErr)
	}
}

func (r* RecogniseService) Recognise(source io.ReadCloser, contentType string, interrupt *context.CancelFunc) *speechtotextv1.SpeechRecognitionResults {

	/* RECOGNIZE */

	fmt.Println("url content-type is: " + contentType)

	// Create a new RecognizeOptions for ContentType "audio/mp3"
	recognizeOptions := r.service.
		NewRecognizeOptions(source, contentType)

	recognizeOptions.SetTimestamps(true)
	recognizeOptions.SetProfanityFilter(false)

	// Call the speechToText Recognize method
	ctx, cancel := context.WithCancel(context.Background())
	if interrupt != nil && *interrupt == nil {
		*interrupt = cancel
	}
	response, responseErr := r.service.Recognize(recognizeOptions, ctx)

	// Check successful call
	if responseErr != nil {
		fmt.Println(responseErr)
		return nil
	}

	// Cast recognize.Result to the specific dataType returned by Recognize
	// NOTE: most methods have a corresponding Get<methodName>Result() function
	recognizeResult := r.service.GetRecognizeResult(response)

	//// Check successful casting
	//if recognizeResult != nil {
	//	recResultText, err := json.Marshal(recognizeResult)
	//	if err != nil {
	//		fmt.Println(err)
	//		return nil
	//	}
	//	return MakeSRTFromJSON(string(recResultText))
	//}

	return recognizeResult
}

func (r* RecogniseService) UploadToRecogniseWithAudioSplitting(file io.ReadCloser, fileExt string, token string, waitConverted *sync.WaitGroup) *SubtitlesRecognitionResultArray {
	waitUntilUploaded := sync.WaitGroup{}
	subtitlesLock := sync.Mutex{}
	subtitles := SubtitlesRecognitionResultArray{}

	ictx := gmf.NewCtx()
	defer ictx.Close()

	inFormat := audio.ConvertExt2Format(fileExt)

	if err := ictx.SetInputFormat(inFormat); err != nil {
		log.Fatal(err)
	}

	audioReader := AudioFileReader{file}

	avioCtx, err := gmf.NewAVIOContext(ictx, &gmf.AVIOHandlers{ReadPacket: audioReader.customReader})
	defer gmf.Release(avioCtx)
	if err != nil {
		log.Fatal(err)
	}

	err = ictx.SetPb(avioCtx).OpenInput("")
	if err != nil {
		fmt.Println(err)
		return nil
	}

	as := audio.AudioSplitter{InputCtx:ictx}

	tokenData := TokeData{make(chan Subtitles), new(context.CancelFunc)}
	TokenMap[token] = tokenData

	callback := func(readCloser *ReadCloserWithCallback) {
		waitUntilUploaded.Done()
	}

	var position uint = 0
	for a := range as.GetAudioChunks() {
		fmt.Println("Audio FIle ")
		waitUntilUploaded.Add(1)
		waitConverted.Add(1)

		chunkAudioReader := bytes.NewReader(a.Data.Bytes())
		// Converting
		go func(position uint, duration float64) {
			recognizeResult := r.Recognise(&ReadCloserWithCallback{
				chunkAudioReader,
				callback},
				"audio/" + a.CodecName,
				tokenData.Cancel)

			if recognizeResult == nil {
				TokenMap[token].Subtitles <- Subtitles{"", "", errors.New("Error")}
				return
			}

			recResultData, err := json.Marshal(recognizeResult)
			if err != nil {
				fmt.Println(err)
				TokenMap[token].Subtitles <- Subtitles{"", "", errors.New("Error")}
				return
			}

			//fmt.Println(strings.Join(srtStrs, "\n"))
			subtitlesLock.Lock()
			subtitles = append(subtitles, SubtitlesChunk{string(recResultData), position, duration})
			subtitlesLock.Unlock()

			waitConverted.Done()
			//TokenMap[token].Subtitles <- strings.Join(srtStrs, "\n")
		}(position, a.Duration)
		position++
	}

	// Wait for request body closing by http client
	// Because body will be closed if we pass without waiting
	fmt.Println("Wait for body closing")
	waitUntilUploaded.Wait()

	return &subtitles
}
