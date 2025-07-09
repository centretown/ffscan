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
	verbose     = true
	verboseHelp = "display all messages"
	iscopy      = false
	iscopyHelp  = "copy only"
)

var (
	IsWindows  bool
	CurrentDir string
)

func init() {
	flag.BoolVar(&iscopy, "c", iscopy, iscopyHelp)
	flag.StringVar(&inDir, "i", inDir, inHelp)
	flag.StringVar(&outDir, "o", outDir, outHelp)
	flag.StringVar(&script, "s", script, scriptHelp)
	flag.BoolVar(&write, "w", write, writeHelp)
	flag.BoolVar(&verbose, "v", verbose, verboseHelp)
	IsWindows = os.IsPathSeparator('\\')
}

func main() {
	var (
		err  error
		info os.FileInfo
	)

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
		log.Println(CurrentDir, outDir)
		outDir = filepath.Join(CurrentDir, outDir)
		log.Println("result after join", outDir)
	}

	builder := &FFBuilder{in: inDir, out: outDir, script: script}
	// build folder information
	_, err = Build(inDir, outDir, script, builder, write, verbose)

	if err != nil {
		log.Fatalln(err)
	}
}
