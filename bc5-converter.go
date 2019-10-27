package main

import (
	"blackbird/compression/bc5"
	"errors"
	"fmt"
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

var (
	mode Mode
	isDir bool
	target string
	outPath string
	outFmt OutputFormat
	usingHeader bool = true //TODO turn this off by default and change when making bc5 into go-bc5
	pattern string = "*"
)

func main() {

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
		case "-h", "--header":
			usingHeader = true
		case "-p", "--pattern":
			pattern = arg
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

	/*patternRegex, err := regexp.Compile(pattern)
	if err != nil {
		fmt.Printf("Invalid search pattern.\n%s", helpText)
		os.Exit(1)
	}*/

	files := make([]string, 0)
	if isDir {
		err := filepath.Walk(target, func(path string, info os.FileInfo, err error) error {
			if strings.HasSuffix(path, ".jpg") || strings.HasSuffix(path, ".png") || strings.HasSuffix(path, ".gif") {
				files = append(files, path)
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

	fmt.Printf("Done! Converted %d files in %f seconds (%f files/sec).\n", len(files), timeTaken.Seconds(), float64(len(files)) / timeTaken.Seconds())
}

func compressFile(filename string) {
	f, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Unable to open %s: %s.\n", filename, err.Error())
		os.Exit(1)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		fmt.Printf("Error reading file: %s.\n", err.Error())
		os.Exit(1)
	}
	imgRgba := image.NewRGBA(img.Bounds())
	draw.Draw(imgRgba, imgRgba.Bounds(), img, img.Bounds().Min, draw.Src)

	fmt.Printf("Compressing %s... ", filename)
	compressed, err := bc5.NewBC5FromRGBA(imgRgba)
	if err != nil {
		panic(err)
	}
	fmt.Print("done.\n")

	outFile, err := os.Create(strings.TrimSuffix(outPath, string(os.PathSeparator)) + string(os.PathSeparator) + filename + ".bc5")
	if err != nil {
		fmt.Printf("Error creating output file: %s.\n", err.Error())
		os.Exit(1)
	}
	defer outFile.Close()

	err = bc5.Encode(compressed, outFile)
	if err != nil {
		panic(err)
	}
}

func decompressFile(filename string) {
	f, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Unable to open %s: %s.\n", err.Error())
		os.Exit(1)
	}
	defer f.Close()

	img, err := bc5.Decode(f)
	if err != nil {
		fmt.Printf("Error decoding BC5 data: %s.\n", err.Error())
		os.Exit(1)
	}

	fmt.Printf("Decompressing %s... ", filename)
	decomp := img.Decompress()
	fmt.Printf("done.\n")

	outFile, err := os.Create(strings.TrimSuffix(outPath, string(os.PathSeparator)) + string(os.PathSeparator) + filename + "." + formatExt(outFmt))
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