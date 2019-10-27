package main

const helpText = `BC5 conversion utility 1.0 - usage:
	-id, --inputdir			Specifies an input directory. All image files matching -p will be converted.
	-i, --input				Specifies an input file. Only this file will be converted.
	-o, --output			Specifies an output directory to write to. If none is specified the working directory is used.
	-c, --compress			Sets the mode to compress the input file into a .bc5 output file. This is the default if neither mode flag is specified.
	-d, --decompress		Sets the mode to decompress the input file into the output directory in the format specified by -of.
	-of, --outformat		Specifies the output format for decompression. Currently only "jpg", "gif", and "png" are supported.
	-h, --header			Specifies whether to include a 12 byte header containing "BC5 ", and the width and height in 3 uint32 values.
	-p, --pattern			Specifies the search pattern for input files. If not specified then "*" is assumed.

Examples: 
	bc5-converter -c -i image.jpeg -o ./compressed -h
	bc5-converter --compress --inputdir ../textures -p *.png
	bc5-converter -d -id ./compressed -h -o ./uncompressed -of jpg
	bc5-converter --decompress -i ./compressed/image.bc5 -o ./uncompressed -h --outformat png

`
