# flac-to-aac

A small Go CLI that walks a folder tree, finds every `.flac` file, and transcodes them to **AAC 320 kbps** in an **`.m4a`** container (AAC-LC via `ffmpeg`), suitable for Apple Music and similar players.

The **output folder layout matches the input**: same subfolders and file names, with `.m4a` instead of `.flac`.

## Requirements

- [Go](https://go.dev/) (see `go.mod` for the toolchain version)
- [ffmpeg](https://ffmpeg.org/) on your `PATH`

## Build

```bash
go build -o flac-to-aac .
```

## Usage

Run the binary and answer two prompts:

1. **Input directory** — root folder containing your FLAC files (nested folders are fine).
2. **Output directory** — root where converted `.m4a` files are written; subfolders are created as needed.

Example: if input is `~/Music/Albums/Artist/01 Track.flac`, output might be `~/AAC/Albums/Artist/01 Track.m4a`.

To keep `.m4a` files **next to** the originals with the same structure, use the **same path** for both input and output.

## Behavior notes

- **Paths from the terminal** — If you paste a path that looks like `My\ Album/track.flac` (backslashes before spaces), the tool strips shell-style escapes so it matches the real folder on disk.
- **Color** — Prompts and progress use ANSI colors when supported. Set `NO_COLOR` (any value) or `TERM=dumb` to disable them.
