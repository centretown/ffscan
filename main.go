// Copyright 2018 Dave Marsh. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// ffscan project main.go

/*
ffscan:
Scans for ".avi", ".mp4", ".mpeg", ".mpg" and ".mkv" file extensions and generates
bash scripts that use ffmpeg to convert these files to HEVC/AAC files with an ".mkv"
extension.

Scans for ".flac" and ".mp3" file extensions and generates bash scripts that use ffmpeg
to convert these files to mp3.

The volume in all audio streams are normalized or brought to a standard level.

The folder structure is preserved under the destination base folder as designated by the -o flag.

The scripts are hierarchical, in that running a generated script will also run the generated scripts
in each of its descendant folders.

Usage:
  -h help
  -o string
        output folder (default "gen")
  -s string
        name of output script (default "run")
  -v    verbose: display all messages
  -w    write: create output folders and write shell command scripts (default true)
*/
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/centretown/scan"
)

var (
	out         = "gen"
	outHelp     = "output folder root"
	script      = "run"
	scriptHelp  = "name of output generating script"
	write       = true
	writeHelp   = "write output folders and scripts"
	verbose     = false
	verboseHelp = "display all messages"
)

func init() {
	flag.StringVar(&out, "o", out, outHelp)
	flag.StringVar(&script, "s", script, scriptHelp)
	flag.BoolVar(&write, "w", write, writeHelp)
	flag.BoolVar(&verbose, "v", verbose, verboseHelp)
}

func main() {
	var (
		in  string
		err error
	)

	flag.Parse()
	//if verbose {
	flag.VisitAll(func(f *flag.Flag) {
		fmt.Printf("%s=%v\n", f.Name, f.Value)
	})
	//}

	// in folder is current working directory
	in, err = os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}

	builder := &myBuilder{in: in, out: out, script: script}
	// build folder information
	_, err = scan.Build(in, out, script, builder, write, verbose)
	if err != nil {
		log.Fatalln(err)
	}
}

type myBuilder struct {
	in     string
	out    string
	script string
}

func (b *myBuilder) Filter(info os.FileInfo) bool {
	name := info.Name()
	if info.IsDir() {
		return name != b.out
	}
	ext := strings.ToLower(path.Ext(name))
	switch ext {
	case ".mkv", ".avi",
		".mp4", ".mpeg",
		".mpg", ".flac",
		".mp3", ".srt",
		".sub", ".idx",
		".wmv", ".wav", ".webm":
		return true
	}
	return false
}

const (
	fcopy = `cp -n "%s" "%s/%s"` + "\n"
	ffcpy = `ffmpeg -y -i "%s" -bufsize 10240k -filter:a loudnorm -metadata title="%s" -c:v copy -c:a aac "%s/%s"` + "\n"
	fhevc = `ffmpeg -y -i "%s" -bufsize 10240k -filter:a loudnorm -metadata title="%s" -c:v libx265 -c:a aac "%s/%s"` + "\n"
	fnorm = `ffmpeg -y -i "%s" -bufsize 1024k -filter:a loudnorm -ab 128k -map_metadata 0 -id3v2_version 3 "%s/%s"` + "\n"
)

func (b *myBuilder) Format(info os.FileInfo, folder *scan.Folder) (cmd string) {
	destination := folder.Destination
	name := info.Name()
	ext := path.Ext(name)
	title := strings.TrimSuffix(name, ext)

	switch strings.ToLower(ext) {
	case ".mkv":
		// ok, err := mkvContains(name, "HEVC", "AAC")
		// Copy x265/aac and normalize audio.
		// 	if err == nil && ok {
		// 	cmd = fmt.Sprintf(ffcpy, name, title, destination, name)
		// } else {
		// 	cmd = fmt.Sprintf(fhevc, name, title, destination, name)
		// }
		// Convert to x265/aac and normalize audio.
		cmd = fmt.Sprintf(fhevc, name, title, destination, name)

	case ".avi", ".mp4", ".mpeg", ".mpg", ".wmv", ".webm":
		// Convert to x265/aac and normalize audio.
		cmd = fmt.Sprintf(fhevc, name, title, destination, title+".mkv")

	case ".srt", ".sub", ".idx":
		// Copy sub title files if not there already.
		cmd = fmt.Sprintf(fcopy, name, destination, name)

	case ".mp3":
		// normalize audio only
		cmd = fmt.Sprintf(fnorm, name, destination, name)

	case ".flac", ".wav":
		// convert to mp3
		cmd = fmt.Sprintf(fnorm, name, destination, title+".mp3")
	}
	return
}
