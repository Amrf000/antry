package libs

import (
	"bytes"
	"encoding/binary"
	"errors"
	"image"
	"image/color"
	"io"
	"math"
	"os"
)

func init() {
	image.RegisterFormat("dds", "DDS ", Decode, DecodeConfig)
}

type DDSimg struct {
	h   DDSHeader
	Buf []DDSByte

	rBit, gBit, bBit, aBit uint

	stride, pitch int
}

func DecodeConfig(r io.Reader) (image.Config, error) {
	return image.Config{}, nil
}

func (i *DDSimg) ColorModel() color.Model {
	return color.NRGBAModel
}

func (i *DDSimg) Bounds() image.Rectangle {
	return image.Rect(0, 0, int(i.h.Width), int(i.h.Height))
}

func (i *DDSimg) At(x, y int) color.Color {
	imgWidth := MAX(1, int(i.h.Width))
	imgHeight := MAX(1, int(i.h.Height))
	imgDepth := MAX(1, int(i.h.Depth))

	//blocksX := MAX(1, imgWidth/4)
	//blocksY := MAX(1, imgHeight/4)
	bytesPerPixel := i.h.PixelFormat.RGBBitCount / 8

	pxIndex := (imgDepth*imgWidth*imgHeight + y*imgWidth + x) * int(bytesPerPixel)
	r := uint8(i.Buf[pxIndex+0])
	g := uint8(i.Buf[pxIndex+1])
	b := uint8(i.Buf[pxIndex+2])
	a := uint8(i.Buf[pxIndex+3])
	return color.NRGBA{r, g, b, a}
}

func Decode(r io.Reader) (image.Image, error) {
	im, err := ddsLoadFromFileFromReader(r)
	if err != nil {
		return &DDSimg{}, err
	}
	return &DDSimg{h: im.Header,
		Buf: im.Pixels,

		//pitch: int(pitch),
		//stride: int(stride),
		//
		//rBit: lowestSetBit(h.pixelFormat.rBitMask),
		//gBit: lowestSetBit(h.pixelFormat.gBitMask),
		//bBit: lowestSetBit(h.pixelFormat.bBitMask),
		//aBit: lowestSetBit(h.pixelFormat.aBitMask),
	}, nil
}

// DDS constants for dds_header::flags
const (
	DDSD_CAPS        = 0x1
	DDSD_HEIGHT      = 0x2
	DDSD_WIDTH       = 0x4
	DDSD_PITCH       = 0x8
	DDSD_PIXELFORMAT = 0x1000
	DDSD_MIPMAPCOUNT = 0x20000
	DDSD_LINEARSIZE  = 0x80000
	DDSD_DEPTH       = 0x800000
)

// DDS constants for dds_header::caps
const (
	DDSCAPS_COMPLEX = 0x8
	DDSCAPS_MIPMAP  = 0x400000
	DDSCAPS_TEXTURE = 0x1000
)

// DDS constants for dds_header::caps2
const (
	DDSCAPS2_CUBEMAP           = 0x200
	DDSCAPS2_CUBEMAP_POSITIVEX = 0x400
	DDSCAPS2_CUBEMAP_NEGATIVEX = 0x800
	DDSCAPS2_CUBEMAP_POSITIVEY = 0x1000
	DDSCAPS2_CUBEMAP_NEGATIVEY = 0x2000
	DDSCAPS2_CUBEMAP_POSITIVEZ = 0x4000
	DDSCAPS2_CUBEMAP_NEGATIVEZ = 0x8000
	DDSCAPS2_VOLUME            = 0x200000
)

// DDS constants for dds_pixelformat::flags
const (
	DDPF_ALPHAPIXELS = 0x1
	DDPF_ALPHA       = 0x2
	DDPF_FOURCC      = 0x4
	DDPF_RGB         = 0x40
	DDPF_YUV         = 0x200
	DDPF_LUMINANCE   = 0x20000
)

// DDSUint and DDSByte types
type DDSUint uint32
type DDSByte byte

