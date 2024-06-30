package graphic

import (
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/veandco/go-sdl2/sdl"
	"sync"
	"time"
)

type COpenGLWinApp struct {
	//hWnd HWND ; // Handle to application window
	oglControl COpenGLControl // OpenGL Control
	//hInstance HINSTANCE // Application's instance
	sAppName string
	hMutex   sync.Mutex

	bAppActive     bool // To check if application is active (not minimized)
	tLastFrame     time.Time
	fFrameInterval float32
}

var AppMain COpenGLWinApp

func (this *COpenGLWinApp) ResetTimer() {
	this.tLastFrame = time.Now()
	this.fFrameInterval = 0.0
}

func (this *COpenGLWinApp) UpdateTimer() {
	var tCur = time.Now()
	this.fFrameInterval = float32(tCur.Sub(this.tLastFrame)) / float32(time.Second)
	this.tLastFrame = tCur
}

func (this *COpenGLWinApp) sof(fVal float32) float32 {
	return fVal * this.fFrameInterval
}

func (this *COpenGLWinApp) InitializeApp(a_sAppName string) bool {
	this.sAppName = a_sAppName

	//if(GetLastError() == ERROR_ALREADY_EXISTS) {
	//	MessageBox(NULL, "This application already runs!", "Multiple Instances Found.", MB_ICONINFORMATION | MB_OK);
	//	return 0;
	//}
	return true
}

//func GlobalMessageHandler() {
//	AppMain.msgHandlerMain()
//}

func (this *COpenGLWinApp) RegisterAppClass() {

}

func (this *COpenGLWinApp) CreateAppWindow(sTitle string) bool {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	// defer sdl.Quit()

	//if(MessageBox(NULL, "Would you like to run in fullscreen?", "Fullscreen", MB_ICONQUESTION | MB_YESNO) == IDYES) {
	//	DEVMODE dmSettings = {0};
	//	EnumDisplaySettings(NULL, ENUM_CURRENT_SETTINGS, &dmSettings); // Get current display settings
	//
	//	hWnd = CreateWindowEx(0, sAppName.c_str(), sTitle.c_str(), WS_POPUP | WS_CLIPSIBLINGS | WS_CLIPCHILDREN, // This is the commonly used style for fullscreen
	//		0, 0, dmSettings.dmPelsWidth, dmSettings.dmPelsHeight, NULL,
	//		NULL, hInstance, NULL);
	//} else {
	//	hWnd = CreateWindowEx(0, sAppName.c_str(), sTitle.c_str(), WS_OVERLAPPEDWINDOW|WS_CLIPCHILDREN,
	//		0, 0, CW_USEDEFAULT, CW_USEDEFAULT, NULL,
	//		NULL, hInstance, NULL)
	//}
	//
	// 初始化OpenGL和GLFW
	//err := glfw.Init()
	//if err != nil {
	//	fmt.Println("无法初始化GLFW:", err)
	//	return false
	//}
	//defer glfw.Terminate()

	// 创建窗口
	//glfw.WindowHint(glfw.Resizable, glfw.False)
	//window, err := glfw.CreateWindow(600, 600, "OpenGL Texture Example", nil, nil)
	//if err != nil {
	//	fmt.Println("无法创建窗口:", err)
	//	return false
	//}
	//window.MakeContextCurrent()
	// 初始化OpenGL
	if !this.oglControl.InitOpenGL(3, 3, InitScene, RenderScene, ReleaseScene) { //, &this.oglControl
		return false
	}
	//
	//ShowWindow(hWnd, SW_SHOW);
	//
	//// Just to send WM_SIZE message again
	//ShowWindow(hWnd, SW_SHOWMAXIMIZED);
	//UpdateWindow(hWnd);
	return true
}

