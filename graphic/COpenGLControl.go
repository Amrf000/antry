package graphic

import (
	"fmt"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/veandco/go-sdl2/sdl"
	"os"
	"time"
)

var bClassRegistered bool = false
var bGlewInitialized bool = false

type Greeting func()
type GreetingEx func(oglControl *COpenGLControl)
type COpenGLControl struct {
	SdlWindow  *sdl.Window
	SdlContext sdl.GLContext
	// hDC HDC
	//hWnd *HWND
	//hRC HGLRC

	iMajorVersion int
	iMinorVersion int

	// Used for FPS calculation
	iFPSCount   int
	iCurrentFPS int
	tLastSecond time.Time

	// Matrix for perspective projection
	mProjection mgl32.Mat4
	// Matrix for orthographic 2D projection
	mOrtho mgl32.Mat4

	// Viewport parameters
	iViewportWidth  int32
	iViewportHeight int32

	ptrInitScene    Greeting
	ptrRenderScene  GreetingEx
	ptrReleaseScene Greeting
}

func NewCOpenGLControl() *COpenGLControl {
	this := COpenGLControl{}
	this.iFPSCount = 0
	this.iCurrentFPS = 0
	return &this
}

/*
-----------------------------------------------

	Name:	InitGLEW

	Params:	none

	Result:	Creates fake window and OpenGL rendering
			context, within which GLEW is initialized.

	/*---------------------------------------------
*/
const SIMPLE_OPENGL_CLASS_NAME = "simple_openGL_class_name"

func (this *COpenGLControl) InitGLEW() bool {
	if bGlewInitialized {
		return true
	}

	this.RegisterSimpleOpenGLClass()
	var err error
	this.SdlWindow, err = sdl.CreateWindow(SIMPLE_OPENGL_CLASS_NAME, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		800, 600, sdl.WINDOW_OPENGL) //sdl.WINDOW_SHOWN) //"FAKE", WS_OVERLAPPEDWINDOW | WS_MAXIMIZE | WS_CLIPCHILDREN,
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create window: %s\n", err)
		return false
	}
	this.SdlContext, err = this.SdlWindow.GLCreateContext()
	if err != nil {
		panic(err)
	}
	if err := gl.Init(); err != nil {
		fmt.Println("无法初始化OpenGL:", err)
		return false
	}

	//this.hDC = GetDC(hWndFake);

	// First, choose false pixel format

	//	var pfd PIXELFORMATDESCRIPTOR
	//	memset(&pfd, 0, sizeof(PIXELFORMATDESCRIPTOR));
	//	pfd.nSize		= sizeof(PIXELFORMATDESCRIPTOR);
	//	pfd.nVersion   = 1;
	//	pfd.dwFlags    = PFD_DOUBLEBUFFER | PFD_SUPPORT_OPENGL | PFD_DRAW_TO_WINDOW;
	//	pfd.iPixelType = PFD_TYPE_RGBA;
	//	pfd.cColorBits = 32;
	//	pfd.cDepthBits = 32;
	//	pfd.iLayerType = PFD_MAIN_PLANE;
	//
	//	iPixelFormat := ChoosePixelFormat(this.hDC, &pfd);
	//	if (iPixelFormat == 0) {
	//		return false
	//	}
	//
	//	if !SetPixelFormat(this.hDC, iPixelFormat, &pfd)  {
	//		return false
	//	}
	//
	//// Create the false, old style context (OpenGL 2.1 and before)
	//
	//	var hRCFake HGLRC  = wglCreateContext(this.hDC);
	//	wglMakeCurrent(this.hDC, hRCFake);
	//
	//	bResult := true;
	//
	//	if(!bGlewInitialized)  {
	//		if(glewInit() != GLEW_OK)  {
	//			MessageBox(*this.hWnd, "Couldn't initialize GLEW!", "Fatal Error", MB_ICONERROR);
	//			bResult = false;
	//		}
	//		bGlewInitialized = true;
	//	}
	//
	//	wglMakeCurrent(NULL, NULL);
	//	wglDeleteContext(hRCFake);
	//	DestroyWindow(hWndFake);

	//return bResult;
	return true
}

