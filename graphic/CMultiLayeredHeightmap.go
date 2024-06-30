package graphic

import (
	"fmt"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	_ "golang.org/x/image/bmp"
	"image"
	"image/color"
	_ "image/png"
	"os"
	"unsafe"
)

const NUMTERRAINSHADERS = 3

type CMultiLayeredHeightmap struct {
	uiVAO uint32

	bLoaded              bool
	bShaderProgramLoaded bool
	iRows                int
	iCols                int

	vRenderScale mgl32.Vec3

	vboHeightmapData    *CVertexBufferObject
	vboHeightmapIndices *CVertexBufferObject
}

var spTerrain CShaderProgram
var shTerrainShaders [NUMTERRAINSHADERS]CShader

func NewCMultiLayeredHeightmap() *CMultiLayeredHeightmap {
	this := CMultiLayeredHeightmap{}
	this.vRenderScale = mgl32.Vec3{1.0, 1.0, 1.0}
	return &this
}

// 将颜色转换为灰度值
func colorToGray(c color.Color) float64 {
	r, g, b, _ := c.RGBA()
	// 将16位色彩值转换为8位色彩值
	r8, g8, b8 := float64(r>>8), float64(g>>8), float64(b>>8)
	return 0.299*r8 + 0.587*g8 + 0.114*b8
}
func (this *CMultiLayeredHeightmap) LoadHeightMapFromImage(sImagePath string) bool {
	if this.bLoaded {
		this.bLoaded = false
		this.ReleaseHeightmap()
	}
	f, err := os.Open(sImagePath)
	if err != nil {
		return false
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		fmt.Println("无法解码图片:", err)
		return false
	}

	// 获取图片宽度和高度
	bounds := img.Bounds()
	this.iCols, this.iRows = bounds.Max.X, bounds.Max.Y

	// We also require our image to be either 24-bit (classic RGB) or 8-bit (luminance)
	if this.iRows == 0 || this.iCols == 0 {
		panic("this.iRows == 0")
		return false
	}

	this.vboHeightmapData = NewCVertexBufferObject()
	this.vboHeightmapData.CreateVBO(0)
	// All vertex data are here (there are iRows*iCols vertices in this heightmap), we will get to normals later
	vVertexData := make([][]mgl32.Vec3, this.iRows)
	for i := range vVertexData {
		vVertexData[i] = make([]mgl32.Vec3, this.iCols)
	}
	vCoordsData := make([][]mgl32.Vec2, this.iRows)
	for i := range vCoordsData {
		vCoordsData[i] = make([]mgl32.Vec2, this.iCols)
	}

	var fTextureU float32 = float32(this.iCols) * 0.1
	var fTextureV float32 = float32(this.iRows) * 0.1

	for i := 0; i < this.iRows; i++ {
		for j := 0; j < this.iCols; j++ {
			var fScaleC float32 = float32(j) / float32(this.iCols-1)
			var fScaleR float32 = float32(i) / float32(this.iRows-1)
			var fVertexHeight float32 = 0
			// 获取图片某个位置的像素值
			m := img.ColorModel()
			if _, ok := m.(color.Palette); ok {
				palettedImg, ok := img.(*image.Paletted)
				if ok {
					palette := palettedImg.Palette
					index := palettedImg.ColorIndexAt(i, j)
					// 获取该颜色索引对应的颜色
					col := palette[index]
					// 将颜色转换为灰度值
					grayValue := colorToGray(col)
					fVertexHeight = float32(grayValue / 255.0)
				} else {
					panic("palettedImg invalid")
				}
			} else {
				switch m {
				case color.GrayModel:
					// 灰度图像，获取灰度值
					gray := img.(interface {
						At(i, j int) color.Color
					}).At(i, j).(color.Gray)
					fmt.Printf("灰度值为: %d\n", gray.Y)
					fVertexHeight = float32(gray.Y / 255.0)
				case color.RGBAModel, color.NRGBAModel:
					// RGB图像，获取各个通道的像素值
					rgba := img.(interface {
						At(i, j int) color.Color
					}).At(i, j).(color.RGBA)
					fmt.Printf("R: %d, G: %d, B: %d, A: %d\n", rgba.R, rgba.G, rgba.B, rgba.A)
					gray := uint8(0.2126*float64(rgba.R) + 0.7152*float64(rgba.G) + 0.0722*float64(rgba.B))
					fVertexHeight = float32(gray / 255.0)
				case color.YCbCrModel:
					yimg := img.(*image.YCbCr)
					rgbaImg := image.NewRGBA(yimg.Bounds())
					for y := 0; y < yimg.Bounds().Dy(); y++ {
						for x := 0; x < yimg.Bounds().Dx(); x++ {
							rgbaImg.Set(x, y, yimg.At(x, y))
						}
					}
					rgba := rgbaImg.At(i, j).(color.RGBA)
					fmt.Printf("R: %d, G: %d, B: %d, A: %d\n", rgba.R, rgba.G, rgba.B, rgba.A)
					gray := uint8(0.2126*float64(rgba.R) + 0.7152*float64(rgba.G) + 0.0722*float64(rgba.B))
					fVertexHeight = float32(gray / 255.0)
				default:
					fmt.Println("不支持的图片类型,%v", m)
					panic("不支持的图片类型")
				}
			}

			vVertexData[i][j] = mgl32.Vec3{-0.5 + fScaleC, fVertexHeight, -0.5 + fScaleR}
			vCoordsData[i][j] = mgl32.Vec2{fTextureU * fScaleC, fTextureV * fScaleR}
		}
	}

	// Normals are here - the heightmap contains ( (iRows-1)*(iCols-1) quads, each one containing 2 triangles, therefore array of we have 3D array)
	vNormals := [2][][]mgl32.Vec3{}
	for i := 0; i < 2; i++ {
		arr := make([][]mgl32.Vec3, this.iRows-1)
		for i := range arr {
			arr[i] = make([]mgl32.Vec3, this.iCols-1)
		}
		vNormals[i] = arr
	}

	for i := 0; i < this.iRows-1; i++ {
		for j := 0; j < this.iCols-1; j++ {
			vTriangle0 := []mgl32.Vec3{
				vVertexData[i][j],
				vVertexData[i+1][j],
				vVertexData[i+1][j+1],
			}
			vTriangle1 := []mgl32.Vec3{
				vVertexData[i+1][j+1],
				vVertexData[i][j+1],
				vVertexData[i][j],
			}

			vTriangleNorm0 := vTriangle0[0].Sub(vTriangle0[1]).Cross(vTriangle0[1].Sub(vTriangle0[2]))
			vTriangleNorm1 := vTriangle1[0].Sub(vTriangle1[1]).Cross(vTriangle1[1].Sub(vTriangle1[2]))

			vNormals[0][i][j] = vTriangleNorm0.Normalize()
			vNormals[1][i][j] = vTriangleNorm1.Normalize()
		}
	}

	vFinalNormals := make([][]mgl32.Vec3, this.iRows)
	for i := range vFinalNormals {
		vFinalNormals[i] = make([]mgl32.Vec3, this.iCols)
	}

	for i := 0; i < this.iRows; i++ {
		for j := 0; j < this.iCols; j++ {
			// Now we wanna calculate final normal for [i][j] vertex. We will have a look at all triangles this vertex is part of, and then we will make average vector
			// of all adjacent triangles' normals

			var vFinalNormal = mgl32.Vec3{0.0, 0.0, 0.0}

			// Look for upper-left triangles
			if j != 0 && i != 0 {
				for k := 0; k < 2; k++ {
					vFinalNormal = vFinalNormal.Add(vNormals[k][i-1][j-1])
				}
			}

			// Look for upper-right triangles
			if i != 0 && j != this.iCols-1 {
				vFinalNormal = vFinalNormal.Add(vNormals[0][i-1][j])
			}
			// Look for bottom-right triangles
			if i != this.iRows-1 && j != this.iCols-1 {
				for k := 0; k < 2; k++ {
					vFinalNormal = vFinalNormal.Add(vNormals[k][i][j])
				}
			}
			// Look for bottom-left triangles
			if i != this.iRows-1 && j != 0 {
				vFinalNormal = vFinalNormal.Add(vNormals[1][i][j-1])
			}
			vFinalNormal = vFinalNormal.Normalize()

			vFinalNormals[i][j] = vFinalNormal // Store final normal of j-th vertex in i-th row
		}
	}

	// First, create a VBO with only vertex data
	this.vboHeightmapData.CreateVBO(this.iRows * this.iCols * int((2*unsafe.Sizeof(mgl32.Vec3{}) + unsafe.Sizeof(mgl32.Vec2{})))) // Preallocate memory
	for i := 0; i < this.iRows; i++ {
		for j := 0; j < this.iCols; j++ {
			this.vboHeightmapData.AddData(EncodeToBytes(vVertexData[i][j]), int32(unsafe.Sizeof(mgl32.Vec3{})))   // Add vertex
			this.vboHeightmapData.AddData(EncodeToBytes(vCoordsData[i][j]), int32(unsafe.Sizeof(mgl32.Vec2{})))   // Add tex. coord
			this.vboHeightmapData.AddData(EncodeToBytes(vFinalNormals[i][j]), int32(unsafe.Sizeof(mgl32.Vec3{}))) // Add normal
		}
	}
	// Now create a VBO with heightmap indices
	this.vboHeightmapIndices = NewCVertexBufferObject()
	this.vboHeightmapIndices.CreateVBO(0)
	var iPrimitiveRestartIndex int32 = int32(this.iRows * this.iCols)
	for i := 0; i < this.iRows-1; i++ {
		for j := 0; j < this.iCols; j++ {
			for k := 0; k < 2; k++ {
				var iRow int = i + (1 - k)
				var iIndex int32 = int32(iRow*this.iCols + j)
				this.vboHeightmapIndices.AddData(EncodeToBytes(iIndex), int32(unsafe.Sizeof(int32(0))))
			}
		}
		// Restart triangle strips
		this.vboHeightmapIndices.AddData(EncodeToBytes(iPrimitiveRestartIndex), int32(unsafe.Sizeof(int32(0))))
	}

	gl.GenVertexArrays(1, &uiVAO)
	gl.BindVertexArray(uiVAO)
	// Attach vertex data to this VAO
	this.vboHeightmapData.BindVBO(gl.ARRAY_BUFFER)
	this.vboHeightmapData.UploadDataToGPU(gl.STATIC_DRAW)

	// Vertex positions
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, int32(2*unsafe.Sizeof(mgl32.Vec3{})+unsafe.Sizeof(mgl32.Vec2{})), nil)
	// Texture coordinates
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointerWithOffset(1, 2, gl.FLOAT, false, int32(2*unsafe.Sizeof(mgl32.Vec3{})+unsafe.Sizeof(mgl32.Vec2{})),
		unsafe.Sizeof(mgl32.Vec3{}))
	// Normal vectors
	gl.EnableVertexAttribArray(2)
	gl.VertexAttribPointerWithOffset(2, 3, gl.FLOAT, false, int32(2*unsafe.Sizeof(mgl32.Vec3{})+unsafe.Sizeof(mgl32.Vec2{})),
		unsafe.Sizeof(mgl32.Vec3{})+unsafe.Sizeof(mgl32.Vec2{}))

	// And now attach index data to this VAO
	// Here don't forget to bind another type of VBO - the element array buffer, or simplier indices to vertices
	this.vboHeightmapIndices.BindVBO(gl.ELEMENT_ARRAY_BUFFER)
	this.vboHeightmapIndices.UploadDataToGPU(gl.STATIC_DRAW)

	this.bLoaded = true // If get here, we succeeded with generating heightmap
	return true
}
func LoadTerrainShaderProgram() bool {
	bOK := true
	bOK = bOK && shShaders[0].LoadShader("data\\shaders\\terrain.vert", gl.VERTEX_SHADER)
	bOK = bOK && shShaders[1].LoadShader("data\\shaders\\terrain.frag", gl.FRAGMENT_SHADER)
	bOK = bOK && shShaders[2].LoadShader("data\\shaders\\dirLight.frag", gl.FRAGMENT_SHADER)

	spTerrain.CreateProgram()
	for i := 0; i < NUMTERRAINSHADERS; i++ {
		spTerrain.AddShaderToProgram(&shShaders[i])
	}
	spTerrain.LinkProgram()

	return bOK
}

