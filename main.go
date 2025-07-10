package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

const blotto = "ftp://192.168.10.161/Public/Shared%20Videos"

var (
	inDir       = ""
	inHelp      = "input folder root"
	outDir      = "gen"
	outHelp     = "output folder root"
	script      = "run"
	scriptHelp  = "name of output script"
	write       = true
	writeHelp   = "write output folders and scripts"
	verbose     = false
	verboseHelp = "display all messages"
	iscopy      = false
	iscopyHelp  = "copy only"
	isH264      = false
	isH264Help  = "use libx264 to encode"
)

var (
	IsWindows  = os.IsPathSeparator('\\')
	CurrentDir string
)

func main() {
	var (
		err  error
		info os.FileInfo
	)

	flag.BoolVar(&iscopy, "c", iscopy, iscopyHelp)
	flag.StringVar(&inDir, "i", inDir, inHelp)
	flag.StringVar(&outDir, "o", outDir, outHelp)
	flag.StringVar(&script, "s", script, scriptHelp)
	flag.BoolVar(&write, "w", write, writeHelp)
	flag.BoolVar(&verbose, "v", verbose, verboseHelp)
	flag.BoolVar(&isH264, "f", isH264, isH264Help)

	flag.Parse()
	flag.VisitAll(func(f *flag.Flag) {
		fmt.Printf("%s=%v\n", f.Name, f.Value)
	})

	// in folder is current working directory
	CurrentDir, err = os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}

	if len(inDir) == 0 {
		inDir = CurrentDir
	} else {
		inDir, err = filepath.Abs(inDir)
	}

	info, err = os.Stat(inDir)
	if os.IsNotExist(err) {
		log.Fatalf("%s does not exist\n", inDir)
		return
	}

	if !info.IsDir() {
		log.Fatalf("%s is not a directory", inDir)
		return
	}

	// ensure absolute output folder
	if !filepath.IsAbs(outDir) {
		msgln(CurrentDir, outDir)
		outDir = filepath.Join(CurrentDir, outDir)
		msgln("result after join", outDir)
	}

	builder := &FFBuilder{in: inDir, out: outDir, script: script, isH264: isH264}

	// build folder information
	_, err = Build(inDir, outDir, script, builder, write, verbose)

	if err != nil {
		log.Fatalln(err)
	}
}

func msgf(format string, args ...any) {
	if verbose {
		log.Printf(format, args...)
	}
}

func msgln(args ...any) {
	if verbose {
		log.Println(args...)
	}
}
