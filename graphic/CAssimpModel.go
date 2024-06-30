package graphic

import (
	"github.com/bloeys/assimp-go/asig"
	"github.com/bloeys/gglm/gglm"
	"github.com/go-gl/gl/v4.1-core/gl"
	"unsafe"
)

type CMaterial struct {
	iTexture int
}

var vboModelData CVertexBufferObject
var uiVAO uint32
var CAtTextures []CTexture

type CAssimpModel struct {
	bLoaded           bool
	iMeshStartIndices []int32
	iMeshSizes        []int32
	iMaterialIndices  []uint
	iNumMaterials     int
}

func GetDirectoryPath(sFilePath string) string {
	// Get directory path
	sDirectory := ""
	for i := len(sFilePath) - 1; i >= 0; i-- {
		if sFilePath[i] == '\\' || sFilePath[i] == '/' {
			sDirectory = sFilePath[0 : i+1]
			break
		}
	}

	return sDirectory
}

func NewCAssimpModel() *CAssimpModel {
	this := CAssimpModel{}
	this.bLoaded = false
	return &this
}

func (this *CAssimpModel) LoadModelFromFile(sFilePath string) bool {
	if vboModelData.GetBufferID() == 0 {
		vboModelData.CreateVBO(0)
		CAtTextures = make([]CTexture, 50)
	}
	scene, release, err := asig.ImportFile(sFilePath, asig.PostProcessCalcTangentSpace|
		asig.PostProcessTriangulate|asig.PostProcessJoinIdenticalVertices|
		asig.PostProcessSortByPType)
	if err != nil {
		panic(err)
		//buttons := []sdl.MessageBoxButtonData{
		//	{0, 0, "no"},
		//	{sdl.MESSAGEBOX_BUTTON_RETURNKEY_DEFAULT, 1, "yes"},
		//	{sdl.MESSAGEBOX_BUTTON_ESCAPEKEY_DEFAULT, 2, "cancel"},
		//}
		//
		//colorScheme := sdl.MessageBoxColorScheme{
		//	Colors: [5]sdl.MessageBoxColor{
		//		sdl.MessageBoxColor{255, 0, 0},
		//		sdl.MessageBoxColor{0, 255, 0},
		//		sdl.MessageBoxColor{255, 255, 0},
		//		sdl.MessageBoxColor{0, 0, 255},
		//		sdl.MessageBoxColor{255, 0, 255},
		//	},
		//}
		//sdl.ShowMessageBox(&sdl.MessageBoxData{
		//	Flags:       sdl.MESSAGEBOX_INFORMATION,
		//	Window:      nil,
		//	Title:       "Couldn't load model ",
		//	Message:     "Error Importing Asset",
		//	Buttons:     buttons,
		//	ColorScheme: &colorScheme}) //AppMain.hWnd, "Couldn't load model ", "Error Importing Asset", MB_ICONERROR
		return false
	}
	defer release()
	const iVertexTotalSize int32 = int32(unsafe.Sizeof(gglm.Vec3{})*2 + unsafe.Sizeof(gglm.Vec2{}))
	var iTotalVertices int = 0
	for i := 0; i < len(scene.Meshes); i++ {
		var mesh *asig.Mesh = scene.Meshes[i]
		var iMeshFaces int = len(mesh.Faces)
		this.iMaterialIndices = append(this.iMaterialIndices, mesh.MaterialIndex)
		var iSizeBefore int32 = vboModelData.GetCurrentSize()
		this.iMeshStartIndices = append(this.iMeshStartIndices, iSizeBefore/iVertexTotalSize)
		for j := 0; j < iMeshFaces; j++ {
			var face *asig.Face = &mesh.Faces[j]
			for k := 0; k < 3; k++ {
				var pos gglm.Vec3 = mesh.Vertices[face.Indices[k]]
				var uv gglm.Vec3 = mesh.TexCoords[0][face.Indices[k]]
				var normal *gglm.Vec3
				if len(mesh.Normals) > 0 {
					normal = &mesh.Normals[face.Indices[k]]
				} else {
					normal = gglm.NewVec3(1.0, 1.0, 1.0)
				}
				vboModelData.AddData(EncodeToBytes(pos), int32(unsafe.Sizeof(gglm.Vec3{})))
				vboModelData.AddData(EncodeToBytes(uv), int32(unsafe.Sizeof(gglm.Vec3{})))
				vboModelData.AddData(EncodeToBytes(normal), int32(unsafe.Sizeof(gglm.Vec3{})))
			}
		}
		var iMeshVertices int = len(mesh.Vertices)
		iTotalVertices += iMeshVertices
		this.iMeshSizes = append(this.iMeshSizes, (vboModelData.GetCurrentSize()-iSizeBefore)/iVertexTotalSize)
	}

	this.iNumMaterials = len(scene.Materials)

	materialRemap := make([]uint, this.iNumMaterials)

	for i := 0; i < this.iNumMaterials; i++ {
		var material *asig.Material = scene.Materials[i]
		var texIndex uint = 0
		info, err := asig.GetMaterialTexture(material, asig.TextureTypeDiffuse, texIndex)
		if err == nil {
			var sDir string = GetDirectoryPath(sFilePath)

			var sTextureName string = info.Path

			var sFullPath string = sDir + sTextureName

			var iTexFound int = -1
			for j := 0; j < len(CAtTextures); j++ {
				if sFullPath == CAtTextures[j].GetPath() {
					iTexFound = int(j)
					break
				}
			}
			if iTexFound != -1 {
				materialRemap[i] = uint(iTexFound)
			} else {
				var tNew CTexture
				tNew.LoadTexture2D(sFullPath, true)
				materialRemap[i] = uint(len(CAtTextures))
				CAtTextures = append(CAtTextures, tNew)
			}
		}
	}

	for i := 0; i < len(this.iMeshSizes); i++ {
		var iOldIndex uint = this.iMaterialIndices[i]
		this.iMaterialIndices[i] = materialRemap[iOldIndex]
	}
	this.bLoaded = true
	return this.bLoaded
}