func (this *CMultiLayeredHeightmap) SetRenderSize3(fRenderX, fHeight, fRenderZ float32) {
	this.vRenderScale = mgl32.Vec3{fRenderX, fHeight, fRenderZ}
}

func (this *CMultiLayeredHeightmap) SetRenderSize(fQuadSize, fHeight float32) {
	this.vRenderScale = mgl32.Vec3{float32(this.iCols) * fQuadSize, fHeight, float32(this.iRows) * fQuadSize}
}
func (this *CMultiLayeredHeightmap) RenderHeightmap() {
	spTerrain.UseProgram()

	spTerrain.SetUniformF32("fRenderHeight", this.vRenderScale.Y())
	spTerrain.SetUniformF32("fMaxTextureU", float32(this.iCols)*float32(0.1))
	spTerrain.SetUniformF32("fMaxTextureV", float32(this.iRows)*float32(0.1))

	spTerrain.SetUniformM4("HeightmapScaleMatrix", mgl32.Scale3D(this.vRenderScale.X(), this.vRenderScale.Y(), this.vRenderScale.Z()))

	// Now we're ready to render - we are drawing set of triangle strips using one call, but we g otta enable primitive restart
	gl.BindVertexArray(uiVAO)
	gl.Enable(gl.PRIMITIVE_RESTART)
	gl.PrimitiveRestartIndex(uint32(this.iRows * this.iCols))

	var iNumIndices int32 = int32((this.iRows-1)*this.iCols*2 + this.iRows - 1)
	gl.DrawElements(gl.TRIANGLE_STRIP, iNumIndices, gl.UNSIGNED_INT, nil)
}
func (this *CMultiLayeredHeightmap) ReleaseHeightmap() {
	if !this.bLoaded {
		return // Heightmap must be loaded
	}
	this.vboHeightmapData.DeleteVBO()
	this.vboHeightmapIndices.DeleteVBO()
	gl.DeleteVertexArrays(1, &uiVAO)
	this.bLoaded = false
}
func GetShaderProgram() *CShaderProgram {
	return &spTerrain
}
func ReleaseTerrainShaderProgram() {
	spTerrain.DeleteProgram()
	for i := 0; i < NUMTERRAINSHADERS; i++ {
		shShaders[i].DeleteShader()
	}
}
func (this *CMultiLayeredHeightmap) GetNumHeightmapRows() int {
	return this.iRows
}

func (this *CMultiLayeredHeightmap) GetNumHeightmapCols() int {
	return this.iCols
}