/*-----------------------------------------------

  Name:	InitOpenGL

  Params:	hInstance - application instance
  		a_hWnd - window to init OpenGL into
  		a_iMajorVersion - Major version of OpenGL
  		a_iMinorVersion - Minor version of OpenGL
  		a_initScene - pointer to init function
  		a_renderScene - pointer to render function
  		a_releaseScene - optional parameter of release
  						function

  Result:	Initializes OpenGL rendering context
  		of specified version. If succeeds,
  		returns true.

  /*---------------------------------------------*/

func (this *COpenGLControl) InitOpenGL(iMajorVersion int, iMinorVersion int, a_ptrInitScene Greeting, a_ptrRenderScene GreetingEx, a_ptrReleaseScene Greeting) bool {
	if !this.InitGLEW() {
		return false
	}
	//
	//hWnd = a_hWnd;
	//hDC = GetDC(*hWnd);
	//
	//bError := false;
	//var pfd PIXELFORMATDESCRIPTOR
	//
	//if iMajorVersion <= 2 {
	//	memset(&pfd, 0, sizeof(PIXELFORMATDESCRIPTOR));
	//	pfd.nSize		= sizeof(PIXELFORMATDESCRIPTOR);
	//	pfd.nVersion   = 1;
	//	pfd.dwFlags    = PFD_DOUBLEBUFFER | PFD_SUPPORT_OPENGL | PFD_DRAW_TO_WINDOW;
	//	pfd.iPixelType = PFD_TYPE_RGBA;
	//	pfd.cColorBits = 32;
	//	pfd.cDepthBits = 32;
	//	pfd.iLayerType = PFD_MAIN_PLANE;
	//
	//	var iPixelFormat int= ChoosePixelFormat(hDC, &pfd);
	//	if (iPixelFormat == 0){
	//		return false;
	//	}
	//
	//	if(!SetPixelFormat(hDC, iPixelFormat, &pfd)){
	//		return false;
	//	}
	//
	//	// Create the old style context (OpenGL 2.1 and before)
	//	hRC = wglCreateContext(hDC);
	//	if(hRC){
	//		wglMakeCurrent(hDC, hRC);
	//	}else {
	//		bError = true;
	//	}
	//} else if(WGLEW_ARB_create_context && WGLEW_ARB_pixel_format) {
	//	var iPixelFormatAttribList = []int {
	//		WGL_DRAW_TO_WINDOW_ARB, GL_TRUE,
	//		WGL_SUPPORT_OPENGL_ARB, GL_TRUE,
	//		WGL_DOUBLE_BUFFER_ARB, GL_TRUE,
	//		WGL_PIXEL_TYPE_ARB, WGL_TYPE_RGBA_ARB,
	//		WGL_COLOR_BITS_ARB, 32,
	//		WGL_DEPTH_BITS_ARB, 24,
	//		WGL_STENCIL_BITS_ARB, 8,
	//		0 // End of attributes list
	//	}
	//	var  iContextAttribs = []int{
	//		WGL_CONTEXT_MAJOR_VERSION_ARB, iMajorVersion,
	//		WGL_CONTEXT_MINOR_VERSION_ARB, iMinorVersion,
	//		WGL_CONTEXT_FLAGS_ARB, WGL_CONTEXT_FORWARD_COMPATIBLE_BIT_ARB,
	//		0 // End of attributes list
	//	}
	//
	//	var iPixelFormat int
	//	var iNumFormats int
	//	wglChoosePixelFormatARB(hDC, iPixelFormatAttribList, NULL, 1, &iPixelFormat, (UINT*)&iNumFormats);
	//
	//	// PFD seems to be only redundant parameter now
	//	if(!SetPixelFormat(hDC, iPixelFormat, &pfd)){
	//		return false;
	//	}
	//
	//	hRC = wglCreateContextAttribsARB(hDC, 0, iContextAttribs);
	//	// If everything went OK
	//	if(hRC){
	//		wglMakeCurrent(hDC, hRC);
	//	}else {
	//		bError = true;
	//	}
	//
	//} else {
	//	bError = true;
	//}
	//
	//if(bError) {
	//	// Generate error messages
	//	var sErrorMessage string
	//	var sErrorTitle string
	//	fmt.Sprintf(sErrorMessage, "OpenGL %d.%d is not supported! Please download latest GPU drivers!", iMajorVersion, iMinorVersion);
	//	fmt.Sprintf(sErrorTitle, "OpenGL %d.%d Not Supported", iMajorVersion, iMinorVersion);
	//	MessageBox(*hWnd, sErrorMessage, sErrorTitle, MB_ICONINFORMATION);
	//	return false;
	//}
	//
	this.ptrRenderScene = a_ptrRenderScene
	this.ptrInitScene = a_ptrInitScene
	this.ptrReleaseScene = a_ptrReleaseScene
	//
	if this.ptrInitScene != nil {
		this.ptrInitScene()
	}

	return true
}

