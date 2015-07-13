// jigsaw package reads in template images and creates jigsaw puzzle pieces using the pixels from a main image

package Jigsaw

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
)

const PIECE_PREFIX = "Piece"
const PIECE_DIRECTORY = "out"

//holds relevant info about a piece
type Piece struct {
	FileLocation   string `json:"fileLocation"`
	PieceLocationX int    `json:pieceLocationX`
	PieceLocationY int    `json:pieceLocationY`
	TemplateImage  image.Image
	PieceImage     image.RGBA
}

//the main jigsaw struct. holds config info as well as list of pieces
type Jigsaw struct {
	FullImageLocation  string  `json:"fullImageLocation"`
	TemplateLocation   string  `json:"templateLocation"`
	Pieces             []Piece `json:"pieces"`
	PieceWidth         int     `json:"pieceWidth"`
	PieceHeight        int     `json:"pieceHeight"`
	PieceOverflow      int     `json:"pieceOverflow"`
	PieceColumns       int     `json:"pieceColumns"`
	PieceRows          int     `json:"pieceRows"`
	TemplateOff        int     `json:"templateOff"`
	TemplateOn         int     `json:"templateOn"`
	OffColor           color.RGBA
	OnColor            color.RGBA
	FullImage          image.Image
	ImageRootDirectory string
}

//just a simple struct for offsets
type PieceOffsets struct {
	offsetTop    int
	offsetBottom int
	offsetRight  int
	offsetLeft   int
}

//initializes the main image, creates the on and off colors
func (jigsaw *Jigsaw) Init(imageRootDirectory string) bool {
	jigsaw.ImageRootDirectory = imageRootDirectory
	os.Mkdir(imageRootDirectory+string(filepath.Separator)+PIECE_DIRECTORY, 0777)
	image.RegisterFormat("png", "png", png.Decode, png.DecodeConfig)
	fullImage, err := createImage(jigsaw.ImageRootDirectory + "/" + jigsaw.FullImageLocation)
	jigsaw.OffColor = color.RGBA{uint8(jigsaw.TemplateOff >> 24), uint8(jigsaw.TemplateOff >> 16), uint8(jigsaw.TemplateOff >> 8), uint8(jigsaw.TemplateOff & 0x000000FF)}
	jigsaw.OnColor = color.RGBA{uint8(jigsaw.TemplateOn >> 24), uint8(jigsaw.TemplateOn >> 16), uint8(jigsaw.TemplateOn >> 8), uint8(jigsaw.TemplateOn & 0x000000FF)}
	if err == nil {
		jigsaw.FullImage = fullImage
		return true
	}
	return false
}

//initialize the pieces.
//ie. we need to load the template image
//we need to create an empty image that's the matching size as template.
//then communicate on the channel when a piece is ready
func (jigsaw *Jigsaw) InitPieces(readyImages chan<- int) {
	fmt.Println("starting")
	for i := range jigsaw.Pieces {
		fmt.Println("Adding piece ", i)
		templateImage, err := createImage(jigsaw.ImageRootDirectory + "/" + jigsaw.TemplateLocation + "/" + jigsaw.Pieces[i].FileLocation)
		if err != nil {
			fmt.Println("Could not create template piece", i)
			close(readyImages)
			return
		}
		jigsaw.Pieces[i].TemplateImage = templateImage
		jigsaw.Pieces[i].PieceImage = *image.NewRGBA(templateImage.Bounds())
		readyImages <- i
	}
	close(readyImages)
}

