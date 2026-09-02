#!/bin/sh
# Regenerates testdata/sample.heic. Requires python3+PIL and heif-enc.
set -e
cd "$(dirname "$0")/.."
python3 testdata/make_fixture.py
heif-enc -q 100 testdata/sample.jpg -o testdata/sample.heic
rm -f testdata/sample.jpg
echo "wrote testdata/sample.heic"
