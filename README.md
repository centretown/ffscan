# ffscan
### Command line tool to scan and convert media files.

The tool generates a hierarchy of shell scripts that mirrors the current folder's structure and converts targeted media files to HEVC and MP3 formats.  All files are volume 'normalized'.

Scans for ".avi", ".mp4", ".mpeg", ".mpg" and ".mkv" file extensions and generates
bash scripts that use ffmpeg to convert these files to HEVC/AAC files with an ".mkv"
extension.

Scans for ".flac" and ".mp3" file extensions and generates bash scripts that use ffmpeg
to convert these files to mp3.

The volume in all audio streams are normalized or brought to a standard level.

The folder structure is preserved under the destination base folder as designated by the -o flag.

The scripts are hierarchical, in that running a generated script will also run the generated scripts
in each of its descendant folders.

```
Usage:
	-h help:
	-o output: output folder (default "gen")
	-s script: name of output script (default "run")
	-v verbose: display all messages
	-w write: create output folders and write shell command scripts (default true)
```