// DDSPixelFormat struct
type DDSPixelFormat struct {
	Size        DDSUint
	Flags       DDSUint
	FourCC      DDSUint
	RGBBitCount DDSUint
	RBitMask    DDSUint
	GBitMask    DDSUint
	BBitMask    DDSUint
	ABitMask    DDSUint
}

// DDSHeader struct
type DDSHeader struct {
	Size              DDSUint
	Flags             DDSUint
	Height            DDSUint
	Width             DDSUint
	PitchOrLinearSize DDSUint
	Depth             DDSUint
	MipmapCount       DDSUint
	Reserved1         [11]DDSUint
	PixelFormat       DDSPixelFormat
	Caps              DDSUint
	Caps2             DDSUint
	Caps3             DDSUint
	Caps4             DDSUint
	Reserved2         DDSUint
}

// DDSHeaderDXT10 struct
type DDSHeaderDXT10 struct {
	DXGIFormat        DDSUint
	ResourceDimension DDSUint
	MiscFlag          DDSUint
	ArraySize         DDSUint
	MiscFlags2        DDSUint
}

// DDSImage struct
type DDSImage struct {
	Header      DDSHeader
	HeaderDXT10 DDSHeaderDXT10
	Pixels      []DDSByte // Always RGBA; currently no support for mipmaps
}

// Constants
const (
	FOURCC_DXT1 = 0x31545844 // Equivalent to "DXT1"
)

func FOURCC(str string) DDSUint {
	return DDSUint((str[3] << 24) | (str[2] << 16) | (str[1] << 8) | str[0])
}
func MAX(x, y int) int {
	if x > y {
		return x
	}
	return y
}
func IMAGE_PITCH(width, blockSize int) int {
	return MAX(1, ((width+3)/4)) * blockSize
}

// Function to calculate left shift
func ddsCalculateLeftShift(rightShift *DDSUint, bitCount DDSUint) DDSUint {
	if bitCount >= 8 {
		*rightShift += bitCount - 8
		return 0
	}
	return 8 - bitCount
}

// Function to parse uncompressed DDS image
func ddsParseUncompressed(image *DDSImage, data []byte, dataLength int) {
	imgWidth := MAX(1, int(image.Header.Width))
	imgHeight := MAX(1, int(image.Header.Height))
	imgDepth := MAX(1, int(image.Header.Depth))

	rRightShift, gRightShift, bRightShift, aRightShift := DDSUint(math.MaxUint32), DDSUint(math.MaxUint32), DDSUint(math.MaxUint32), DDSUint(math.MaxUint32)
	rLeftShift, gLeftShift, bLeftShift, aLeftShift := DDSUint(0), DDSUint(0), DDSUint(0), DDSUint(0)

	for i := 0; i < 32; i++ {
		if (image.Header.PixelFormat.RBitMask>>i)&1 != 0 {
			if rRightShift == math.MaxUint32 {
				rRightShift = DDSUint(i)
			}
			rLeftShift++
		}
		if (image.Header.PixelFormat.GBitMask>>i)&1 != 0 {
			if gRightShift == math.MaxUint32 {
				gRightShift = DDSUint(i)
			}
			gLeftShift++
		}
		if (image.Header.PixelFormat.BBitMask>>i)&1 != 0 {
			if bRightShift == math.MaxUint32 {
				bRightShift = DDSUint(i)
			}
			bLeftShift++
		}
		if (image.Header.PixelFormat.ABitMask>>i)&1 != 0 {
			if aRightShift == math.MaxUint32 {
				aRightShift = DDSUint(i)
			}
			aLeftShift++
		}
	}

	// Avoid undefined behavior
	if rRightShift == math.MaxUint32 {
		rRightShift = 0
	}
	if gRightShift == math.MaxUint32 {
		gRightShift = 0
	}
	if bRightShift == math.MaxUint32 {
		bRightShift = 0
	}
	if aRightShift == math.MaxUint32 {
		aRightShift = 0
	}

	// Fix left/right shift based on the bit count (currently stored in X_left_shift)
	rLeftShift = ddsCalculateLeftShift(&rRightShift, rLeftShift)
	gLeftShift = ddsCalculateLeftShift(&gRightShift, gLeftShift)
	bLeftShift = ddsCalculateLeftShift(&bRightShift, bLeftShift)
	aLeftShift = ddsCalculateLeftShift(&aRightShift, aLeftShift)

	bytesPerPixel := image.Header.PixelFormat.RGBBitCount / 8

	data = data[128:] // skip the header

	// Read the actual data
	for z := 0; z < imgDepth; z++ {
		for x := 0; x < imgWidth; x++ {
			for y := 0; y < imgHeight; y++ {
				pxIndex := (z*imgWidth*imgHeight + y*imgWidth + x) * int(bytesPerPixel)
				dataIndex := (z*imgWidth*imgHeight + (imgHeight-y-1)*imgWidth + x) * 4

				// Get the data into uint
				var px DDSUint
				switch bytesPerPixel {
				case 1:
					px = DDSUint(data[pxIndex])
				case 2:
					px = DDSUint(binary.LittleEndian.Uint16(data[pxIndex:]))
				case 4:
					px = DDSUint(binary.LittleEndian.Uint32(data[pxIndex:]))
				}

				// Decode
				image.Pixels[dataIndex+0] = DDSByte(((px & image.Header.PixelFormat.RBitMask) >> rRightShift) << rLeftShift)
				image.Pixels[dataIndex+1] = DDSByte(((px & image.Header.PixelFormat.GBitMask) >> gRightShift) << gLeftShift)
				image.Pixels[dataIndex+2] = DDSByte(((px & image.Header.PixelFormat.BBitMask) >> bRightShift) << bLeftShift)
				if image.Header.PixelFormat.ABitMask == 0 {
					image.Pixels[dataIndex+3] = 0xFF
				} else {
					image.Pixels[dataIndex+3] = DDSByte(((px & image.Header.PixelFormat.ABitMask) >> aRightShift) << aLeftShift)
				}
			}
		}
	}
}

