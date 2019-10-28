// Copyright 2019 Adam Leyland
// Use of this source code is governed by a BSD-2 style license that can be found in the LICENSE file.

// Main package for BC5 converter CLI tool.
package main

import (
	"errors"
	"fmt"
	"github.com/leylandski/go-bc5"
	"image"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const helpText = `BC5 compression/decompression tool - v1.0 - usage:
	-id, --inputdir			Specifies an input directory. All image files matching supported file type extensions will be converted.
	-i, --input				Specifies an input file. Only this file will be converted.
	-o, --output			Specifies an output directory to write to. If none is specified the working directory is used.
	-c, --compress			Sets the mode to compress the input file into a .bc5 output file. This is the default if neither mode flag is specified.
	-d, --decompress		Sets the mode to decompress the input file into the output directory in the format specified by -of.
	-of, --outformat		Specifies the output format for decompression. Currently only "jpg", "gif", and "png" are supported.
	-b, --blue				Specified the how the blue component is determined during decompression. Acceptable values are:
							0		- Sets every output pixel's blue component to 0.
							1		- Sets every output pixel's blue component to 255.
							gs		- Sets every output pixel's blue component to that of its red component. Use this for greyscale images.
							cn		- Sets every output pixel's blue component to the computed normal, assuming the map was normalised prior to compression.

Examples: 
	bc5-converter.exe -c -i C:\tmp\image.jpeg -o C:\compressed -h
	bc5-converter.exe --compress --inputdir C:\textures
	bc5-converter.exe -d -id C:\compressed -h -o C:\uncompressed -of jpg -b gs
	bc5-converter.exe --decompress -i C:\compressed\image.bc5 -o C:\uncompressed --outformat png --blue 1

`

// Program mode (compress or decompress)
type Mode int

const (
	Compress Mode = iota
	Decompress
)

type OutputFormat int

const (
	Unknown OutputFormat = iota
	JPG
	PNG
	GIF
)

// Parse a string into the output image format.
func parseFormat(s string) OutputFormat {
	switch strings.ToLower(s) {
	case "jpg":
		return JPG
	case "png":
		return PNG
	case "gif":
		return GIF
	default:
		return Unknown
	}
}

// Return a string version of the output format.
func formatExt(f OutputFormat) string {
	switch f {
	case JPG:
		return "jpg"
	case PNG:
		return "png"
	case GIF:
		return "gif"
	default:
		return ""
	}
}

// Return the blue computation mode parsed from a string.
func parseBlueMode(bm string) bc5.BlueMode {
	switch bm {
	case "1":
		return bc5.One
	case "gs":
		return bc5.Greyscale
	case "cn":
		return bc5.ComputeNormal
	default:
		return bc5.Zero
	}
}

var (
	mode     Mode         //Program mode
	isDir    bool         //Operate on a dir
	target   string       //Input target
	outPath  string       //Output path
	outFmt   OutputFormat //Output format (if decompressing)
	blueMode bc5.BlueMode //Blue computation mode
)

// Main entry point
func main() {

	//Print intro info
	fmt.Printf("BC5 compression/decompression tool - v1.0\n")
	fmt.Printf("Copyright 2019 Adam Leyland (https://github.com/leylandski)\n\n")

	//Process args
	argName := ""
	for _, arg := range os.Args[1:] {
		if arg == "-c" || arg == "--compress" {
			mode = Compress
			continue
		}

		if arg == "-d" || arg == "--decompress" {
			mode = Decompress
			continue
		}

		if argName == "" {
			argName = arg
			continue
		}

		switch argName {
		case "-id", "--inputdir":
			isDir = true
			target = arg
		case "-i", "--input":
			target = arg
		case "-o", "--output":
			outPath = arg
		case "-of", "--outformat":
			outFmt = parseFormat(arg)
		case "-b", "--blue":
			blueMode = parseBlueMode(arg)
		default:
			fmt.Printf("Unknown argument: %s.\n%s", argName, helpText)
			os.Exit(1)
		}
		argName = ""
	}
	if mode == Decompress && outFmt == Unknown {
		fmt.Printf("Unsupported output format. Supported formats include PNG, GIF, and JPG.\n%s", helpText)
		os.Exit(1)
	}
	if outPath == "" {
		wd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		outPath = wd
	}

	//Make a list of files to convert
	files := make([]string, 0)
	if isDir {
		//Walk through the filepath and get any files we can convert
		err := filepath.Walk(target, func(path string, info os.FileInfo, err error) error {
			if mode == Compress {
				if strings.HasSuffix(path, ".jpg") || strings.HasSuffix(path, ".png") || strings.HasSuffix(path, ".gif") {
					files = append(files, path)
				}
			} else {
				if strings.HasSuffix(path, ".bc5") {
					files = append(files, path)
				}
			}
			return err
		})
		if err != nil {
			panic(err)
		}
	} else {
		files = append(files, target)
	}
	fmt.Printf("Converting %d files...\n", len(files))

	//Begin compression/decompression
	start := time.Now()
	if mode == Compress {
		for _, filename := range files {
			compressFile(filename)
		}
	} else {
		for _, filename := range files {
			decompressFile(filename)
		}
	}
	end := time.Now()
	timeTaken := end.Sub(start)

	fmt.Printf("Done! Converted %d files in %f seconds (%f files/sec).\n", len(files), timeTaken.Seconds(), float64(len(files))/timeTaken.Seconds())
}

// Compress the given file using the current program settings
func compressFile(filename string) {
	f, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Unable to open %s: %s\n", filename, err.Error())
		os.Exit(1)
	}
	defer f.Close()

	//Decode image to generic
	img, _, err := image.Decode(f)
	if err != nil {
		fmt.Printf("Error reading file: %s\n", err.Error())
		os.Exit(1)
	}

	//Redraw as RGBA
	imgRgba := image.NewRGBA(img.Bounds())
	draw.Draw(imgRgba, imgRgba.Bounds(), img, img.Bounds().Min, draw.Src)

	//Compress the RGBA data to BC5
	fmt.Printf("Compressing %s... ", filename)
	compressed, err := bc5.NewBC5FromRGBA(imgRgba)
	if err != nil {
		panic(err)
	}
	fmt.Print("done.\n")

	//Save the BC5 output
	fnameParts := strings.Split(strings.Replace(filename, "\\", "/", -1), "/")
	outFile, err := os.Create(strings.TrimSuffix(outPath, string(os.PathSeparator)) + string(os.PathSeparator) + fnameParts[len(fnameParts)-1] + ".bc5")
	if err != nil {
		fmt.Printf("Error creating output file: %s\n", err.Error())
		os.Exit(1)
	}
	defer outFile.Close()

	err = bc5.Encode(compressed, outFile)
	if err != nil {
		panic(err)
	}
}

