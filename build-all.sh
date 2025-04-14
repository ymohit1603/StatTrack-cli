#!/bin/bash

APP_NAME="stattrack"
OUTPUT_DIR="build"

# List of OS/ARCH combinations
platforms=(
  "darwin amd64"
  "darwin arm64"
  "freebsd 386"
  "freebsd amd64"
  "freebsd arm"
  "linux 386"
  "linux amd64"
  "linux arm"
  "linux arm64"
  "netbsd 386"
  "netbsd amd64"
  "netbsd arm"
  "openbsd 386"
  "openbsd amd64"
  "openbsd arm"
  "openbsd arm64"
  "windows 386"
  "windows amd64"
  "windows arm64"
)

mkdir -p "$OUTPUT_DIR"

for platform in "${platforms[@]}"; do
  os=$(echo $platform | cut -d' ' -f1)
  arch=$(echo $platform | cut -d' ' -f2)
  
  output_name="${APP_NAME}-${os}-${arch}"
  if [ "$os" = "windows" ]; then
    output_name="${output_name}.exe"
  fi

  env GOOS=$os GOARCH=$arch go build -o "$OUTPUT_DIR/$output_name"

  # Zip it
  (cd "$OUTPUT_DIR" && zip "${APP_NAME}-${os}-${arch}.zip" "$output_name" && rm "$output_name")
done