// Function to parse DXT1 compressed DDS image
func ddsParseDXT1(image *DDSImage, data []byte, dataLength int) {
	imgWidth := MAX(1, int(image.Header.Width))
	imgHeight := MAX(1, int(image.Header.Height))
	imgDepth := MAX(1, int(image.Header.Depth))

	blocksX := MAX(1, imgWidth/4)
	blocksY := MAX(1, imgHeight/4)

	for z := 0; z < imgDepth; z++ {
		for x := 0; x < blocksX; x++ {
			for y := 0; y < blocksY; y++ {
				var color0, color1 uint16
				var codes uint32

				// Read the block data
				blockOffset := (y*blocksX + x) * 8
				color0 = binary.LittleEndian.Uint16(data[blockOffset:])
				color1 = binary.LittleEndian.Uint16(data[blockOffset+2:])
				codes = binary.LittleEndian.Uint32(data[blockOffset+4:])

				// Unpack the color data
				r0 := DDSByte((color0 & 0b1111100000000000) >> 8)
				g0 := DDSByte((color0 & 0b0000011111100000) >> 3)
				b0 := DDSByte((color0 & 0b0000000000011111) << 3)
				r1 := DDSByte((color1 & 0b1111100000000000) >> 8)
				g1 := DDSByte((color1 & 0b0000011111100000) >> 3)
				b1 := DDSByte((color1 & 0b0000000000011111) << 3)

				// Process the data
				for b := 0; b < 16; b++ {
					pxIndex := ((z * 4) * imgHeight * imgWidth) + (imgHeight-((y*4)+b/4)-1)*imgWidth + (x*4+b%4)*4

					code := (codes >> (2 * b)) & 3
					image.Pixels[pxIndex+3] = 0xFF
					switch code {
					case 0:
						// color0
						image.Pixels[pxIndex+0] = r0
						image.Pixels[pxIndex+1] = g0
						image.Pixels[pxIndex+2] = b0
					case 1:
						// color1
						image.Pixels[pxIndex+0] = r1
						image.Pixels[pxIndex+1] = g1
						image.Pixels[pxIndex+2] = b1
					case 2:
						if color0 > color1 {
							// (2*color0 + color1) / 3
							image.Pixels[pxIndex+0] = (2*r0 + r1) / 3
							image.Pixels[pxIndex+1] = (2*g0 + g1) / 3
							image.Pixels[pxIndex+2] = (2*b0 + b1) / 3
						} else {
							// (color0 + color1) / 2
							image.Pixels[pxIndex+0] = (r0 + r1) / 2
							image.Pixels[pxIndex+1] = (g0 + g1) / 2
							image.Pixels[pxIndex+2] = (b0 + b1) / 2
						}
					case 3:
						if color0 > color1 {
							// (color0 + 2*color1) / 3
							image.Pixels[pxIndex+0] = (r0 + 2*r1) / 3
							image.Pixels[pxIndex+1] = (g0 + 2*g1) / 3
							image.Pixels[pxIndex+2] = (b0 + 2*b1) / 3
						} else {
							// black
							image.Pixels[pxIndex+0] = 0x00
							image.Pixels[pxIndex+1] = 0x00
							image.Pixels[pxIndex+2] = 0x00
						}
					}
				}
			}
		}

		// Skip this slice
		data = data[blocksX*blocksY*16:]
	}
}

