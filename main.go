package main

import (
	"fmt"
	"github.com/notedit/gst"
	"github.com/notedit/resample"
)

var pipelineStr = "audiotestsrc wave=sine ! audio/x-raw, format=S16LE,rate=48000, channels=2 ! audioconvert  ! opusenc ! appsink name=sink"

type Transcode struct {
	inSampleFormat  resample.SampleFormat
	outSampleFormat resample.SampleFormat
	enc             *resample.AudioEncoder
	dec             *resample.AudioDecoder

	inChannels    int
	outChannels   int
	outbitrate    int
	inSampleRate  int
	outSampleRate int
}

func (t *Transcode) Setup() error {
	dec, err := resample.NewAudioDecoder("libopus")
	if err != nil {
		return err
	}
	dec.SetSampleRate(t.inSampleRate)
	dec.SetSampleFormat(t.inSampleFormat)
	dec.SetChannels(t.inChannels)
	err = dec.Setup()
	if err != nil {
		return err
	}
	t.dec = dec
	enc, err := resample.NewAudioEncoder("aac")
	if err != nil {
		return err
	}
	enc.SetSampleRate(t.outSampleRate)
	enc.SetSampleFormat(t.outSampleFormat)
	enc.SetChannels(t.outChannels)
	enc.SetBitrate(t.outbitrate)
	err = enc.Setup()
	if err != nil {
		return err
	}
	t.enc = enc
	return nil
}

func (t *Transcode) SetInSampleRate(samplerate int) error {
	t.inSampleRate = samplerate
	return nil
}

func (t *Transcode) SetInChannels(channels int) error {
	t.inChannels = channels
	return nil
}

func (t *Transcode) SetInSampleFormat(sampleformat resample.SampleFormat) error {
	t.inSampleFormat = sampleformat
	return nil
}

func (t *Transcode) SetOutSampleRate(samplerate int) error {
	t.outSampleRate = samplerate
	return nil
}

func (t *Transcode) SetOutChannels(channels int) error {
	t.outChannels = channels
	return nil
}

func (t *Transcode) SetOutSampleFormat(sampleformat resample.SampleFormat) error {
	t.outSampleFormat = sampleformat
	return nil
}

func (t *Transcode) SetOutBitrate(bitrate int) error {
	t.outbitrate = bitrate
	return nil
}

func (t *Transcode) Do(data []byte) (out [][]byte, err error) {

	var frame resample.AudioFrame
	var ok bool
	if ok, frame, err = t.dec.Decode(data); err != nil {
		return
	}

	if !ok {
		fmt.Println("does not get one frame")
		return
	}

	if out, err = t.enc.Encode(frame); err != nil {
		return
	}

	return
}

func (t *Transcode) Close() {
	t.enc.Close()
	t.dec.Close()
}

func main() {

	trans := &Transcode{}

	trans.SetInSampleRate(48000)
	trans.SetInChannels(2)
	trans.SetInSampleFormat(resample.S16)
	trans.SetOutChannels(2)
	trans.SetOutSampleFormat(resample.FLTP)
	trans.SetOutSampleRate(48000)
	trans.SetOutBitrate(48000)

	err := trans.Setup()
	if err != nil {
		fmt.Println(err)
	}

	pipeline, err := gst.ParseLaunch(pipelineStr)

	if err != nil {
		panic(err)
	}

	element := pipeline.GetByName("sink")
	pipeline.SetState(gst.StatePlaying)

	for {

		sample, err := element.PullSample()
		if err != nil {
			if element.IsEOS() == true {
				fmt.Println("eos")
				return
			} else {
				fmt.Println(err)
				continue
			}
		}

		_, err = trans.Do(sample.Data)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("got sample ", sample.Duration)
	}

	pipeline.SetState(gst.StateNull)
	pipeline = nil
	element = nil
}
