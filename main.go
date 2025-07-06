package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
)

var (
	out         = "gen"
	outHelp     = "output folder root"
	script      = "run"
	scriptHelp  = "name of output generating script"
	write       = true
	writeHelp   = "write output folders and scripts"
	verbose     = true
	verboseHelp = "display all messages"
	iscopy      = false
	iscopyHelp  = "copy only"
)

var IsWindows bool

func init() {
	flag.BoolVar(&iscopy, "c", iscopy, iscopyHelp)
	flag.StringVar(&out, "o", out, outHelp)
	flag.StringVar(&script, "s", script, scriptHelp)
	flag.BoolVar(&write, "w", write, writeHelp)
	flag.BoolVar(&verbose, "v", verbose, verboseHelp)
	IsWindows = os.IsPathSeparator('\\')
}

func main() {
	var (
		in  string
		err error
	)

	flag.Parse()
	flag.VisitAll(func(f *flag.Flag) {
		fmt.Printf("%s=%v\n", f.Name, f.Value)
	})

	// in folder is current working directory
	in, err = os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}

	builder := &myBuilder{in: in, out: out, script: script}
	// build folder information
	_, err = Build(in, out, script, builder, write, verbose)

	if IsWindows {
		log.Println("Windows Build")
	}

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
	fcopyw = `copy "%s" "%s` + string(os.PathSeparator) + `%s"` + "\n"
	fcopy  = `cp "%s" "%s` + string(os.PathSeparator) + `%s"` + "\n"
	ffcpy  = `ffmpeg -y -i "%s" -metadata title="%s" -c copy "%s` + string(os.PathSeparator) + `%s"` + "\n"
	fhevc  = `ffmpeg -y -i "%s" -bufsize 10240k -filter:a loudnorm -metadata title="%s" -c:v libx265 -c:a aac "%s` + string(os.PathSeparator) + `%s"` + "\n"
	fnorm  = `ffmpeg -y -i "%s" -bufsize 1024k -filter:a loudnorm -ab 128k -map_metadata 0 -id3v2_version 3 "%s` + string(os.PathSeparator) + `%s"` + "\n"
)

func (b *myBuilder) Format(info os.FileInfo, folder *Folder) (cmd string) {
	destination := folder.Destination
	name := info.Name()
	ext := path.Ext(name)
	title := strings.TrimSuffix(name, ext)

	switch strings.ToLower(ext) {
	case ".mkv":
		if iscopy {
			cmd = fmt.Sprintf(ffcpy, name, title, destination, name)
		} else {
			cmd = fmt.Sprintf(fhevc, name, title, destination, name)
		}

	case ".avi", ".mp4", ".mpeg", ".mpg", ".wmv", ".webm":
		// Convert to x265/aac and normalize audio.
		cmd = fmt.Sprintf(fhevc, name, title, destination, title+".mkv")

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
