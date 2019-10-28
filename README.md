# bc5-converter
BC5 compression/conversion tool. See the library this tool uses for more details on the implementation: https://github.com/leylandski/go-bc5

Current release: https://github.com/leylandski/bc5-converter/releases/tag/v1.0
 
## Installation
1. Have Go (golang) installed. This was developed with `1.11.5`, but earlier versions may work just as well. Use at your own risk.
1. Run `go get github.com/leylandski/bc5-converter` in your terminal to download the source and dependencies (https://github.com/leylandski/go-bc5) into your GOPATH.
1. Navigate to the repo directory and run `go build bc5-converter.go` to build the code into an executable.
 
## Usage
### Compression
This tool can convert PNG, GIF, and JPG data to a raw BC5 encoded block file. 

It uses the **go-bc5** library found here: https://github.com/leylandski/go-bc5 to do so. As per that implementation, it writes a 12-byte header to the start of the stream containing the DWORD/uint32 value "BC5 " in Big Endian order and two subsequent uint32 values denoting the width and height respectively, followed by the block data.
 
The output files are given the extension .bc5.
 
### Decompression
The tool can also decompress .bc5 files into any of the original input formats (PNG, GIF, and JPG). When decompressing, the 12-byte header is expected or the decompression will fail. An output format must be specified if using the decompression mode, and optionally a blue component computation method. See the flags below for more details.
 
## Program Arguments
`-id, --inputdir` - Specifies an input directory. All image files matching supported file type extensions will be converted.
 
`-i, --input` - Specifies an input file. Only this file will be converted.
 
`-o, --output` - Specifies an output directory to write to. If none is specified the working directory is used.
 
`-c, --compress` - Sets the mode to compress the input file into a .bc5 output file. This is the default if neither mode flag is specified.

`-d, --decompress` - Sets the mode to decompress the input file into the output directory in the format specified by -of.

`-of, --outformat` - Specifies the output format for decompression. Currently only "jpg", "gif", and "png" are supported.
 
`-b, --blue` - Specified the how the blue component is determined during decompression. Acceptable values are:
 ```
0		- Sets every output pixel's blue component to 0.
1		- Sets every output pixel's blue component to 255.
gs		- Sets every output pixel's blue component to that of its red component. Use this for greyscale images.
cn		- Sets every output pixel's blue component to the computed normal, assuming the map was normalised prior to compression.
```

## Notes
This uses a relatively untested implementation of the BC5 algorithm. Use at your own risk. Feedback, comments, and suggestions are appreciated.

## TODO
* Allow user to define their own header format.
