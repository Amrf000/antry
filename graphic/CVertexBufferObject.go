package graphic

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"unsafe"
)

type CVertexBufferObject struct {
	uiBuffer     uint32
	iSize        int
	iCurrentSize int32
	iBufferType  uint32
	data         []byte

	bDataUploaded bool
}

func NewCVertexBufferObject() *CVertexBufferObject {
	this := CVertexBufferObject{}
	this.bDataUploaded = false
	this.uiBuffer = 0
	return &this
}

func (this *CVertexBufferObject) CreateVBO(a_iSize int) {
	gl.GenBuffers(1, &this.uiBuffer)
	this.data = make([]byte, a_iSize)
	this.iSize = a_iSize
	this.iCurrentSize = 0
}

func (this *CVertexBufferObject) DeleteVBO() {
	gl.DeleteBuffers(1, &this.uiBuffer)
	this.bDataUploaded = false
	this.data = nil
}

func (this *CVertexBufferObject) MapBufferToMemory(iUsageHint uint32) unsafe.Pointer {
	if !this.bDataUploaded {
		return nil
	}
	ptrRes := gl.MapBuffer(this.iBufferType, iUsageHint)
	return ptrRes
}

func (this *CVertexBufferObject) MapSubBufferToMemory(iUsageHint uint32, uiOffset, uiLength int) unsafe.Pointer {
	if !this.bDataUploaded {
		return nil
	}
	ptrRes := gl.MapBufferRange(this.iBufferType, uiOffset, uiLength, iUsageHint)
	return ptrRes
}

func (this *CVertexBufferObject) UnmapBuffer() {
	gl.UnmapBuffer(this.iBufferType)
}

func (this *CVertexBufferObject) BindVBO(a_iBufferType uint32) {
	this.iBufferType = a_iBufferType
	gl.BindBuffer(this.iBufferType, this.uiBuffer)
}

func (this *CVertexBufferObject) UploadDataToGPU(iDrawingHint uint32) {
	gl.BufferData(this.iBufferType, len(this.data), unsafe.Pointer(&this.data[0]), iDrawingHint)
	this.bDataUploaded = true
	this.data = nil
}

func (this *CVertexBufferObject) AddData(ptrData []byte, uiDataSize int32) {
	this.data = append(this.data, ptrData[0:uiDataSize]...)
	this.iCurrentSize += uiDataSize
}
func (this *CVertexBufferObject) GetDataPointer() unsafe.Pointer {
	if this.bDataUploaded {
		return nil
	}
	return unsafe.Pointer(&this.data[0])
}

func (this *CVertexBufferObject) GetBufferID() uint32 {
	return this.uiBuffer
}

func (this *CVertexBufferObject) GetCurrentSize() int32 {
	return this.iCurrentSize
}