// Function to parse DXT3 compressed DDS image
func ddsParseDXT3(image *DDSImage, data []byte, dataLength int) {
	imgWidth := MAX(1, int(image.Header.Width))
	imgHeight := MAX(1, int(image.Header.Height))
	imgDepth := MAX(1, int(image.Header.Depth))

	blocksX := MAX(1, imgWidth/4)
	blocksY := MAX(1, imgHeight/4)

	for z := 0; z < imgDepth; z++ {
		for x := 0; x < blocksX; x++ {
			for y := 0; y < blocksY; y++ {
				var color0, color1 uint16
				var codes uint32
				var alphaData uint64

				// Read the block data
				blockOffset := (y*blocksX + x) * 16
				alphaData = binary.LittleEndian.Uint64(data[blockOffset:])
				color0 = binary.LittleEndian.Uint16(data[blockOffset+8:])
				color1 = binary.LittleEndian.Uint16(data[blockOffset+10:])
				codes = binary.LittleEndian.Uint32(data[blockOffset+12:])

				// Unpack the color data
				r0 := DDSByte((color0 & 0b1111100000000000) >> 8)
				g0 := DDSByte((color0 & 0b0000011111100000) >> 3)
				b0 := DDSByte((color0 & 0b0000000000011111) << 3)
				r1 := DDSByte((color1 & 0b1111100000000000) >> 8)
				g1 := DDSByte((color1 & 0b0000011111100000) >> 3)
				b1 := DDSByte((color1 & 0b0000000000011111) << 3)

				// Process the data
				for b := 0; b < 16; b++ {
					pxIndex := ((z * 4) * imgHeight * imgWidth) + (imgHeight-((y*4)+b/4)-1)*imgWidth + (x*4+b%4)*4
					code := (codes >> (2 * b)) & 0b0011

					alpha := DDSByte((alphaData >> (4 * b)) & 0b1111)
					image.Pixels[pxIndex+3] = alpha

					switch code {
					case 0:
						// color0
						image.Pixels[pxIndex+0] = r0
						image.Pixels[pxIndex+1] = g0
						image.Pixels[pxIndex+2] = b0
					case 1:
						// color1
						image.Pixels[pxIndex+0] = r1
						image.Pixels[pxIndex+1] = g1
						image.Pixels[pxIndex+2] = b1
					case 2:
						// (2*color0 + color1) / 3
						image.Pixels[pxIndex+0] = (2*r0 + r1) / 3
						image.Pixels[pxIndex+1] = (2*g0 + g1) / 3
						image.Pixels[pxIndex+2] = (2*b0 + b1) / 3
					case 3:
						// (color0 + 2*color1) / 3
						image.Pixels[pxIndex+0] = (r0 + 2*r1) / 3
						image.Pixels[pxIndex+1] = (g0 + 2*g1) / 3
						image.Pixels[pxIndex+2] = (b0 + 2*b1) / 3
					}
				}
			}
		}

		// Skip this slice
		data = data[blocksX*blocksY*16:]
	}
}

