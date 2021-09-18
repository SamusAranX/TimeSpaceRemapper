package main

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	mem "github.com/shirou/gopsutil/mem"
	"image"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	version = "0.1"
)

func main() {
	binName := filepath.Base(os.Args[0])
	binNameNoExt := strings.TrimSuffix(binName, filepath.Ext(binName))

	opts := CommandLineOpts{}

	parser := flags.NewParser(&opts, flags.Default)

	_, err := parser.Parse()
	if err != nil {
		os.Exit(0)
	}

	if opts.Version {
		fmt.Printf("%s v%s - another useless tool by Peter Wunder\n", binNameNoExt, version)
		fmt.Println("-----")
		fmt.Printf("example usage: %s -i \"frame_directory\" -p \"*.png\" -o \"output_dir\"\n", binName)
		fmt.Println("               add -M to the options if you have loads of memory to keep frames in")
		fmt.Println("               don't worry, it'll auto-disable itself if you don't")
		os.Exit(0)
	}

	files, err := ioutil.ReadDir(opts.InputDir)
	if err != nil {
		panic(err)
	}

	var filePaths []string
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if opts.InputPattern != "" {
			match, _ := filepath.Match(opts.InputPattern, file.Name())
			if !match {
				continue
			}
		}

		filePaths = append(filePaths, filepath.Join(opts.InputDir, file.Name()))
	}

	numInputFrames := len(filePaths)
	fmt.Printf("%d files found\n", numInputFrames)

	fmt.Println("Preflight check…")

	var height, width int
	for _, path := range filePaths {
		imageFile, err := os.Open(path)
		if err != nil {
			panic(err)
		}

		config, _, err := image.DecodeConfig(imageFile)
		if err != nil {
			panic(err)
		}

		if imageFile.Close() != nil {
			panic(err)
		}

		sizeEmpty := height == 0 && width == 0
		sizeEqual := config.Height == height && config.Width == width

		if !sizeEmpty && !sizeEqual {
			panic("All frames must be of equal size")
		}

		height = config.Height
		width = config.Width
	}

	if width <= 0 {
		panic("Invalid width")
	}

	fmt.Printf("%dx%d\n", width, height)

	if opts.MemoryHog {
		vmmem, err := mem.VirtualMemory()
		if err != nil {
			panic(err)
		}
		bytesPerFrame := width * height * 4

		framesInMemory := float64(vmmem.Free) / float64(bytesPerFrame)
		framesInMemoryRounded := int(math.Floor(framesInMemory))
		if framesInMemoryRounded < numInputFrames {
			fmt.Printf("You have enough free memory for %d frames, but %d frames were found.\n",
				framesInMemoryRounded,
				numInputFrames)
			opts.MemoryHog = false
			fmt.Println("Disabling Hog Mode.")
			time.Sleep(1 * time.Second)
		}
	}

	fmt.Println("Building new frames…")
	inputFrames := make([]image.Image, numInputFrames)

	// there will be as many new frames as pixels the input frames are wide
	for frameX := opts.StartIndex; frameX < width; frameX++ {

		// new frames are as high as the input frames
		// but as wide as the total number of input frames
		img := image.NewRGBA(image.Rect(0, 0, numInputFrames, height))

		percent := -1

		// loop through input frames and fill new frames with 1px-wide pixel columns
		for dstX, path := range filePaths {
			var frameRaw image.Image

			if inputFrames[dstX] == nil {
				frameFile, err := os.Open(path)
				if err != nil {
					panic(err)
				}

				frameRaw, _, err = image.Decode(frameFile)
				if err != nil {
					panic(err)
				}

				if frameFile.Close() != nil {
					panic(err)
				}

				if opts.MemoryHog {
					inputFrames[dstX] = frameRaw
				}
			} else {
				frameRaw = inputFrames[dstX]
			}

			srcRect := image.Rect(frameX, 0, frameX+1, height)
			dstPt := image.Pt(dstX, 0)
			r := srcRect.Sub(srcRect.Min).Add(dstPt)

			//fmt.Println(dstX)
			//fmt.Printf("srcRect: %v\n", srcRect)
			//fmt.Printf("  dstPt: %v\n", dstPt)
			//fmt.Printf("      r: %v\n", r)
			//fmt.Println()

			draw.Draw(img, r, frameRaw, srcRect.Min, draw.Src)

			newPercent := int(math.Floor(float64(dstX) / float64(numInputFrames) * 100))
			if newPercent >= percent+1 {
				percent = newPercent
				fmt.Printf("%03d%% (%d/%d)\n",
					newPercent,
					dstX,
					numInputFrames)
			}
		}

		fmt.Printf("%03d%% (%d/%d)\n",
			100,
			numInputFrames,
			numInputFrames)

		formatStr := fmt.Sprintf("frame%%0%dd.png", len(strconv.Itoa(width)))

		newFileName := filepath.Join(opts.OutputDir, fmt.Sprintf(formatStr, frameX))

		newFrameFile, err := os.Create(newFileName)
		if err != nil {
			panic(err)
		}
		err = png.Encode(newFrameFile, img)
		if err != nil {
			panic(err)
		}

		_ = newFrameFile.Close()
	}

}
