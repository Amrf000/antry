package graphic

import (
	"fmt"
	"github.com/flopp/go-findfont"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"log"
	"os"
	"unsafe"
)

type CFreeTypeFont struct {
	tCharTextures    [256]CTexture
	iAdvX            [256]int32
	iAdvY            [256]int
	iBearingX        [256]int32
	iBearingY        [256]int
	iCharWidth       [256]int
	iCharHeight      [256]int
	iLoadedPixelSize float64
	iNewLine         int32

	bLoaded bool

	uiVAO   uint32
	vboData *CVertexBufferObject

	ftFace          font.Face
	shShaderProgram *CShaderProgram
}

func NewCFreeTypeFont() *CFreeTypeFont {
	this := CFreeTypeFont{}
	this.bLoaded = false
	return &this
}

func (this *CFreeTypeFont) next_p2(n int) int {
	res := 1
	for res < n {
		res <<= 1
	}
	return res
}

func (this *CFreeTypeFont) CreateChar(iIndex int) {
	//FT_Load_Glyph(this.ftFace, FT_Get_Char_Index(this.ftFace, iIndex), FT_LOAD_DEFAULT)
	//
	//FT_Render_Glyph(this.ftFace.glyph, FT_RENDER_MODE_NORMAL)
	//var pBitmap *FT_Bitmap = &this.ftFace.glyph.bitmap
	//
	//var iW int = pBitmap.width
	//var iH int = pBitmap.rows
	//var iTW int = this.next_p2(iW)
	//var iTH int = this.next_p2(iH)
	//
	//bData := make([]byte, iTW*iTH)
	//// Copy glyph data and add dark pixels elsewhere
	//for ch := 0; ch < iTH; ch++ {
	//	for cw := 0; cw < iTW; cw++ {
	//		if ch >= iH || cw >= iW {
	//			bData[ch*iTW+cw] = 0
	//		} else {
	//			bData[ch*iTW+cw] = pBitmap.buffer[(iH-ch-1)*iW+cw]
	//		}
	//	}
	//}

	// And create a texture from it

	//this.tCharTextures[iIndex].CreateFromData(unsafe.Pointer(&bData[0]), int32(iTW), int32(iTH), 8, gl.DEPTH_COMPONENT, false)
	//this.tCharTextures[iIndex].SetFiltering(TEXTURE_FILTER_MAG_BILINEAR, TEXTURE_FILTER_MIN_BILINEAR)
	//
	//this.tCharTextures[iIndex].SetSamplerParameter(gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	//this.tCharTextures[iIndex].SetSamplerParameter(gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

	// Calculate glyph data
	//bounds,advance,ok := this.ftFace.GlyphBounds()
	//this.iAdvX[iIndex] = advance. >> 6
	//this.iBearingX[iIndex] = this.ftFace.glyph.metrics.horiBearingX >> 6
	//this.iCharWidth[iIndex] = this.ftFace.glyph.metrics.width >> 6
	//
	//this.ftFace.GlyphAdvance()
	//this.iAdvY[iIndex] = (this.ftFace.glyph.metrics.height - this.ftFace.glyph.metrics.horiBearingY) >> 6
	//this.iBearingY[iIndex] = this.ftFace.glyph.metrics.horiBearingY >> 6
	//this.iCharHeight[iIndex] = this.ftFace.glyph.metrics.height >> 6
	//
	//this.iNewLine = math.Max(this.iNewLine, int32(this.ftFace.glyph.metrics.height>>6))

	// Rendering data, texture coordinates are always the same, so now we waste a little memory
	//vQuad := []mgl32.Vec2{
	//	mgl32.Vec2{0.0, float32(-this.iAdvY[iIndex] + iTH)},
	//	mgl32.Vec2{0.0, float32(-this.iAdvY[iIndex])},
	//	mgl32.Vec2{float32(iTW), float32(-this.iAdvY[iIndex] + iTH)},
	//	mgl32.Vec2{float32(iTW), float32(-this.iAdvY[iIndex])},
	//}
	//vTexQuad := []mgl32.Vec2{mgl32.Vec2{0.0, 1.0}, mgl32.Vec2{0.0, 0.0}, mgl32.Vec2{1.0, 1.0}, mgl32.Vec2{1.0, 0.0}}
	//
	//// Add this char to VBO
	//for i := 0; i < 4; i++ {
	//	this.vboData.AddData(*(*[]byte)(unsafe.Pointer(&vQuad[i])), int32(unsafe.Sizeof(mgl32.Vec2{})))
	//	this.vboData.AddData(*(*[]byte)(unsafe.Pointer(&vTexQuad[i])), int32(unsafe.Sizeof(mgl32.Vec2{})))
	//}
	//bData = nil
}

/*-----------------------------------------------

  Name:	LoadFont

  Params:	sFile - path to font file
  		iPXSize - desired font pixel size

  Result:	Loads whole font.

  /*---------------------------------------------*/

