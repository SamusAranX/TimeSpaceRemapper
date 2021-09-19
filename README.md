# TimeSpaceRemapper
Remaps time to space

## Build

`go get -u` followed by `go build -o tsr` or `go build -o tsr.exe` if you're on Windows.

## Usage

```text
$ ./tsr -h
Usage:
  tsr [OPTIONS]

Application Options:
  -i, --input-dir=   Input frame directory
  -o, --output-dir=  Output frame directory
  -p, --pattern=     Input file name glob pattern (optional)
  -s, --start-index= Starting index (default: 0)
  -M, --memory-hog   Hog Mode (will attempt to keep all new frames in memory)
  -V, --verbose      Outputs more status messages
  -v, --version      Show version and exit

Help Options:
  -h, --help         Show this help message
```

`-i` and `-o` are required.

This tool operates on frame sequences, not video files. Use ffmpeg to extract a video's frames like this:

`ffmpeg -i videogoeshere.mp4 -vsync 0 frames/frame%06d.png`

then start TimeSpaceRemapper to generate output frames:

`./tsr -i ./frames -o ./newframes`

You can optionally use `-M` to engage **Hog Mode**, which will try to store input frames in memory, *massively* speeding up frame generation.
Hog Mode will disable itself if it detects there's insufficient free memory. In that case, try closing other memory hogs, downscaling your input frames, or getting more RAM.

## Examples

### Before
![Screenshot of the About window](images/train.gif)

### After
![Screenshot of the About window](images/train_remapped.gif)

Video used: https://www.youtube.com/watch?v=3-MqzjceE9o