package srt

import (
	"encoding/json"
	"fmt"
	"github.com/kfur/subtitler/app"
	"math"
	"strconv"
)

func zipTextNTime(text []string, time []BeginEndTime) []WordStartEndTime {
	var result []WordStartEndTime
	for i := 0; i < len(text); i++ {
		result = append(result, WordStartEndTime{text[i], BeginEndTime{time[i].Begin, time[i].End}})
	}
	return result
}

type TimestampValue string

func (tv *TimestampValue) UnmarshalJSON(b []byte) error {
	if b[0] != '"' {
		*tv = TimestampValue(b)
		return nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	*tv = TimestampValue(s)
	return nil
}

type BeginEndTime struct {
	Begin string
	End   string
}

type WordStartEndTime struct {
	Word string
	Time BeginEndTime
}

func (wt *WordStartEndTime) UnmarshalJSON(b []byte) error {
	var ts [3]TimestampValue
	if err := json.Unmarshal(b, &ts); err != nil {
		return err
	}

	wt.Word = string(ts[0])
	wt.Time.Begin = string(ts[1])
	wt.Time.End = string(ts[2])

	return nil
}

type Alternative struct {
	WordsStartEndTimes []WordStartEndTime `json:"timestamps"`
	Confidence         float64            `json:"confidence"`
	Transcript         string             `json:"transcript"`
}

type Result struct {
	Alternatives []Alternative `json:"alternatives"`
	Final        bool          `json:"final"`
}

type ResultJSON struct {
	Results []Result `json:"results"`
}

func divmod(numerator float64, denominator int64) (quotient int64, remainder float64, _remainder int64) {
	_fnumerator := int64(math.Floor(numerator))
	_remainder = int64((numerator - float64(_fnumerator)) * 1000)

	quotient = _fnumerator / denominator // integer division, decimals are truncated
	remainder = numerator - float64(_fnumerator) + float64(_fnumerator%denominator)
	return
}

func time2time(secs float64) string {
	var s, _m float64
	var h, m, ms int64
	m, s, ms = divmod(secs, 60)
	h, _m, _ = divmod(float64(m), 60)

	return fmt.Sprintf("%d:%02d:%02d,%03d", h, int64(_m), int64(s), ms)
}

func getTimedWordsList(wResp string) []WordStartEndTime {
	var textPieces []string
	var timeRanges []BeginEndTime
	var arr ResultJSON
	err := json.Unmarshal([]byte(wResp), &arr)
	if err != nil {
		fmt.Println(err)
	}

	for _, r := range arr.Results {
		if len(r.Alternatives[0].Transcript) > 100 {
			var chunkTranscritp string
			var chunkBegin float64
			var chunkEnd float64
			for i, wset := range r.Alternatives[0].WordsStartEndTimes {
				if i == 0 || chunkBegin == 0 {
					chunkBegin, _ = strconv.ParseFloat(wset.Time.Begin, 64)
				}

				chunkTranscritp += wset.Word + " "

				chunkEnd, _ = strconv.ParseFloat(wset.Time.End, 64)

				if len(chunkTranscritp) >= 100 || chunkEnd-chunkBegin > 10 {

					textPieces = append(textPieces, chunkTranscritp)
					timeRanges = append(timeRanges, BeginEndTime{strconv.FormatFloat(chunkBegin, 'f', 2, 64), strconv.FormatFloat(chunkEnd, 'f', 2, 64)})

					chunkTranscritp = ""
					chunkBegin = 0
					chunkEnd = 0
				}
			}
			if chunkTranscritp != "" {
				textPieces = append(textPieces, chunkTranscritp)
				timeRanges = append(timeRanges, BeginEndTime{strconv.FormatFloat(chunkBegin, 'f', 2, 64), strconv.FormatFloat(chunkEnd, 'f', 2, 64)})
			}
		} else {
			textPieces = append(textPieces, r.Alternatives[0].Transcript)
			timeRanges = append(timeRanges, BeginEndTime{r.Alternatives[0].WordsStartEndTimes[0].Time.Begin, r.Alternatives[0].WordsStartEndTimes[len(r.Alternatives[0].WordsStartEndTimes)-1].Time.End})
		}
	}

	return zipTextNTime(textPieces, timeRanges)
}

func MakeSRTFromJSONChunks(recogniseResults app.SubtitlesRecognitionResultArray) []string {
	var tempChunk []string
	var timeOffset float64
	var positionOffset int

	for i, recResult := range recogniseResults {
		timedWordsList := getTimedWordsList(recResult.Chunk)

		for j, chunk := range timedWordsList {
			tempChunk = append(tempChunk, strconv.Itoa(j+1 + positionOffset))
			b, _ := strconv.ParseFloat(chunk.Time.Begin, 64)
			e, _ := strconv.ParseFloat(chunk.Time.End, 64)
			if i != 0 {
				b += timeOffset
				e += timeOffset
			}
			begin := time2time(b)
			end := time2time(e)
			tempChunk = append(tempChunk, fmt.Sprintf("%s --> %s", begin, end))
			tempChunk = append(tempChunk, chunk.Word+"\n")

			if len(timedWordsList) - 1 == j {
				timeOffset = recResult.Duration
				positionOffset += j + 1
			}
		}
	}

	return tempChunk
}