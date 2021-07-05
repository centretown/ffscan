// Copyright 2018 Dave Marsh. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// mkv.go scan mkv files helper
package main

import (
	"os"
	"strings"
	"time"

	"github.com/remko/go-mkvparse"
)

// MyParser - mkv parser
type MyParser struct {
	audioCodec string
	videoCodec string
}

// HandleMasterBegin -
func (p *MyParser) HandleMasterBegin(id mkvparse.ElementID, info mkvparse.ElementInfo) (bool, error) {
	switch id {
	case mkvparse.SegmentElement, mkvparse.TracksElement:
		return true, nil
	case mkvparse.TrackEntryElement:
		return true, nil
	}
	return false, nil
}

// HandleMasterEnd -
func (p *MyParser) HandleMasterEnd(id mkvparse.ElementID, info mkvparse.ElementInfo) error {
	//fmt.Printf("%s%s\n", indent(info.Level), mkvparse.NameForElementID(id))
	return nil
}

// HandleString -
func (p *MyParser) HandleString(id mkvparse.ElementID, value string, info mkvparse.ElementInfo) (err error) {
	if id == mkvparse.CodecIDElement {
		//fmt.Printf("codec %v: %q\n", mkvparse.NameForElementID(id), value)
		switch value[0:1] {
		case "A":
			p.audioCodec = value[2:]
		case "V":
			p.videoCodec = value[2:]
		}
	}
	return
}

// interface stubs

// HandleInteger -
func (p *MyParser) HandleInteger(id mkvparse.ElementID, value int64, info mkvparse.ElementInfo) (err error) {
	return
}

// HandleFloat -
func (p *MyParser) HandleFloat(id mkvparse.ElementID, value float64, info mkvparse.ElementInfo) (err error) {
	return
}

// HandleDate -
func (p *MyParser) HandleDate(id mkvparse.ElementID, value time.Time, info mkvparse.ElementInfo) (err error) {
	return
}

// HandleBinary -
func (p *MyParser) HandleBinary(id mkvparse.ElementID, value []byte, info mkvparse.ElementInfo) (err error) {
	return
}

var handler = MyParser{}

func mkvContains(fname string, codecs ...string) (ok bool, err error) {
	var file *os.File
	file, err = os.Open(fname)
	if err != nil {
		return
	}

	defer file.Close()
	err = mkvparse.ParseSections(file, &handler, []mkvparse.ElementID{mkvparse.TracksElement}...)
	if err != nil {
		return
	}

	v, a := handler.videoCodec, handler.audioCodec
	for _, c := range codecs {
		if !strings.Contains(v, c) && !strings.Contains(a, c) {
			return
		}
	}

	ok = true
	return
}