//actually do the copying of pixels. When a piece is ready, we copy all pixels from the main image
//into our piece image using the template as a guide
func (jigsaw *Jigsaw) CreateImage(readyImages <-chan int, createdPieces chan<- int) {
	for {
		job, more := <-readyImages
		if more {
			fmt.Println("Transferring pixels for piece ", job)
			pieceOffsets := createOffsets(job, jigsaw)
			currentPiece := &jigsaw.Pieces[job]
			pieceRect := (*currentPiece).TemplateImage.Bounds()
			//the starting points on the full image
			var currentColumn int = job % jigsaw.PieceColumns
			var startingX int = currentColumn*jigsaw.PieceWidth - jigsaw.PieceOverflow + pieceOffsets.offsetLeft
			var currentRow int = job / jigsaw.PieceColumns
			var startingY int = currentRow*jigsaw.PieceHeight - jigsaw.PieceOverflow + pieceOffsets.offsetTop

			for i := 0; i < pieceRect.Dy(); i++ {
				for k := 0; k < pieceRect.Dx(); k++ {
					if k > jigsaw.FullImage.Bounds().Dx() || i > jigsaw.FullImage.Bounds().Dy() {
						currentPiece.PieceImage.Set(k, i, jigsaw.OffColor)
						continue
					}
					r, g, b, a := currentPiece.TemplateImage.At(k, i).RGBA()
					offR, offG, offB, offA := jigsaw.OffColor.RGBA()
					if r == offR && g == offG && b == offB && a == offA {
						currentPiece.PieceImage.Set(k, i, jigsaw.OffColor)
					} else {
						fullPieceColor := jigsaw.FullImage.At(k+startingX-pieceOffsets.offsetLeft, i+startingY-pieceOffsets.offsetTop)
						currentPiece.PieceImage.Set(k, i, fullPieceColor)
					}
				}
			}
			createdPieces <- job
		} else {
			fmt.Println("received all jobs")
			close(createdPieces)
			return
		}
	}
}

//simply saves the piece image
func (jigsaw *Jigsaw) SaveImage(createdPieces <-chan int, done chan<- bool) {
	for {
		job, more := <-createdPieces
		if more {
			fmt.Println("finishing piece", job)
			pieceImage := jigsaw.Pieces[job].PieceImage.SubImage(jigsaw.Pieces[job].TemplateImage.Bounds())
			out, err := os.Create(jigsaw.ImageRootDirectory + string(filepath.Separator) + PIECE_DIRECTORY + string(filepath.Separator) + PIECE_PREFIX + strconv.Itoa(job) + ".png")
			defer out.Close()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			err = png.Encode(out, pieceImage)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		} else {
			fmt.Println("received all jobs")
			done <- true
			return
		}
	}
}

//private functions

//create image is a general purpose function that loads an image from a given location on
//disk and returns that image
func createImage(fileLocation string) (image.Image, error) {
	var img image.Image
	imgfile, err := os.Open(fileLocation)
	defer imgfile.Close()
	if err != nil {
		fmt.Println("image file not found")
		return img, err
	}
	defer imgfile.Close()
	img, _, err = image.Decode(imgfile)
	if err != nil {
		fmt.Println("could not decode image")
		return img, err
	}
	return img, nil
}

//creates the offsets given the current piece location on overall grid
func createOffsets(pieceLocation int, jigsaw *Jigsaw) *PieceOffsets {
	pieceOffsets := new(PieceOffsets)
	if pieceLocation%jigsaw.PieceColumns == 0 {
		//we're on the left edge
		pieceOffsets.offsetLeft = jigsaw.PieceOverflow
	} else {
		pieceOffsets.offsetLeft = 0
	}
	if (pieceLocation)%jigsaw.PieceColumns == jigsaw.PieceColumns-1 {
		//we're on the right edge
		pieceOffsets.offsetRight = jigsaw.PieceOverflow
	} else {
		pieceOffsets.offsetRight = jigsaw.PieceOverflow
	}
	if pieceLocation < jigsaw.PieceColumns {
		//we're at the top edge
		pieceOffsets.offsetTop = jigsaw.PieceOverflow
	} else {
		pieceOffsets.offsetTop = 0
	}
	if pieceLocation >= len(jigsaw.Pieces)-jigsaw.PieceColumns {
		//we're on the bottom edge
		pieceOffsets.offsetBottom = jigsaw.PieceOverflow
	} else {
		pieceOffsets.offsetBottom = 0
	}
	return pieceOffsets
}