/*-----------------------------------------------

  Name:	ResizeOpenGLViewportFull

  Params:	none

  Result:	Resizes viewport to full window.

  /*---------------------------------------------*/

func (this *COpenGLControl) ResizeOpenGLViewportFull() {
	//if this.hWnd == nil {
	//	return
	//}
	//var rRect RECT;
	//GetClientRect(*this.hWnd, &rRect);
	w, h := this.SdlWindow.GetSize()
	gl.Viewport(0, 0, w, h)
	this.iViewportWidth = w
	this.iViewportHeight = h
}

/*-----------------------------------------------

  Name:	SetProjection3D

  Params:	fFOV - field of view angle
  		fAspectRatio - aspect ration of width/height
  		fNear, fFar - distance of near and far clipping plane

  Result:	Calculates projection matrix and stores it.

  /*---------------------------------------------*/

func (this *COpenGLControl) SetProjection3D(fFOV, fAspectRatio, fNear, fFar float32) {
	this.mProjection = mgl32.Perspective(fFOV, fAspectRatio, fNear, fFar) //fFOV
	// mgl32.Perspective(mgl32.DegToRad(45.0), float32(600)/float32(600), 0.1, 3000.0)
}

/*-----------------------------------------------

  Name:	SetOrtho2D

  Params:	width - width of window
  				height - height of window

  Result:	Calculates ortho 2D projection matrix and stores it.

  /*---------------------------------------------*/

func (this *COpenGLControl) SetOrtho2D(width, height int32) {
	this.mOrtho = mgl32.Ortho2D(float32(0.0), float32(width), float32(0.0), float32(height))
}

/*-----------------------------------------------

  Name:	GetProjectionMatrix()

  Params:	none

  Result:	Retrieves pointer to projection matrix.

  /*---------------------------------------------*/

func (this *COpenGLControl) GetProjectionMatrix() *mgl32.Mat4 {
	return &this.mProjection
}

/*-----------------------------------------------

  Name:	GetOrthoMatrix()

  Params:	none

  Result:	Retrieves pointer to ortho matrix.

  /*---------------------------------------------*/

func (this *COpenGLControl) GetOrthoMatrix() *mgl32.Mat4 {
	return &this.mOrtho
}

/*-----------------------------------------------

  Name:	RegisterSimpleOpenGLClass

  Params:	hInstance - application instance

  Result:	Registers simple OpenGL window class.

  /*---------------------------------------------*/

func (this *COpenGLControl) RegisterSimpleOpenGLClass() {
	if bClassRegistered {
		return
	}
	//var wc WNDCLASSEX
	//
	//wc.cbSize = sizeof(WNDCLASSEX);
	//wc.style =  CS_HREDRAW | CS_VREDRAW | CS_OWNDC | CS_DBLCLKS;
	//wc.lpfnWndProc = (WNDPROC) msgHandlerSimpleOpenGLClass;
	//wc.cbClsExtra = 0; wc.cbWndExtra = 0;
	//wc.hInstance = hInstance;
	//wc.hIcon = LoadIcon(hInstance, MAKEINTRESOURCE(IDI_APPLICATION));
	//wc.hIconSm = LoadIcon(hInstance, MAKEINTRESOURCE(IDI_APPLICATION));
	//wc.hCursor = LoadCursor(NULL, IDC_ARROW);
	//wc.hbrBackground = (HBRUSH)(COLOR_MENUBAR+1);
	//wc.lpszMenuName = NULL;
	//wc.lpszClassName = SIMPLE_OPENGL_CLASS_NAME;
	//
	//RegisterClassEx(&wc);

	bClassRegistered = true
}

