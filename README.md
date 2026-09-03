# heic2jpg

Converts every HEIC image in a directory to a high-quality JPEG, preserving
EXIF metadata — capture date, GPS, and the orientation tag that decides
which way up a phone photo appears.

Originals are never modified. Each JPEG is written to a temporary file and
renamed into place, so an interrupted run never leaves a truncated image
behind.

Depends on [`github.com/tom-hoover/darkroom`](https://github.com/tom-hoover/darkroom)
for decoding and JPEG encoding.

## Build

Requires Go 1.27+ and a C compiler (cgo builds the bundled HEVC decoder). No
system libraries are needed.

```sh
go build .
```

## Usage

```
heic2jpg [flags] [directory]      # directory defaults to "."

  -q N     JPEG quality (1-100)                                  (default 95)
  -out D   write JPEGs to D instead of beside the originals
  -r       descend into subdirectories
  -f       overwrite existing JPEGs                       (default: skip them)
  -j N     files to convert in parallel                    (default: all CPUs)
  -n       list what would be converted, writing nothing
  -v       report each file as it is converted
```

```sh
heic2jpg ~/Pictures                       # convert in place
heic2jpg -r -out ~/jpgs ~/Pictures        # mirror the tree elsewhere
heic2jpg -n ~/Pictures                    # see what would happen first
```

`IMG_1234.HEIC` becomes `IMG_1234.jpg` at full resolution. Files whose JPEG
already exists are skipped, so re-running after adding photos costs nothing.
A file that fails to decode is reported on stderr and the rest of the batch
continues; the exit status is 1 if anything failed.

## What it will not do

It only converts HEIC to JPEG. It does not resize, re-encode any other
format, edit pixels, or touch the source files in any way.

## Tests

```sh
go test ./...
```
