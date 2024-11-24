#!/bin/bash

ARCHIVE_NAME="project.zip"

INCLUDE=(
    "Makefile"
    "cmd"
    "docs"
    "engine"
    "fetcher"
    "go.mod"
    "storage"
    "scripts"
)

if [ -f "$ARCHIVE_NAME" ]; then
    echo "Removing existing archive: $ARCHIVE_NAME"
    rm "$ARCHIVE_NAME"
fi

echo "Creating archive: $ARCHIVE_NAME"
zip -r "$ARCHIVE_NAME" "${INCLUDE[@]}" > /dev/null

if [ $? -eq 0 ]; then
    echo "Archive created successfully: $ARCHIVE_NAME"
    echo ""

    echo "Contents of the archive:"
    unzip -l "$ARCHIVE_NAME"

    ARCHIVE_SIZE_BYTES=$(stat -f%z "$ARCHIVE_NAME")
    ARCHIVE_SIZE_KB=$(echo "scale=2; $ARCHIVE_SIZE_BYTES / 1024" | bc)
    ARCHIVE_SIZE_MB=$(echo "scale=2; $ARCHIVE_SIZE_KB / 1024" | bc)
    echo ""
    echo "Archive size:"
    echo "  - $ARCHIVE_SIZE_BYTES bytes"
    echo "  - $ARCHIVE_SIZE_KB KB"
    echo "  - $ARCHIVE_SIZE_MB MB"
else
    echo "Failed to create archive."
    exit 1
fi
