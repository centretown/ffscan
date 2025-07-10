package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang-collections/collections/queue"
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

type QItem struct {
	in, out string
}

func enq(qitem *queue.Queue, in, out string) {
	qitem.Enqueue(&QItem{in: in, out: out})
}

var showMessages = false

func Build(inputBase, outputBase, scriptName string,
	builder Builder, generate bool, verbose bool) (folders Folders, err error) {

	var foldersQ = queue.New()

	defer func() {
		if foldersQ != nil {
			for foldersQ.Len() > 0 {
				foldersQ.Dequeue()
			}
		}
	}()

	if verbose {
		showMessages = true
		defer func(perr error) {
			if perr != nil {
				msgf("%s\n", perr.Error())
			}
		}(err)
	}

	msgf("Build in: %v, write: %v, verbose: %v\n", inputBase, generate, verbose)

	err = os.Chdir(inputBase)
	if err != nil {
		return
	}

	enq(foldersQ, inputBase, outputBase)

	folders, err = scanQ(foldersQ, scriptName, builder)
	if err != nil {
		return
	}

	msgln("generate scripts")
	folders.generate(builder)
	os.Chdir(inputBase)

	if generate {
		msgf("create base output folder %v\n", outputBase)
		err = makeDir(outputBase)
		if err != nil {
			return
		}
	}

	sb := &strings.Builder{}
	for _, fld := range folders {
		msgf("create %s\n", fld.Destination)
		err = makeDir(fld.Destination)
		if err != nil {
			log.Fatal(err)
		}
		if len(fld.Files) == 0 {
			continue
		}
		sb.WriteString(`cd "` + fld.Source + `" ` + "\n")
		sb.WriteString(fld.Code)
	}

	err = os.Chdir(CurrentDir)
	if err != nil {
		return
	}

	cmd := sb.String()
	fmt.Print(cmd)
	if IsWindows {
		scriptName += ".cmd"
	}
	msgf("write script '%s' Windows?=%v\n", scriptName, IsWindows)
	err = os.WriteFile(scriptName, []byte(cmd), os.ModeAppend|os.ModePerm)
	if err != nil {
		return
	}

	return
}

func scanQ(foldersQ *queue.Queue, script string, builder Builder) (fs Folders, err error) {
	var (
		folder *Folder
	)
	// scan and filter each folder in the tree
	for foldersQ.Len() > 0 {
		next := foldersQ.Dequeue().(*QItem)
		msgf("scan and filter %v\n", next.in)
		folder, err = scanFolder(foldersQ, next.in, next.out, script, builder)
		if err != nil {
			return
		}
		fs = append(fs, folder)
	}
	return
}

func scanFolder(foldersQ *queue.Queue, in, out, script string, builder Builder) (folder *Folder, err error) {
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
	msgf("scanning: %s\n", in)
	for _, f := range files {
		name = f.Name()
		msgf("scan:%s\n", name)
		info, _ := f.Info()
		if builder.Filter(info) {
			if f.IsDir() {
				inc, outc = filepath.Join(in, name), filepath.Join(out, name)
				enq(foldersQ, inc, outc)
				folder.Children = append(folder.Children, inc)
				msgf("folder:%s selected\n", name)
			} else {
				folder.Files = append(folder.Files, f)
				msgf("file:%s selected\n", name)
			}
		}
	}
	msgf("children:%v", folder.Children)
	return
}

var (
	formatChild  = `cd "%s"` + "\n./%s\ncd ..\n"
	formatChildw = `cd "%s"` + "\ncall %s\ncd ..\n"
)

func (fs Folders) generate(b Builder) (cmd string) {
	for _, f := range fs {
		cmd := ""
		for _, file := range f.Files {
			info, _ := file.Info()
			cmd += b.Format(info, f)
		}
		f.Code = cmd
	}
	return
}

// makeDir - creates new directories if necessary
func makeDir(dir string) (err error) {
	var info os.FileInfo
	info, err = os.Stat(dir)
	if err != nil {
		err = os.Mkdir(dir, os.ModeDir|os.ModePerm)
		if err != nil {
			msgf("makeDir %s\n%v\n", dir, err)
		}
		return
	}

	if info == nil {
		msgf("makedir: No Info")
		return
	}

	if !info.IsDir() {
		err = fmt.Errorf("%s exists but is not a directory", dir)
	}
	return
}