/*-----------------------------------------------

  Name:	UnregisterSimpleOpenGLClass

  Params:	hInstance - application instance

  Result:	Unregisters simple OpenGL window class.

  /*---------------------------------------------*/

func (this *COpenGLControl) UnregisterSimpleOpenGLClass() {
	if bClassRegistered {
		//UnregisterClass(SIMPLE_OPENGL_CLASS_NAME, hInstance)
		bClassRegistered = false
	}
}

/*-----------------------------------------------

  Name:	msgHandlerSimpleOpenGLClass

  Params:	whatever

  Result:	Handles messages from windows that use
  		simple OpenGL class.

  /*---------------------------------------------*/

func msgHandlerSimpleOpenGLClass() { //CALLBACK
	//var ps PAINTSTRUCT
	//switch uiMsg {
	//case WM_PAINT:
	//	BeginPaint(hWnd, &ps)
	//	EndPaint(hWnd, &ps)
	//default:
	//	return DefWindowProc(hWnd, uiMsg, wParam, lParam) // Default window procedure
	//}
	//return 0
}

/*-----------------------------------------------

  Name:	SwapBuffers

  Params:	none

  Result:	Swaps back and front buffer.

  /*---------------------------------------------*/

func (this *COpenGLControl) SwapBuffers() {
	this.SdlWindow.GLSwap()
}

/*-----------------------------------------------

  Name:	MakeCurrent

  Params:	none

  Result:	Makes current device and OpenGL rendering
  		context to those associated with OpenGL
  		control.

  /*---------------------------------------------*/

func (this *COpenGLControl) MakeCurrent() {
	//wglMakeCurrent(this.hDC, this.hRC)
}

/*-----------------------------------------------

  Name:	Render

  Params:	lpParam - pointer to whatever you want

  Result:	Calls previously set render function.

  /*---------------------------------------------*/

func (this *COpenGLControl) Render() {
	var tCurrent = time.Now()
	if tCurrent.Sub(this.tLastSecond) >= time.Second {
		this.tLastSecond = tCurrent
		this.iFPSCount = this.iCurrentFPS
		this.iCurrentFPS = 0
	}
	if this.ptrRenderScene != nil {
		this.ptrRenderScene(this)
	}
	this.iCurrentFPS++
}

/*-----------------------------------------------

  Name:	ReleaseOpenGLControl

  Params:	lpParam - pointer to whatever you want

  Result:	Calls previously set release function
  		and deletes rendering context.

  /*---------------------------------------------*/

func (this *COpenGLControl) ReleaseOpenGLControl() {
	if this.ptrReleaseScene != nil {
		this.ptrReleaseScene()
	}

	//wglMakeCurrent(NULL, NULL)
	//wglDeleteContext(this.hRC)
	//ReleaseDC(*this.hWnd, this.hDC)
	//
	//this.hWnd = nil
}

/*-----------------------------------------------

  Name:	SetVerticalSynchronization

  Params: bEnabled - whether to enable V-Sync

  Result:	Guess what it does :)

  /*---------------------------------------------*/

func (this *COpenGLControl) SetVerticalSynchronization(bEnabled bool) bool {
	//if !wglSwapIntervalEXT {
	//	return false
	//}
	//
	//if bEnabled {
	//	wglSwapIntervalEXT(1)
	//} else {
	//	wglSwapIntervalEXT(0)
	//}

	return true
}

/*-----------------------------------------------

  Name:	Getters

  Params:	none

  Result:	... They get something :D

  /*---------------------------------------------*/

func (this *COpenGLControl) GetFPS() int {
	return this.iFPSCount
}

func (this *COpenGLControl) GetViewportWidth() int32 {
	return this.iViewportWidth
}

func (this *COpenGLControl) GetViewportHeight() int32 {
	return this.iViewportHeight
}
