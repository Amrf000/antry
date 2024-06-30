package graphic

import (
	"bytes"
	"encoding/binary"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"unsafe"
)

type CSkybox struct {
	uiVAO                                       uint32
	vboRenderData                               *CVertexBufferObject
	tTextures                                   [6]CTexture
	sDirectory                                  string
	sFront, sBack, sLeft, sRight, sTop, sBottom string
}

func EncodeToBytes(p interface{}) []byte {
	var bin_buf bytes.Buffer
	err := binary.Write(&bin_buf, binary.LittleEndian, p)
	if err != nil {
		panic(err)
	}
	return bin_buf.Bytes()
}
func (this *CSkybox) LoadSkybox(a_sDirectory, a_sFront, a_sBack, a_sLeft, a_sRight, a_sTop, a_sBottom string) {
	this.tTextures[0].LoadTexture2D(a_sDirectory+a_sFront, false)
	this.tTextures[1].LoadTexture2D(a_sDirectory+a_sBack, false)
	this.tTextures[2].LoadTexture2D(a_sDirectory+a_sLeft, false)
	this.tTextures[3].LoadTexture2D(a_sDirectory+a_sRight, false)
	this.tTextures[4].LoadTexture2D(a_sDirectory+a_sTop, false)
	this.tTextures[5].LoadTexture2D(a_sDirectory+a_sBottom, false)

	this.sDirectory = a_sDirectory

	this.sFront = a_sFront
	this.sBack = a_sBack
	this.sLeft = a_sLeft
	this.sRight = a_sRight
	this.sTop = a_sTop
	this.sBottom = a_sBottom

	for i := 0; i < 6; i++ {
		this.tTextures[i].SetFiltering(TEXTURE_FILTER_MAG_BILINEAR, TEXTURE_FILTER_MIN_BILINEAR)
		this.tTextures[i].SetSamplerParameter(gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
		this.tTextures[i].SetSamplerParameter(gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	}

	gl.GenVertexArrays(1, &this.uiVAO)
	gl.BindVertexArray(this.uiVAO)
	this.vboRenderData = NewCVertexBufferObject()
	this.vboRenderData.CreateVBO(0)
	this.vboRenderData.BindVBO(gl.ARRAY_BUFFER)

	var vSkyBoxVertices [24]mgl32.Vec3 = [24]mgl32.Vec3{
		// Front face
		mgl32.Vec3{200.0, 200.0, 200.0}, mgl32.Vec3{200.0, -200.0, 200.0}, mgl32.Vec3{-200.0, 200.0, 200.0}, mgl32.Vec3{-200.0, -200.0, 200.0},
		// Back face
		mgl32.Vec3{-200.0, 200.0, -200.0}, mgl32.Vec3{-200.0, -200.0, -200.0}, mgl32.Vec3{200.0, 200.0, -200.0}, mgl32.Vec3{200.0, -200.0, -200.0},
		// Left face
		mgl32.Vec3{-200.0, 200.0, 200.0}, mgl32.Vec3{-200.0, -200.0, 200.0}, mgl32.Vec3{-200.0, 200.0, -200.0}, mgl32.Vec3{-200.0, -200.0, -200.0},
		// Right face
		mgl32.Vec3{200.0, 200.0, -200.0}, mgl32.Vec3{200.0, -200.0, -200.0}, mgl32.Vec3{200.0, 200.0, 200.0}, mgl32.Vec3{200.0, -200.0, 200.0},
		// Top face
		mgl32.Vec3{-200.0, 200.0, -200.0}, mgl32.Vec3{200.0, 200.0, -200.0}, mgl32.Vec3{-200.0, 200.0, 200.0}, mgl32.Vec3{200.0, 200.0, 200.0},
		// Bottom face
		mgl32.Vec3{200.0, -200.0, -200.0}, mgl32.Vec3{-200.0, -200.0, -200.0}, mgl32.Vec3{200.0, -200.0, 200.0}, mgl32.Vec3{-200.0, -200.0, 200.0},
	}

	var vSkyBoxTexCoords [4]mgl32.Vec2 = [4]mgl32.Vec2{
		mgl32.Vec2{0.0, 1.0}, mgl32.Vec2{0.0, 0.0}, mgl32.Vec2{1.0, 1.0}, mgl32.Vec2{1.0, 0.0},
	}

	var vSkyBoxNormals [6]mgl32.Vec3 = [6]mgl32.Vec3{
		mgl32.Vec3{0.0, 0.0, -1.0},
		mgl32.Vec3{0.0, 0.0, 1.0},
		mgl32.Vec3{1.0, 0.0, 0.0},
		mgl32.Vec3{-1.0, 0.0, 0.0},
		mgl32.Vec3{0.0, -1.0, 0.0},
		mgl32.Vec3{0.0, 1.0, 0.0},
	}

	for i := 0; i < 24; i++ {
		this.vboRenderData.AddData(EncodeToBytes(vSkyBoxVertices[i]), int32(unsafe.Sizeof(mgl32.Vec3{})))
		this.vboRenderData.AddData(EncodeToBytes(vSkyBoxTexCoords[i%4]), int32(unsafe.Sizeof(mgl32.Vec2{})))
		this.vboRenderData.AddData(EncodeToBytes(vSkyBoxNormals[i/4]), int32(unsafe.Sizeof(mgl32.Vec3{})))
	}

	this.vboRenderData.UploadDataToGPU(gl.STATIC_DRAW)

	// Vertex positions
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, int32(2*unsafe.Sizeof(mgl32.Vec3{})+unsafe.Sizeof(mgl32.Vec2{})), nil)
	// Texture coordinates
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, int32(2*unsafe.Sizeof(mgl32.Vec3{})+unsafe.Sizeof(mgl32.Vec2{})), unsafe.Pointer(unsafe.Sizeof(mgl32.Vec3{})))
	// Normal vectors
	gl.EnableVertexAttribArray(2)
	gl.VertexAttribPointer(2, 3, gl.FLOAT, false, int32(2*unsafe.Sizeof(mgl32.Vec3{})+unsafe.Sizeof(mgl32.Vec2{})),
		unsafe.Pointer(unsafe.Sizeof(mgl32.Vec3{})+unsafe.Sizeof(mgl32.Vec2{})))
}

func (this *CSkybox) RenderSkybox() {
	gl.DepthMask(false)
	gl.BindVertexArray(this.uiVAO)
	for i := 0; i < 6; i++ {
		this.tTextures[i].BindTexture(0)
		gl.DrawArrays(gl.TRIANGLE_STRIP, int32(i*4), 4)
	}
	gl.DepthMask(true)
}

func (this *CSkybox) DeleteSkybox() {
	for i := 0; i < 6; i++ {
		this.tTextures[i].DeleteTexture()
	}
	gl.DeleteVertexArrays(1, &this.uiVAO)
	this.vboRenderData.DeleteVBO()
}
