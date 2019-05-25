package audio

import (
	"bytes"
	"fmt"
	"github.com/3d0c/gmf"
	"log"
	"runtime/debug"
)

func fatal(err error) {
	debug.PrintStack()
	log.Fatal(err)
}

type AudioChunk struct {
	Data      *bytes.Buffer
	Duration  float64
	CodecName string
}

type AudioSplitter struct {
	data     *bytes.Buffer
	InputCtx *gmf.FmtCtx
}

func (as *AudioSplitter) writer(bts []byte) int {
	newDataLen := len(bts)

	if as.data == nil {
		copyBts := make([]byte, newDataLen)
		copy(copyBts, bts)

		as.data = bytes.NewBuffer(copyBts)

		return newDataLen
	}

	n, _ := as.data.Write(bts)

	return n
}

func (as *AudioSplitter) splitInput(audios chan *AudioChunk) {
	srcStream, err := as.InputCtx.GetBestStream(gmf.AVMEDIA_TYPE_AUDIO)
	if err != nil {
		fmt.Println("GetStream error ", err)
	}
	var outFormatName = ConvertCodecName2Format(srcStream.CodecCtx().Codec().Name())
	outputCtx, err := gmf.NewOutputCtx(gmf.FindOutputFmt(outFormatName, "", ""))
	if err != nil {
		fmt.Println(err)
		return
	}

	avioCtx, err := gmf.NewAVIOContext(outputCtx, &gmf.AVIOHandlers{WritePacket: as.writer})
	if err != nil {
		fmt.Println(err)
	}
	defer gmf.Release(avioCtx)

	outputCtx.SetPb(avioCtx)
	outputCtx.SetFlag(128)

	_, err = outputCtx.AddStreamWithCodeCtx(srcStream.CodecCtx())
	if err != nil {
		fmt.Println("add stream error ", err)
	}

	outStream, err := outputCtx.GetBestStream(gmf.AVMEDIA_TYPE_AUDIO)
	if err != nil {
		fatal(err)
	}

	defer outputCtx.Close()

	//outputCtx.Dump()
	if err := outputCtx.WriteHeader(); err != nil {
		fatal(err)
	}

	var aNum int = 1
	var duration int

	for {
		packet, err := as.InputCtx.GetNextPacket()
		if err != nil {
			break
		}
		if packet.StreamIndex() != srcStream.Index() {
			packet.Free()
			continue
		}

		duration = packet.Time(srcStream.TimeBase())
		packet.SetStreamIndex(outStream.Index())
		if duration < 150*aNum {
			if err := outputCtx.WritePacket(packet); err != nil {
				fatal(err)
			}
		} else {
			aNum++

			outputCtx.WriteTrailer()
			avioCtx.Flush()

			ac := AudioChunk{as.data, float64(duration), outFormatName}
			audios <- &ac

			as.data = nil

			if err := outputCtx.WriteHeader(); err != nil {
				fatal(err)
			}

			if err := outputCtx.WritePacket(packet); err != nil {
				fatal(err)
			}
		}

		packet.Free()
	}

	outputCtx.WriteTrailer()
	avioCtx.Flush()
	//var f, _ = os.OpenFile("ex.webm", os.O_CREATE|os.O_WRONLY, os.ModePerm)
	//as.data.WriteTo(f)
	audios <- &AudioChunk{as.data, float64(duration), outFormatName}
	close(audios)
}

func (as *AudioSplitter) GetAudioChunks() chan *AudioChunk {
	audios := make(chan *AudioChunk)

	go as.splitInput(audios)

	return audios
}