func (this *CFreeTypeFont) LoadFont(sFile string, iPXSize float64) bool {
	fontBytes, err := os.ReadFile(sFile)
	if err != nil {
		log.Println(err)
		return false
	}
	f, err := truetype.Parse(fontBytes)
	this.ftFace = truetype.NewFace(f, &truetype.Options{Size: iPXSize})

	this.iLoadedPixelSize = iPXSize

	gl.GenVertexArrays(1, &uiVAO)
	gl.BindVertexArray(uiVAO)
	this.vboData = NewCVertexBufferObject()
	this.vboData.CreateVBO(0)
	this.vboData.BindVBO(gl.ARRAY_BUFFER)

	//for i := 0; i < 128; i++ {
	//	this.CreateChar(i)
	//}
	this.bLoaded = true

	this.ftFace.Close()

	//this.vboData.UploadDataToGPU(gl.STATIC_DRAW)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, int32(unsafe.Sizeof(mgl32.Vec2{})*2), nil)
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, int32(unsafe.Sizeof(mgl32.Vec2{})*2), unsafe.Pointer(unsafe.Sizeof(mgl32.Vec2{})))
	return true
}

/*-----------------------------------------------

  Name:	LoadSystemFont

  Params:	sName - system font name
  		iPXSize - desired font pixel size

  Result:	Loads system font (from system Fonts
  		directory).

  /*---------------------------------------------*/

func (this *CFreeTypeFont) LoadSystemFont(sName string, iPXSize float64) bool {
	fontPath, err := findfont.Find(sName)
	if err != nil {
		log.Fatalf("无法找到系统字体: %v", err)
	}
	fmt.Printf("使用字体: %s\n", fontPath)

	return this.LoadFont(fontPath, iPXSize)
}

/*-----------------------------------------------

  Name:	GetTextWidth

  Params:	sText - text to get width of
  		iPXSize - it's printed size

  Result:	Returns width as number of pixels the
  		text will occupy.

  /*---------------------------------------------*/

func (this *CFreeTypeFont) GetTextWidth(sText string, iPXSize int32) int32 {
	var iResult int32 = 0
	for i := 0; i < len(sText); i++ {
		iResult += this.iAdvX[sText[i]]
	}

	return int32(float64(iResult*iPXSize) / this.iLoadedPixelSize)
}

/*-----------------------------------------------

  Name:	Print

  Params:	sText - text to print
  		x, y - 2D position
  		iPXSize - printed text size

  Result:	Prints text at specified position
  		with specified pixel size.

  /*---------------------------------------------*/

func (this *CFreeTypeFont) Print(sText string, x, y, iPXSize int32) {
	if !this.bLoaded {
		return
	}

	gl.BindVertexArray(uiVAO)
	this.shShaderProgram.SetUniformI32("gSampler", 0)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	var iCurX int32 = x
	var iCurY int32 = y
	if iPXSize == -1 {
		iPXSize = int32(this.iLoadedPixelSize)
	}
	var fScale float32 = float32(iPXSize) / float32(this.iLoadedPixelSize)
	for i := 0; i < len(sText); i++ {
		if sText[i] == '\n' {
			iCurX = x
			iCurY -= int32(float64(this.iNewLine*iPXSize) / this.iLoadedPixelSize)
			continue
		}

		var iIndex int32 = int32(sText[i])
		iCurX += int32(float64(this.iBearingX[iIndex]*iPXSize) / this.iLoadedPixelSize)
		if sText[i] != ' ' {
			this.tCharTextures[iIndex].BindTexture(0)
		}

		var mModelView mgl32.Mat4 = mgl32.Translate3D(float32(iCurX), float32(iCurY), 0.0)
		mModelView = mModelView.Mul4(mgl32.Scale3D(fScale, fScale, fScale))
		this.shShaderProgram.SetUniformM4("matrices.modelViewMatrix", mModelView)
		// Draw character
		gl.DrawArrays(gl.TRIANGLE_STRIP, iIndex*4, 4)

		iCurX += int32(float64((this.iAdvX[iIndex]-this.iBearingX[iIndex])*iPXSize) / this.iLoadedPixelSize)
	}
	gl.Disable(gl.BLEND)
}

/*-----------------------------------------------

  Name:	PrintFormatted

  Params:	x, y - 2D position
  		iPXSize - printed text size
  		sText - text to print

  Result:	Prints formatted text at specified position
  		with specified pixel size.

  /*---------------------------------------------*/

func (this *CFreeTypeFont) PrintFormatted(x, y, iPXSize int, sText string) {
	//var buf [512]byte
	//va_list ap;
	//va_start(ap, sText);
	//vsprintf(buf, sText, ap);
	//va_end(ap);
	//Print(buf, x, y, iPXSize);
}

/*-----------------------------------------------

  Name:	DeleteFont

  Params:	none

  Result:	Deletes all font textures.

  /*---------------------------------------------*/

func (this *CFreeTypeFont) DeleteFont() {
	for i := 0; i < 128; i++ {
		this.tCharTextures[i].DeleteTexture()
	}
	this.vboData.DeleteVBO()
	gl.DeleteVertexArrays(1, &this.uiVAO)
}

/*-----------------------------------------------

  Name:	SetShaderProgram

  Params:	a_shShaderProgram - shader program

  Result:	Sets shader program that font uses.

  /*---------------------------------------------*/

func (this *CFreeTypeFont) SetShaderProgram(a_shShaderProgram *CShaderProgram) {
	this.shShaderProgram = a_shShaderProgram
}
