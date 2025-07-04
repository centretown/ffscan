package main

import (
	"fmt"
	"io/fs"
	"os"
	"path"
)

type Folder struct {
	Source      string
	Destination string
	Script      string
	Code        string
	Files       []fs.DirEntry
	Children    []string
}

type Folders []*Folder

type Builder interface {
	Filter(info os.FileInfo) bool
	Format(info os.FileInfo, folder *Folder) string
}

var showMessages = false

func Build(inputBase, outputBase, scriptName string,
	builder Builder, generate bool, verbose bool) (folders Folders, err error) {

	// clear queue
	defer clearQ()

	if verbose {
		showMessages = true
		defer displayError(&err)
	}

	msg(`Build(in: "%v", write: %v, verbose: %v)`, inputBase, generate, verbose)

	err = os.Chdir(inputBase)
	if err != nil {
		return
	}

	if !path.IsAbs(outputBase) {
		// ensure absolute output folder
		outputBase = path.Join(inputBase, outputBase)
	}

	if generate {
		// create base output folder if necessary
		msg("create base output folder %v", outputBase)
		err = makeDir(outputBase)
		if err != nil {
			return
		}
	}

	// build folder structure from base
	outputBase = path.Join(outputBase, path.Base(inputBase))

	enq(inputBase, outputBase)

	// scan and filter each folder in the tree
	folders, err = scanQ(scriptName, builder)
	if err != nil {
		return
	}

	// generate scripts
	msg("generate scripts")
	folders.generate(builder)

	os.Chdir(inputBase)
	if generate {
		// create output folder
		msg("create output folder %v", outputBase)
		err = makeDir(outputBase)
		if err != nil {
			return
		}
		msg("write script and create output folders")
		err = folders.Write()
		if err != nil {
			return
		}
	}
	return
}

func scanQ(script string, builder Builder) (fs Folders, err error) {
	var (
		folder *Folder
		qi, qo string
	)
	// scan and filter each folder in the tree
	for foldersQ.Len() > 0 {
		qi, qo = deq()
		msg("scan and filter %v", qi)
		folder, err = scanFolder(qi, qo, script, builder)
		if err != nil {
			return
		}
		fs = append(fs, folder)
	}
	return
}

func scanFolder(in, out, script string, builder Builder) (folder *Folder, err error) {
	var (
		files           []fs.DirEntry
		name, inc, outc string
	)

	err = os.Chdir(in)
	if err != nil {
		return
	}

	files, err = os.ReadDir(in)
	if err != nil {
		return
	}

	folder = &Folder{Source: in, Destination: out, Script: script}
	msg("scanning: %s", in)
	for _, f := range files {
		name = f.Name()
		msg("scan:%s", name)
		info, _ := f.Info()
		if builder.Filter(info) {
			if f.IsDir() {
				inc, outc = path.Join(in, name), path.Join(out, name)
				enq(inc, outc)
				folder.Children = append(folder.Children, inc)
				msg("folder:%s selected.", name)
			} else {
				folder.Files = append(folder.Files, f)
				msg("file:%s selected.", name)
			}
		}
	}
	msg("children:%v", folder.Children)
	return
}

// format to run child scripts
var (
	formatChild = `cd "%s"` + "\n./%s\ncd ..\n"
)

// generate - create a script for the selected files and folders
func (f *Folder) generate(b Builder) {
	cmd := ""
	for _, file := range f.Files {
		info, _ := file.Info()
		cmd += b.Format(info, f)
	}

	for _, child := range f.Children {
		cmd += fmt.Sprintf(formatChild, child, f.Script)
	}

	f.Code = cmd
	msg("generate '%s/%s\n%s'", f.Source, f.Script, f.Code)
}

// Write - create folders and write output files
func (f *Folder) Write(cmd []byte) (err error) {
	msg("navigate to %s", f.Source)
	err = os.Chdir(f.Source)
	if err != nil {
		return
	}

	msg("write script '%s'", f.Script)
	err = os.WriteFile(f.Script, cmd, os.ModeAppend|os.ModePerm)
	if err != nil {
		return
	}

	msg("verify/create destination %s", f.Destination)
	err = makeDir(f.Destination)

	return
}

// Generate - a script for the selected files and folders
func (fs Folders) generate(b Builder) (cmd string) {
	for _, f := range fs {
		f.generate(b)
	}
	return
}

// Write - folders and output files
func (fs Folders) Write() (err error) {
	for _, f := range fs {
		err = f.Write([]byte(f.Code))
		if err != nil {
			return
		}
	}
	return
}

// makeDir - creates new directories if necessary
func makeDir(dir string) (err error) {
	var info os.FileInfo
	info, err = os.Stat(dir)
	if err != nil {
		err = os.Mkdir(dir, os.ModeDir|os.ModePerm)
	} else if !info.IsDir() {
		err = fmt.Errorf("%s exists but is not a directory", dir)
	}
	return
}

// msg - verbose messages
func msg(format string, a ...interface{}) {
	if showMessages {
		fmt.Printf(format+"\n", a)
	}
}

func displayError(err *error) {
	if *err != nil {
		fmt.Println(*err)
	}
}
