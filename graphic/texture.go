package graphic

import _ "image/jpeg"
import (
	"antry/libs"
	_ "antry/libs"
	"fmt"
	_ "github.com/ftrvxmtrx/tga"
	"github.com/go-gl/gl/v4.1-core/gl"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"unsafe"
)

// _ "github.com/lukegb/dds"
type ETextureFiltering int

const (
	TEXTURE_FILTER_MAG_NEAREST         ETextureFiltering = iota // Nearest criterion for magnification
	TEXTURE_FILTER_MAG_BILINEAR                                 // Bilinear criterion for magnification
	TEXTURE_FILTER_MIN_NEAREST                                  // Nearest criterion for minification
	TEXTURE_FILTER_MIN_BILINEAR                                 // Bilinear criterion for minification
	TEXTURE_FILTER_MIN_NEAREST_MIPMAP                           // Nearest criterion for minification, but on closest mipmap
	TEXTURE_FILTER_MIN_BILINEAR_MIPMAP                          // Bilinear criterion for minification, but on closest mipmap
	TEXTURE_FILTER_MIN_TRILINEAR                                // Bilinear criterion for minification on two closest mipmaps, then averaged
)

const NUMTEXTURES = 5

type CTexture struct {
	iWidth            int32
	iHeight           int32
	iBPP              int32  // Texture width, height, and bytes per pixel
	uiTexture         uint32 // Texture name
	uiSampler         uint32 // Sampler name
	bMipMapsGenerated bool

	tfMinification  ETextureFiltering
	tfMagnification ETextureFiltering

	sPath string
}

func NewCTexture() *CTexture {
	this := CTexture{}
	this.bMipMapsGenerated = false
	return &this
}

func (this *CTexture) CreateEmptyTexture(a_iWidth, a_iHeight int32, format uint32) {
	gl.GenTextures(1, &this.uiTexture)
	gl.BindTexture(gl.TEXTURE_2D, this.uiTexture)
	if format == gl.RGBA || format == gl.BGRA {
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, a_iWidth, a_iHeight, 0, format, gl.UNSIGNED_BYTE, nil)
		// We must handle this because of internal format parameter
	} else if format == gl.RGB || format == gl.BGR {
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGB, a_iWidth, a_iHeight, 0, format, gl.UNSIGNED_BYTE, nil)
	} else {
		gl.TexImage2D(gl.TEXTURE_2D, 0, int32(format), a_iWidth, a_iHeight, 0, format, gl.UNSIGNED_BYTE, nil)
	}
	gl.GenSamplers(1, &this.uiSampler)
}

func (this *CTexture) CreateFromData(bData unsafe.Pointer, a_iWidth, a_iHeight, a_iBPP int32, format uint32, bGenerateMipMaps bool) {
	// Generate an OpenGL texture ID for this texture
	gl.GenTextures(1, &this.uiTexture)
	gl.BindTexture(gl.TEXTURE_2D, this.uiTexture)
	if format == gl.RGBA || format == gl.BGRA {
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, a_iWidth, a_iHeight, 0, format, gl.UNSIGNED_BYTE, bData)
		// We must handle this because of internal format parameter
	} else if format == gl.RGB || format == gl.BGR {
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGB, a_iWidth, a_iHeight, 0, format, gl.UNSIGNED_BYTE, bData)
	} else {
		gl.TexImage2D(gl.TEXTURE_2D, 0, int32(format), a_iWidth, a_iHeight, 0, format, gl.UNSIGNED_BYTE, bData)
	}
	if bGenerateMipMaps {
		gl.GenerateMipmap(gl.TEXTURE_2D)
	}
	gl.GenSamplers(1, &this.uiSampler)

	this.sPath = ""
	this.bMipMapsGenerated = bGenerateMipMaps
	this.iWidth = a_iWidth
	this.iHeight = a_iHeight
	this.iBPP = a_iBPP
}

func (this *CTexture) LoadTexture2D(a_sPath string, bGenerateMipMaps bool) bool {
	file, err := os.Open(a_sPath)
	if err != nil {
		fmt.Println("无法打开图片文件:", err)
		panic(err)
		return false
	}
	defer file.Close()

	img, fif, err := image.Decode(file)
	if err != nil {
		fmt.Println("无法打开图片文件:", err)
		panic(err)
		return false
	}
	if fif == "" { // If still unknown, try to guess the file format from the file extension
		fif = filepath.Base(this.sPath)
	}

	if fif == "" {
		panic("unknown fif")
		return false
	}

	// 获取图片宽度和高度
	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	// If somehow one of these failed (they shouldn't), return failure
	if width == 0 || height == 0 {
		panic("unknown width")
		return false
	}
	// 确定位深度
	var format uint32
	var pixels []uint8
	var bpp int32
	// 根据不同的ColorModel获取对应的Pix
	switch img.ColorModel() {
	case color.RGBAModel:
		rgba := img.(*image.RGBA)
		pixels = rgba.Pix
		format = gl.RGBA
		bpp = 32
	case color.NRGBAModel:
		switch v := img.(type) {
		case *image.NRGBA:
			nrgba := v
			pixels = nrgba.Pix
		case *libs.DDSimg:
			nrgba := v
			pixels = *(*[]uint8)(unsafe.Pointer(&nrgba.Buf))
		}

		format = gl.RGBA
		bpp = 32
	case color.GrayModel:
		gray := img.(*image.Gray)
		pixels = gray.Pix
		format = gl.RED
		bpp = 8
	case color.YCbCrModel:
		yimg := img.(*image.YCbCr)
		rgbaImg := image.NewRGBA(yimg.Bounds())
		for y := 0; y < yimg.Bounds().Dy(); y++ {
			for x := 0; x < yimg.Bounds().Dx(); x++ {
				rgbaImg.Set(x, y, yimg.At(x, y))
			}
		}
		pixels = rgbaImg.Pix
		format = gl.RGBA
		bpp = 32
	default:
		fmt.Println("不支持的颜色模型")
		panic("不支持的颜色模型")
	}

	this.CreateFromData(gl.Ptr(pixels), int32(width), int32(height), bpp, format, bGenerateMipMaps)

	this.sPath = a_sPath

	return true // Success
}

