package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
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
				log.Printf("%s\n", perr.Error())
			}
		}(err)
	}

	log.Printf("Build in: %v, write: %v, verbose: %v\n", inputBase, generate, verbose)

	err = os.Chdir(inputBase)
	if err != nil {
		return
	}

	// create base output folder if necessary

	// build folder structure from base
	// outputBase = filepath.Join(outputBase, filepath.Base(inputBase))
	enq(inputBase, outputBase)

	// scan and filter each folder in the tree
	folders, err = scanQ(scriptName, builder)
	if err != nil {
		return
	}

	log.Println("generate scripts")
	folders.generate(builder)
	os.Chdir(inputBase)

	if generate {
		log.Printf("create base output folder %v\n", outputBase)
		err = makeDir(outputBase)
		if err != nil {
			return
		}
	}

	sb := &strings.Builder{}
	for _, fld := range folders {
		log.Printf("create %s\n", fld.Destination)
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
	log.Printf("write script '%s' Windows?=%v\n", scriptName, IsWindows)
	err = os.WriteFile(scriptName, []byte(cmd), os.ModeAppend|os.ModePerm)
	if err != nil {
		return
	}

	return

	if generate {
		log.Printf("create output folder %v\n", outputBase)
		err = makeDir(outputBase)
		if err != nil {
			return
		}
		log.Println("write script and create output folders")
		err = folders.Write()
		if err != nil {
			log.Println("write script", err)
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
		log.Printf("scan and filter %v\n", qi)
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
	log.Printf("scanning: %s\n", in)
	for _, f := range files {
		name = f.Name()
		log.Printf("scan:%s\n", name)
		info, _ := f.Info()
		if builder.Filter(info) {
			if f.IsDir() {
				inc, outc = filepath.Join(in, name), filepath.Join(out, name)
				enq(inc, outc)
				folder.Children = append(folder.Children, inc)
				log.Printf("folder:%s selected\n", name)
			} else {
				folder.Files = append(folder.Files, f)
				log.Printf("file:%s selected\n", name)
			}
		}
	}
	log.Printf("children:%v", folder.Children)
	return
}

// format to run child scripts
var (
	// formatChild = `cd "%s\n"` + string(os.PathSeparator) + "%s\ncd ..\n"
	// formatChild = `cd "%s\n"` + string(os.PathSeparator) + "%s\ncd ..\n"
	formatChild  = `cd "%s"` + "\n./%s\ncd ..\n"
	formatChildw = `cd "%s"` + "\ncall %s\ncd ..\n"
)

// generate - create a script for the selected files and folders
func (f *Folder) generate(b Builder) {
	cmd := ""
	for _, file := range f.Files {
		info, _ := file.Info()
		cmd += b.Format(info, f)
	}

	// for _, child := range f.Children {
	// 	if IsWindows {
	// 		cmd += fmt.Sprintf(formatChildw, child, f.Script)
	// 	} else {
	// 		cmd += fmt.Sprintf(formatChild, child, f.Script)
	// 	}
	// }

	f.Code = cmd
	// log.Printf("generate '%s/%s\n%s'\n", f.Source, f.Script, f.Code)
}

func (f *Folder) WriteCmd(cmd []byte) (err error) {
	log.Printf("navigate to %s\n", f.Source)
	err = os.Chdir(f.Source)
	if err != nil {
		return
	}

	log.Printf("write script '%s'\n", f.Script)
	err = os.WriteFile(f.Script, cmd, os.ModeAppend|os.ModePerm)
	if err != nil {
		return
	}

	err = makeDir(f.Destination)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("verify/create destination %s OK\n", f.Destination)
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
		err = f.WriteCmd([]byte(f.Code))
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
		if err != nil {
			log.Printf("makeDir %s\n%v\n", dir, err)
		}
		return
	}

	if info == nil {
		log.Printf("makedir: No Info")
		return
	}

	if !info.IsDir() {
		err = fmt.Errorf("%s exists but is not a directory", dir)
	}
	return
}
