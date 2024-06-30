package graphic

import (
	"fmt"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/veandco/go-sdl2/sdl"
	"math"
	"os"
)

var vboSceneObjects CVertexBufferObject
var uiVAOSceneObjects uint32

var ftFont CFreeTypeFont

var sbMainSkybox CSkybox
var cCamera *CFlyingCamera

var dlSun *CDirectionalLight

var amModels [3]CAssimpModel

var hmWorld CMultiLayeredHeightmap

/*-----------------------------------------------

Name:    InitScene

Params:  lpParam - Pointer to anything you want.

Result:  Initializes OpenGL features that will
         be used.

/*---------------------------------------------*/

func InitScene() {
	//gl.ClearColor(1.0, 0.0, 0.0, 1.0)
	gl.ClearColor(1.0, 1.0, 1.0, 0.1)

	if !PrepareShaderPrograms() {
		os.Exit(0)
		return
	}

	LoadAllTextures()

	gl.Enable(gl.DEPTH_TEST)
	gl.ClearDepth(1.0)

	// Here we load font with pixel size 32 - this means that if we print with size above 32, the quality will be low
	ftFont.LoadSystemFont("arial.ttf", 32)
	ftFont.SetShaderProgram(&spFont2D)

	cCamera = NewCFlyingCameraEx(mgl32.Vec3{0.0, 30.0, 100.0}, mgl32.Vec3{0.0, 30.0, 99.0}, mgl32.Vec3{0.0, 1.0, 0.0}, 25.0, 0.1)
	cCamera.SetMovingKeys('W', 'S', 'A', 'D')

	sbMainSkybox.LoadSkybox("data\\skyboxes\\elbrus\\", "elbrus_front.jpg", "elbrus_back.jpg",
		"elbrus_right.jpg", "elbrus_left.jpg", "elbrus_top.jpg", "elbrus_top.jpg")

	dlSun = NewCDirectionalLightEx(mgl32.Vec3{1.0, 1.0, 1.0}, mgl32.Vec3{float32(math.Sqrt(2.0) / 2), float32(-math.Sqrt(2.0) / 2), 0}, 0.5)

	amModels[0].LoadModelFromFile("data\\models\\Wolf\\Wolf.obj")
	amModels[1].LoadModelFromFile("data\\models\\house\\house.3ds")
	FinalizeVBO()

	if !LoadTerrainShaderProgram() {
		panic("LoadTerrainShaderProgram")
	}
	if !hmWorld.LoadHeightMapFromImage("data\\worlds\\consider_this_question.bmp") {
		panic("LoadHeightMapFromImage")
	}

}

/*
-----------------------------------------------

	Name:    RenderScene

	Params:  lpParam - Pointer to anything you want.

	Result:  Renders whole scene.

	/*---------------------------------------------
*/
var fAngleOfDarkness float32 = 45.0