// Function to parse DXT5 compressed DDS image
func ddsParseDXT5(image *DDSImage, data []byte, dataLength int) {
	imgWidth := MAX(1, int(image.Header.Width))
	imgHeight := MAX(1, int(image.Header.Height))
	imgDepth := MAX(1, int(image.Header.Depth))

	blocksX := MAX(1, imgWidth/4)
	blocksY := MAX(1, imgHeight/4)

	for z := 0; z < imgDepth; z++ {
		for x := 0; x < blocksX; x++ {
			for y := 0; y < blocksY; y++ {
				var color0, color1 uint16
				var codes uint32
				var alphaCodes uint64
				var alpha0, alpha1 DDSByte

				// Read the block data
				blockOffset := (y*blocksX + x) * 16
				alpha0 = DDSByte(data[blockOffset])
				alpha1 = DDSByte(data[blockOffset+1])
				alphaCodes = binary.LittleEndian.Uint64(data[blockOffset+2:])
				color0 = binary.LittleEndian.Uint16(data[blockOffset+8:])
				color1 = binary.LittleEndian.Uint16(data[blockOffset+10:])
				codes = binary.LittleEndian.Uint32(data[blockOffset+12:])

				// Unpack the color data
				r0 := DDSByte((color0 & 0b1111100000000000) >> 8)
				g0 := DDSByte((color0 & 0b0000011111100000) >> 3)
				b0 := DDSByte((color0 & 0b0000000000011111) << 3)
				r1 := DDSByte((color1 & 0b1111100000000000) >> 8)
				g1 := DDSByte((color1 & 0b0000011111100000) >> 3)
				b1 := DDSByte((color1 & 0b0000000000011111) << 3)

				// Process the data
				for b := 0; b < 16; b++ {
					pxIndex := (z*imgHeight*imgWidth + (imgHeight-((y*4)+b/4)-1)*imgWidth + x*4 + b%4) * 4
					code := (codes >> (2 * b)) & 0b0011
					alphaCode := (alphaCodes >> (3 * b)) & 0b0111

					// Color
					switch code {
					case 0:
						image.Pixels[pxIndex+0] = r0
						image.Pixels[pxIndex+1] = g0
						image.Pixels[pxIndex+2] = b0
					case 1:
						image.Pixels[pxIndex+0] = r1
						image.Pixels[pxIndex+1] = g1
						image.Pixels[pxIndex+2] = b1
					case 2:
						image.Pixels[pxIndex+0] = (2*r0 + r1) / 3
						image.Pixels[pxIndex+1] = (2*g0 + g1) / 3
						image.Pixels[pxIndex+2] = (2*b0 + b1) / 3
					case 3:
						image.Pixels[pxIndex+0] = (r0 + 2*r1) / 3
						image.Pixels[pxIndex+1] = (g0 + 2*g1) / 3
						image.Pixels[pxIndex+2] = (b0 + 2*b1) / 3
					}

					// Alpha
					alpha := DDSByte(0xFF)
					switch {
					case alphaCode == 0:
						alpha = alpha0
					case alphaCode == 1:
						alpha = alpha1
					default:
						if alpha0 > alpha1 {
							alpha = DDSByte(((8-int(alphaCode))*int(alpha0) + (int(alphaCode)-1)*int(alpha1)) / 7)
						} else {
							switch alphaCode {
							case 6:
								alpha = 0
							case 7:
								alpha = 255
							default:
								alpha = DDSByte(((6-int(alphaCode))*int(alpha0) + (int(alphaCode)-1)*int(alpha1)) / 5)
							}
						}
					}
					image.Pixels[pxIndex+3] = alpha
				}
			}
		}

		// Skip this slice
		data = data[blocksX*blocksY*16:]
	}
}