// Decompress the given file using the current program settings
func decompressFile(filename string) {
	f, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Unable to open %s: %s\n", filename, err.Error())
		os.Exit(1)
	}
	defer f.Close()

	//Decode the BC5 data into a struct
	img, err := bc5.Decode(f)
	if err != nil {
		fmt.Printf("Error decoding BC5 data: %s.\n", err.Error())
		os.Exit(1)
	}

	//Decompress using the current settings
	fmt.Printf("Decompressing %s... ", filename)
	img.BlueMode = blueMode
	decomp := img.Decompress()
	fmt.Printf("done.\n")

	//Write the decompressed contents to the output file
	fnameParts := strings.Split(strings.Replace(filename, "\\", "/", -1), "/")
	outFile, err := os.Create(strings.TrimSuffix(outPath, string(os.PathSeparator)) + string(os.PathSeparator) + fnameParts[len(fnameParts)-1] + "." + formatExt(outFmt))
	if err != nil {
		fmt.Printf("Error creating output file: %s.\n", err.Error())
		os.Exit(1)
	}
	defer outFile.Close()

	switch outFmt {
	case PNG:
		err = png.Encode(outFile, decomp)
	case JPG:
		err = jpeg.Encode(outFile, decomp, nil)
	case GIF:
		err = gif.Encode(outFile, decomp, nil)
	default:
		err = errors.New("unsupported output format")
	}
	if err != nil {
		fmt.Printf("Error creating output file: %s.\n", err.Error())
	}
}
