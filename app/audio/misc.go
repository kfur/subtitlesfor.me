package audio

import "strings"

// convert stream codec name to audio format
var CodecName2Format = map[string]string {
	"opus": "webm",
	"vorbis": "webm",
	"mpeg": "mp3",
	"webm": "webm",
	"flac": "flac",
	"mp3": "mp3",
	"ogg": "ogg",
	"wav": "wav",
}

// convert file format extension name to ffmpeg format name
var Ext2Format = map[string]string {
	"mkv": "matroska",
}

func ConvertExt2Format(ext string) string {
	if strings.Contains(ext, ".") {
		ext = ext[1:]
	}
	inFormat := Ext2Format[ext]
	if inFormat == "" {
		inFormat = ext
	}

	return inFormat
}

func ConvertCodecName2Format(codecName string) string {
	format := CodecName2Format[codecName]
	if format == "" {
		format = codecName
	}

	return format
}