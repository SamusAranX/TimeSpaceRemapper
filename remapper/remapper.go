package remapper

import (
	"errors"
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/shirou/gopsutil/mem"
	"image"
	"image/draw"
	"image/png"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"strconv"
)

type Remapper struct {
	inputFrames []string
	hogLevel    int
	verbose     bool

	frameWidth, frameHeight int
}

func NewMapper(hogMode []bool, verbose bool) Remapper {
	r := Remapper{
		frameWidth:  -1,
		frameHeight: -1,
		hogLevel:    len(hogMode),
		verbose:     verbose,
	}
	return r
}

func (r *Remapper) loadFiles(inputDir string) {
	r.loadFilesWithPattern(inputDir, "")
}

func (r *Remapper) loadFilesWithPattern(inputDir string, inputPattern string) {
	files, err := ioutil.ReadDir(inputDir)
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if inputPattern != "" {
			match, _ := filepath.Match(inputPattern, file.Name())
			if !match {
				continue
			}
		}

		r.inputFrames = append(r.inputFrames, filepath.Join(inputDir, file.Name()))
	}

	fmt.Printf("%d files found\n", len(r.inputFrames))
}

func (r *Remapper) preflightCheck() error {
	for frameIndex, path := range r.inputFrames {
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

		sizeEmpty := config.Width <= 0 || config.Height <= 0
		sizeEqual := config.Width == r.frameWidth && config.Height == r.frameHeight

		if sizeEmpty {
			return errors.New("invalid frame size")
		}

		if r.frameWidth > 0 && r.frameHeight > 0 && !sizeEqual {
			errMsg1 := "all frames must be of equal size"
			errMsg2 := fmt.Sprintf("frame %d: %dx%d is not %dx%d", frameIndex,
				config.Width, config.Height,
				r.frameWidth, r.frameHeight)
			return errors.New(errMsg1 + "\n" + errMsg2)
		}

		r.frameHeight = config.Height
		r.frameWidth = config.Width
	}

	if r.frameWidth <= 0 || r.frameHeight <= 0 {
		return errors.New("invalid final size")
	}

	return nil
}

//goland:noinspection GoErrorStringFormat
func (r *Remapper) hogCheck() error {
	switch r.hogLevel {
	case 0:
		fmt.Println("Hog Mode disabled")
		break
	case 1:
		vmmem, err := mem.VirtualMemory()
		if err != nil {
			r.hogLevel = 0
			return errors.New("Couldn't determine how much free memory is left, disabling Hog Mode")
		}

		freeMemory := vmmem.Free
		bytesPerFrame := r.frameWidth * r.frameHeight * 4
		framesInMemory := float64(freeMemory) / float64(bytesPerFrame)
		framesInMemoryRounded := int(math.Floor(framesInMemory))
		if framesInMemoryRounded < len(r.inputFrames) {
			r.hogLevel = 0
			errMsg := fmt.Sprintf("Not enough memory for %d frames, disabling Hog Mode\n", len(r.inputFrames))
			errMsg += fmt.Sprintf("(Needed/Free: %s/%s)\n",
				humanize.Bytes(uint64(bytesPerFrame*len(r.inputFrames))),
				humanize.Bytes(freeMemory))
			return errors.New(errMsg)
		}
		break
	default:
		fmt.Println("Forcing Hog Mode on")
	}

	return nil
}

func (r *Remapper) RemapFrames(inputDir, inputPattern, outputDir string, startFrame int) error {
	r.loadFilesWithPattern(inputDir, inputPattern)

	fmt.Println("Preflight check…")
	err := r.preflightCheck()
	if err != nil {
		return err
	}

	fmt.Println("Hog check…")
	err = r.hogCheck()
	if err != nil {
		// This is the one non-critical error here, so we merely print it instead of panicking
		fmt.Println(err.Error())
	}

	inputFrameCache := make([]image.Image, len(r.inputFrames))

	numDigits := len(strconv.Itoa(r.frameWidth))

	// there will be as many new frames as pixels the input frames are wide
	for frameX := startFrame; frameX < r.frameWidth; frameX++ {
		// new frames are as high as the input frames
		// but as wide as the total number of input frames
		img := image.NewRGBA(image.Rect(0, 0, len(r.inputFrames), r.frameHeight))

		fmt.Printf("Building frame %0[3]*[1]d/%0[3]*[2]d…\n", frameX+1, r.frameWidth, numDigits)

		if r.hogLevel > 0 && frameX == 0 {
			fmt.Println("(Subsequent frames will be processed much faster)")
		}

		percent := -1

		// loop through input frames and fill new frames with 1px-wide pixel columns
		for dstX, path := range r.inputFrames {
			var frameRaw image.Image

			if inputFrameCache[dstX] == nil {
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

				if r.hogLevel > 0 {
					inputFrameCache[dstX] = frameRaw
				}
			} else {
				frameRaw = inputFrameCache[dstX]
			}

			srcRect := image.Rect(frameX, 0, frameX+1, r.frameHeight)
			dstPt := image.Pt(dstX, 0)
			rect := srcRect.Sub(srcRect.Min).Add(dstPt)

			draw.Draw(img, rect, frameRaw, srcRect.Min, draw.Src)

			newPercent := int(math.Floor(float64(dstX) / float64(len(r.inputFrames)) * 100))
			if newPercent >= percent+1 {
				percent = newPercent
				if r.verbose {
					fmt.Printf("%03[1]d%% (%0[4]*[2]d/%0[4]*[3]d)\n",
						newPercent,
						dstX, len(r.inputFrames),
						numDigits)
				}
			}
		}

		if r.verbose {
			fmt.Printf("%03[1]d%% (%0[4]*[2]d/%0[4]*[3]d)\n",
				100,
				len(r.inputFrames), len(r.inputFrames),
				numDigits)
		}

		newFileName := fmt.Sprintf("frame%0[2]*[1]d.png", frameX, numDigits)
		newFilePath := filepath.Join(outputDir, newFileName)
		newFrameFile, err := os.Create(newFilePath)
		if err != nil {
			return err
		}

		err = png.Encode(newFrameFile, img)
		if err != nil {
			return err
		}

		_ = newFrameFile.Close()
	}

	return nil
}
