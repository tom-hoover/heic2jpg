#!/usr/bin/env python3
"""Generates the JPEG source for the HEIC test fixture.

Run via testdata/make_fixture.sh, which then encodes the result to HEIC
with heif-enc. Checked in so the fixture can be regenerated, but the
resulting .heic is what the tests actually read.
"""
from PIL import Image

W, H = 200, 120
img = Image.new("RGB", (W, H))
px = img.load()
for y in range(H):
    for x in range(W):
        # Distinctive gradient plus a solid red corner block, so a
        # flipped, rotated, or misdecoded output is obvious.
        if x < 40 and y < 40:
            px[x, y] = (255, 0, 0)
        else:
            px[x, y] = (x * 255 // W, y * 255 // H, 128)

exif = Image.Exif()
exif[0x0112] = 6                      # Orientation: rotate 90 CW
exif[0x010F] = "TestMake"             # Make
exif[0x0110] = "TestModel"            # Model
exif[0x0132] = "2020:01:02 03:04:05"  # DateTime

img.save("testdata/sample.jpg", quality=100, exif=exif)
print(f"wrote testdata/sample.jpg {W}x{H}")