/*-----------------------------------------------

  Name:	FinalizeVBO

  Params: none

  Result: Uploads all loaded model data in one global
  		models' VBO.

  /*---------------------------------------------*/

func FinalizeVBO() {
	gl.GenVertexArrays(1, &uiVAO)
	gl.BindVertexArray(uiVAO)
	vboModelData.BindVBO(gl.ARRAY_BUFFER)
	vboModelData.UploadDataToGPU(gl.STATIC_DRAW)
	// Vertex positions
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, int32(2*unsafe.Sizeof(gglm.Vec3{})+unsafe.Sizeof(gglm.Vec2{})), nil)
	// Texture coordinates
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, int32(2*unsafe.Sizeof(gglm.Vec3{})+unsafe.Sizeof(gglm.Vec2{})),
		unsafe.Pointer(unsafe.Sizeof(gglm.Vec3{})))
	// Normal vectors
	gl.EnableVertexAttribArray(2)
	gl.VertexAttribPointer(2, 3, gl.FLOAT, false, int32(2*unsafe.Sizeof(gglm.Vec3{})+unsafe.Sizeof(gglm.Vec2{})),
		unsafe.Pointer(unsafe.Sizeof(gglm.Vec3{})+unsafe.Sizeof(gglm.Vec2{})))
}

/*-----------------------------------------------

  Name:	BindModelsVAO

  Params: none

  Result: Binds VAO of models with their VBO.

  /*---------------------------------------------*/

func BindModelsVAO() {
	gl.BindVertexArray(uiVAO)
}

/*-----------------------------------------------

  Name:	RenderModel

  Params: none

  Result: Guess what it does ^^.

  /*---------------------------------------------*/

func (this *CAssimpModel) RenderModel() {
	if !this.bLoaded {
		return
	}
	var iNumMeshes int = len(this.iMeshSizes)
	for i := 0; i < iNumMeshes; i++ {
		var iMatIndex uint = this.iMaterialIndices[i]
		CAtTextures[iMatIndex].BindTexture(0)
		gl.DrawArrays(gl.TRIANGLES, this.iMeshStartIndices[i], this.iMeshSizes[i])
	}
}