// Constants
const DDSMagic = 0x20534444 // 'DDS '
// Load DDS image from memory
func ddsLoadFromMemory(data []byte) (*DDSImage, error) {
	dataReader := bytes.NewReader(data)

	var magic uint32
	if err := binary.Read(dataReader, binary.LittleEndian, &magic); err != nil {
		return nil, err
	}

	if magic != DDSMagic {
		return nil, errors.New("not a DDS file")
	}

	ret := &DDSImage{}

	// Read the header
	if err := binary.Read(dataReader, binary.LittleEndian, &ret.Header); err != nil {
		return nil, err
	}

	// Check if the header size is 124
	if ret.Header.Size != 124 {
		return nil, errors.New("invalid DDS header size")
	}

	// Check required flags
	if !((ret.Header.Flags&DDSD_CAPS) != 0 && (ret.Header.Flags&DDSD_HEIGHT) != 0 && (ret.Header.Flags&DDSD_WIDTH) != 0 && (ret.Header.Flags&DDSD_PIXELFORMAT) != 0) {
		return nil, errors.New("required DDS header flags are not set")
	}

	// Check if texture is valid
	if (ret.Header.Caps & DDSCAPS_TEXTURE) == 0 {
		return nil, errors.New("invalid DDS texture")
	}

	// Check if we need to load DDS_HEADER_DXT10
	if (ret.Header.PixelFormat.Flags&DDPF_FOURCC) != 0 && ret.Header.PixelFormat.FourCC == FOURCC("DX10") {
		if err := binary.Read(dataReader, binary.LittleEndian, &ret.HeaderDXT10); err != nil {
			return nil, err
		}
	}

	// Allocate pixel data
	imgWidth := MAX(1, int(ret.Header.Width))
	imgHeight := MAX(1, int(ret.Header.Height))
	imgDepth := MAX(1, int(ret.Header.Depth))
	ret.Pixels = make([]DDSByte, imgWidth*imgHeight*imgDepth*4)

	// Parse/decompress the pixel data
	dataLoc, err := io.ReadAll(dataReader)
	if err != nil {
		return nil, err
	}
	dataLength := len(dataLoc)
	if (ret.Header.PixelFormat.Flags & DDPF_FOURCC) != 0 {
		switch ret.Header.PixelFormat.FourCC {
		case FOURCC("DXT1"):
			ddsParseDXT1(ret, dataLoc, dataLength)
		case FOURCC("DXT2"):
			ddsParseDXT3(ret, dataLoc, dataLength)
		case FOURCC("DXT3"):
			ddsParseDXT3(ret, dataLoc, dataLength)
		case FOURCC("DXT4"):
			ddsParseDXT5(ret, dataLoc, dataLength)
		case FOURCC("DXT5"):
			ddsParseDXT5(ret, dataLoc, dataLength)
		case FOURCC("DX10"):
			// ddsParseDXT10
		case FOURCC("ATI1"):
			// ddsParseATI1
		case FOURCC("ATI2"):
			// ddsParseATI2
		case FOURCC("A2XY"):
			// ddsParseA2XY
		default:
			ddsParseUncompressed(ret, dataLoc, dataLength)
		}
	}

	return ret, nil
}

func ddsLoadFromFileFromReader(file io.Reader) (*DDSImage, error) {
	// Read the whole file
	//ddsData := make([]byte, fileSize)
	//_, err := file.Read(ddsData)
	//if err != nil && err != io.EOF {
	//	return nil, err
	//}
	ddsData, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	// Parse the data
	ddsImage, err := ddsLoadFromMemory(ddsData)
	if err != nil {
		return nil, err
	}

	return ddsImage, nil
}

// Load DDS image from file
func ddsLoadFromFile(filename string) (*DDSImage, error) {
	// Open the file
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	// Get file size
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}
	fileSize := fileInfo.Size()

	if fileSize < 128 {
		return nil, errors.New("file size is too small to be a valid DDS file")
	}
	return ddsLoadFromFileFromReader(file)
}

// Free DDS image
func ddsImageFree(image *DDSImage) {
	image.Pixels = nil
}