func (this *COpenGLWinApp) AppBody() {
	this.bAppActive = true
	this.oglControl.ResizeOpenGLViewportFull()
	this.oglControl.SetProjection3D(mgl32.DegToRad(45.0), float32(600)/float32(600), 0.5, 1000.0)
	this.oglControl.SetOrtho2D(600, 600)
	for {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch t := event.(type) {
			case sdl.QuitEvent:
				this.bAppActive = false
			case sdl.MouseMotionEvent:
			//	fmt.Printf("[%d ms] MouseMotion\tid:%d\tx:%d\ty:%d\txrel:%d\tyrel:%d\n", t.Timestamp, t.Which, t.X, t.Y, t.XRel, t.YRel)
			case sdl.WindowEvent:
				switch t.Event {
				case sdl.WINDOWEVENT_FOCUS_GAINED:
					this.bAppActive = true
					//fmt.Println("Window focus gained")
				case sdl.WINDOWEVENT_FOCUS_LOST:
					this.bAppActive = false
					//fmt.Println("Window focus lost")
				//case sdl.WINDOWEVENT_MINIMIZED:
				//	this.bAppActive = false
				//	fmt.Println("Window mini")
				case sdl.WINDOWEVENT_ENTER:
					//fmt.Println("Mouse entered window")
				case sdl.WINDOWEVENT_LEAVE:
					//fmt.Println("Mouse left window")
				case sdl.WINDOWEVENT_CLOSE:
					//fmt.Println("Window close requested")
					return
				case sdl.WINDOWEVENT_RESIZED:
					we := event.(sdl.WindowEvent)
					fmt.Printf("Window resized to %d x %d\n", we.Data1, we.Data2)
					this.oglControl.ResizeOpenGLViewportFull()
					this.oglControl.SetProjection3D(mgl32.DegToRad(45.0), float32(we.Data1)/float32(we.Data2), 0.5, 1000.0)
					this.oglControl.SetOrtho2D(we.Data1, we.Data2)
				default:
					//fmt.Printf("Other window event: %d\n", t.Event)
				}
			default:
				//fmt.Printf("Other event: %T\n", t)
				this.msgHandlerMain(t)
			}

		}
		if this.bAppActive {
			this.UpdateTimer()
			this.oglControl.Render()
		} else {
			sdl.Delay(uint32(100))
		}

	}

	//gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	//gl.UseProgram(program)
	//// 其他渲染代码...
	//
	//window.SwapBuffers()

	//glfw.PollEvents()

}

func (this *COpenGLWinApp) Shutdown() {
	sdl.GLDeleteContext(this.oglControl.SdlContext)
	this.oglControl.SdlWindow.Destroy()
	sdl.Quit()
}

func (this *COpenGLWinApp) msgHandlerMain(event sdl.Event) {
	//var ps PAINTSTRUCT
	//
	//switch t := event.(type) {
	//case sdl.RenderEvent:
	//case sdl.QuitEvent:
	//case sdl.ActiveEvent:
	// case WM_PAINT:
	//	 BeginPaint(hWnd, &ps);
	//	 EndPaint(hWnd, &ps);
	//	 break;
	//
	// case WM_CLOSE:
	//	 PostQuitMessage(0);
	//	 break;
	//
	// case WM_ACTIVATE:
	//	 {
	//		 switch(LOWORD(wParam))
	//		 {
	//		 case WA_ACTIVE:
	//		 case WA_CLICKACTIVE:
	//			 bAppActive = true;
	//			 ResetTimer();
	//			 break;
	//		 case WA_INACTIVE:
	//			 bAppActive = false;
	//			 break;
	//		 }
	//		 break;
	//	 }
	//
	// case WM_SIZE:
	//	 oglControl.ResizeOpenGLViewportFull();
	//	 oglControl.SetProjection3D(45.0f, float(LOWORD(lParam))/float(HIWORD(lParam)), 0.5f, 1000.0f);
	//	 oglControl.SetOrtho2D(LOWORD(lParam), HIWORD(lParam));
	//	 break;
	//
	// default:
	//	 return DefWindowProc(hWnd, uiMsg, wParam, lParam);
	///}
	//return 0;
}

func (this *COpenGLWinApp) GetInstance() {

}
