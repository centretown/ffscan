package main

import (
	"fmt"
	"os"
	"path"
	"strings"
)

type FFBuilder struct {
	in     string
	out    string
	script string
	isH264 bool
}

func (b *FFBuilder) Filter(info os.FileInfo) bool {
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
	fcopyw = `copy "%s" "%s` + string(os.PathSeparator) + `%s"` + "\n"
	fcopy  = `cp "%s" "%s` + string(os.PathSeparator) + `%s"` + "\n"
	ffcpy  = `ffmpeg -y -i "%s" -metadata title="%s" -c copy "%s` + string(os.PathSeparator) + `%s"` + "\n"
	fhevc  = `ffmpeg -y -i "%s" -bufsize 10240k -filter:a loudnorm -metadata title="%s" -c:v libx265 -c:a aac "%s` + string(os.PathSeparator) + `%s"` + "\n"
	favc   = `ffmpeg -y -i "%s" -bufsize 10240k -filter:a loudnorm -metadata title="%s" -c:v libx264 -c:a aac "%s` + string(os.PathSeparator) + `%s"` + "\n"
	fnorm  = `ffmpeg -y -i "%s" -bufsize 1024k -filter:a loudnorm -ab 128k -map_metadata 0 -id3v2_version 3 "%s` + string(os.PathSeparator) + `%s"` + "\n"
)

func (ffb *FFBuilder) Format(info os.FileInfo, folder *Folder) (cmd string) {
	destination := folder.Destination
	name := info.Name()
	ext := path.Ext(name)
	title := strings.TrimSuffix(name, ext)

	switch strings.ToLower(ext) {
	case ".mkv":
		if iscopy {
			cmd = fmt.Sprintf(ffcpy, name, title, destination, name)
		} else if ffb.isH264 {
			cmd = fmt.Sprintf(favc, name, title, destination, name)
		} else {
			cmd = fmt.Sprintf(fhevc, name, title, destination, name)
		}

	case ".avi", ".mp4", ".mpeg", ".mpg", ".wmv", ".webm":
		// Convert to x265/aac and normalize audio.
		if ffb.isH264 {
			cmd = fmt.Sprintf(favc, name, title, destination, title+".mkv")
		} else {
			cmd = fmt.Sprintf(fhevc, name, title, destination, title+".mkv")
		}

	case ".srt", ".sub", ".idx":
		// Copy sub title files if not there already.
		if IsWindows {
			cmd = fmt.Sprintf(fcopyw, name, destination, name)
		} else {
			cmd = fmt.Sprintf(fcopy, name, destination, name)
		}

	case ".mp3":
		// normalize audio only
		cmd = fmt.Sprintf(fnorm, name, destination, name)

	case ".flac", ".wav":
		// convert to mp3
		cmd = fmt.Sprintf(fnorm, name, destination, title+".mp3")
	}
	return
}