func RenderScene(oglControl *COpenGLControl) {
	// Typecast lpParam to COpenGLControl pointer
	//var oglControl *COpenGLControl= (COpenGLControl*)lpParam;
	oglControl.ResizeOpenGLViewportFull()

	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	spMain.UseProgram()

	spMain.SetUniformM4("matrices.projMatrix", *oglControl.GetProjectionMatrix())
	spMain.SetUniformM4("matrices.viewMatrix", cCamera.Look())

	spMain.SetUniformI32("gSampler", 0)

	spMain.SetUniformM4("matrices.modelMatrix", mgl32.Ident4())
	spMain.SetUniformM4("matrices.normalMatrix", mgl32.Ident4())
	spMain.SetUniformV4("vColor", mgl32.Vec4{1, 1, 1, 1})

	// This values will set the darkness of whole scene, that's why such name of variable :D
	//var fAngleOfDarkness float32= 45.0f;
	// You can play with direction of light with '+' and '-' key
	keys := sdl.GetKeyboardState()
	if keys[sdl.SCANCODE_KP_PLUS] != 0 {
		fAngleOfDarkness += AppMain.sof(90)
	}
	if keys[sdl.SCANCODE_KP_MEMSUBTRACT] != 0 {
		fAngleOfDarkness -= AppMain.sof(90)
	}
	// Set the directional vector of light
	dlSun.vDirection = mgl32.Vec3{float32(-math.Sin(float64(fAngleOfDarkness * 3.1415 / 180.0))), float32(-math.Cos(float64(fAngleOfDarkness * 3.1415 / 180.0))), 0.0}
	dlSun.SetUniformData(&spMain, "sunLight")

	spMain.SetUniformM4("matrices.modelMatrix", mgl32.Translate3D(cCamera.vEye.X(), cCamera.vEye.Y(), cCamera.vEye.Z()).Mul4(mgl32.Ident4()))
	sbMainSkybox.RenderSkybox()

	spMain.SetUniformM4("matrices.modelMatrix", mgl32.Ident4())

	// Render a house

	BindModelsVAO()

	var mModel mgl32.Mat4 = mgl32.Translate3D(40.0, 17.0, 0).Mul4(mgl32.Ident4())
	mModel = mgl32.Scale3D(8, 8, 8).Mul4(mModel) // Casino :D

	spMain.SetModelAndNormalMatrix("matrices.modelMatrix", "matrices.normalMatrix", mModel)
	amModels[1].RenderModel()

	// ... and also ONE wolf now only :P

	mModel = mgl32.Translate3D(-20.0, 22.0, 50).Mul4(mgl32.Ident4())
	mModel = mgl32.Scale3D(2.8, 2.8, 2.8).Mul4(mModel)

	spMain.SetModelAndNormalMatrix("matrices.modelMatrix", "matrices.normalMatrix", mModel)
	amModels[0].RenderModel()

	// Now we're going to render terrain

	hmWorld.SetRenderSize3(300.0, 35.0, 300.0)
	var spTerrain *CShaderProgram = GetShaderProgram()

	spTerrain.UseProgram()

	spTerrain.SetUniformM4("matrices.projMatrix", *oglControl.GetProjectionMatrix())
	spTerrain.SetUniformM4("matrices.viewMatrix", cCamera.Look())

	// We bind all 5 textures - 3 of them are textures for layers, 1 texture is a "path" texture, and last one is
	// the places in heightmap where path should be and how intense should it be
	for i := 0; i < 5; i++ {
		var sSamplerName string
		sSamplerName = fmt.Sprintf("gSampler[%d]", i)
		tTextures[i].BindTexture(uint32(i))
		spTerrain.SetUniformI32(sSamplerName, int32(i))
	}

	// ... set some uniforms
	spTerrain.SetModelAndNormalMatrix("matrices.modelMatrix", "matrices.normalMatrix", mgl32.Ident4())
	spTerrain.SetUniformV4("vColor", mgl32.Vec4{1, 1, 1, 1})

	dlSun.SetUniformData(spTerrain, "sunLight")

	// ... and finally render heightmap
	hmWorld.RenderHeightmap()

	cCamera.Update()

	// Print something over scene

	spFont2D.UseProgram()
	gl.Disable(gl.DEPTH_TEST)
	spFont2D.SetUniformM4("matrices.projMatrix", *oglControl.GetOrthoMatrix())

	//var w int32 = oglControl.GetViewportWidth()
	var h int32 = oglControl.GetViewportHeight()

	spFont2D.SetUniformV4("vColor", mgl32.Vec4{1.0, 1.0, 1.0, 1.0})
	ftFont.Print("www.mbsoftworks.sk", 20, 20, 24)

	ftFont.PrintFormatted(20, int(h-30), 20, fmt.Sprintf("FPS: %d", oglControl.GetFPS()))
	ftFont.PrintFormatted(20, int(h-80), 20, fmt.Sprintf("Heightmap size: %dx%d", hmWorld.GetNumHeightmapRows(), hmWorld.GetNumHeightmapCols()))

	gl.Enable(gl.DEPTH_TEST)

	keys = sdl.GetKeyboardState()
	if keys[sdl.SCANCODE_ESCAPE] != 0 {
		os.Exit(0)
	}

	oglControl.SwapBuffers()
}

/*-----------------------------------------------

  Name:    ReleaseScene

  Params:  lpParam - Pointer to anything you want.

  Result:  Releases OpenGL scene.

  /*---------------------------------------------*/

func ReleaseScene() {
	for i := 0; i < NUMTEXTURES; i++ {
		tTextures[i].DeleteTexture()
	}
	sbMainSkybox.DeleteSkybox()

	spMain.DeleteProgram()
	spOrtho2D.DeleteProgram()
	spFont2D.DeleteProgram()
	for i := 0; i < NUMSHADERS; i++ {
		shShaders[i].DeleteShader()
	}
	ftFont.DeleteFont()

	gl.DeleteVertexArrays(1, &uiVAOSceneObjects)
	vboSceneObjects.DeleteVBO()

	hmWorld.ReleaseHeightmap()
	ReleaseTerrainShaderProgram()
}