func (this *CTexture) SetSamplerParameter(parameter uint32, value int32) {
	gl.SamplerParameteri(this.uiSampler, parameter, value)
}

func (this *CTexture) SetFiltering(a_tfMagnification, a_tfMinification ETextureFiltering) {
	gl.BindSampler(0, this.uiSampler)

	// Set magnification filter
	if a_tfMagnification == TEXTURE_FILTER_MAG_NEAREST {
		gl.SamplerParameteri(this.uiSampler, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	} else if a_tfMagnification == TEXTURE_FILTER_MAG_BILINEAR {
		gl.SamplerParameteri(this.uiSampler, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	}
	// Set minification filter
	if a_tfMinification == TEXTURE_FILTER_MIN_NEAREST {
		gl.SamplerParameteri(this.uiSampler, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	} else if a_tfMinification == TEXTURE_FILTER_MIN_BILINEAR {
		gl.SamplerParameteri(this.uiSampler, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	} else if a_tfMinification == TEXTURE_FILTER_MIN_NEAREST_MIPMAP {
		gl.SamplerParameteri(this.uiSampler, gl.TEXTURE_MIN_FILTER, gl.NEAREST_MIPMAP_NEAREST)
	} else if a_tfMinification == TEXTURE_FILTER_MIN_BILINEAR_MIPMAP {
		gl.SamplerParameteri(this.uiSampler, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_NEAREST)
	} else if a_tfMinification == TEXTURE_FILTER_MIN_TRILINEAR {
		gl.SamplerParameteri(this.uiSampler, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_LINEAR)
	}
	this.tfMinification = a_tfMinification
	this.tfMagnification = a_tfMagnification
}

func (this *CTexture) BindTexture(iTextureUnit uint32) {
	gl.ActiveTexture(gl.TEXTURE0 + iTextureUnit)
	gl.BindTexture(gl.TEXTURE_2D, this.uiTexture)
	gl.BindSampler(iTextureUnit, this.uiSampler)
}

func (this *CTexture) DeleteTexture() {
	gl.DeleteSamplers(1, &this.uiSampler)
	gl.DeleteTextures(1, &this.uiTexture)
}

func (this *CTexture) GetMinificationFilter() ETextureFiltering {
	return this.tfMinification
}

func (this *CTexture) GetMagnificationFilter() ETextureFiltering {
	return this.tfMagnification
}

func (this *CTexture) GetWidth() int32 {
	return this.iWidth
}

func (this *CTexture) GetHeight() int32 {
	return this.iHeight
}

func (this *CTexture) GetBPP() int32 {
	return this.iBPP
}

func (this *CTexture) GetTextureID() uint32 {
	return this.uiTexture
}

func (this *CTexture) GetPath() string {
	return this.sPath
}

func (this *CTexture) ReloadTexture() bool {
	file, err := os.Open(this.sPath)
	if err != nil {
		fmt.Println("无法打开图片文件:", err)
		panic(err)
		return false
	}
	defer file.Close()

	img, fif, err := image.Decode(file)
	if err != nil {
		fmt.Println("无法打开图片文件:", err)
		panic(err)
		return false
	}
	if fif == "" { // If still unknown, try to guess the file format from the file extension
		fif = filepath.Base(this.sPath)
	}

	if fif == "" {
		panic("fif == \"\"")
		return false
	}

	// 获取图片宽度和高度
	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	// If somehow one of these failed (they shouldn't), return failure
	if width == 0 || height == 0 {
		panic("width == 0")
		return false
	}
	// 确定位深度
	var format uint32
	var pixels []uint8
	// 根据不同的ColorModel获取对应的Pix
	switch img.ColorModel() {
	case color.RGBAModel:
		rgba := img.(*image.RGBA)
		pixels = rgba.Pix
		format = gl.RGBA
	case color.NRGBAModel:
		nrgba := img.(*image.NRGBA)
		pixels = nrgba.Pix
		format = gl.RGBA
	case color.GrayModel:
		gray := img.(*image.Gray)
		pixels = gray.Pix
		format = gl.RED
	default:
		fmt.Println("不支持的颜色模型")
		panic("不支持的颜色模型")
	}

	gl.BindTexture(gl.TEXTURE_2D, this.uiTexture)
	gl.TexSubImage2D(gl.TEXTURE_2D, 0, 0, 0, this.iWidth, this.iHeight, format, gl.UNSIGNED_BYTE, gl.Ptr(pixels))
	if this.bMipMapsGenerated {
		gl.GenerateMipmap(gl.TEXTURE_2D)
	}

	return true // Success
}

var tTextures [NUMTEXTURES]CTexture

func LoadAllTextures() {
	// Load textures

	var sTextureNames = []string{"fungus.dds", "sand_grass_02.jpg", "rock_2_4w.jpg", "sand.jpg", "path.png"}

	for i := 0; i < NUMTEXTURES; i++ {
		flag := tTextures[i].LoadTexture2D("data\\textures\\"+sTextureNames[i], true)
		if !flag {
			panic("LoadTexture2D failed")
		}
		tTextures[i].SetFiltering(TEXTURE_FILTER_MAG_BILINEAR, TEXTURE_FILTER_MIN_BILINEAR_MIPMAP)
	}
}
