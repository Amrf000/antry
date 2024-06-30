package graphic

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/veandco/go-sdl2/sdl"
	"math"
)

type CFlyingCamera struct {
	vEye, vView, vUp mgl32.Vec3
	fSpeed           float32
	fSensitivity     float32 // How many degrees to rotate per pixel moved by mouse (nice value is 0.10)

	pCur                        sdl.Point // For mosue rotation
	iForw, iBack, iLeft, iRight int
}

var PI = float32(math.Atan(1.0) * 4.0)

func NewCFlyingCamera() *CFlyingCamera {
	this := CFlyingCamera{}
	this.vEye = mgl32.Vec3{0.0, 0.0, 0.0}
	this.vView = mgl32.Vec3{0.0, 0.0, -1.0}
	this.vUp = mgl32.Vec3{0.0, 1.0, 0.0}
	this.fSpeed = 25.0
	this.fSensitivity = 0.1
	return &this
}

func NewCFlyingCameraEx(a_vEye, a_vView, a_vUp mgl32.Vec3, a_fSpeed, a_fSensitivity float32) *CFlyingCamera {
	this := CFlyingCamera{}
	this.vEye = a_vEye
	this.vView = a_vView
	this.vUp = a_vUp
	this.fSpeed = a_fSpeed
	this.fSensitivity = a_fSensitivity
	return &this
}

/*-----------------------------------------------

  Name:	rotateWithMouse

  Params:	none

  Result:	Checks for moving of mouse and rotates
  		camera.

  /*---------------------------------------------*/

func (this *CFlyingCamera) RotateWithMouse() {

	pCurX, pCurY, _ := sdl.GetMouseState()
	width, height := AppMain.oglControl.SdlWindow.GetSize()
	var iCentX int32 = width / 2
	var iCentY int32 = height / 2

	var deltaX float32 = (float32)(iCentX-pCurX) * this.fSensitivity
	var deltaY float32 = (float32)(iCentY-pCurY) * this.fSensitivity

	if deltaX != 0.0 {
		this.vView = this.vView.Sub(this.vEye)
		this.vView = mgl32.QuatRotate(deltaX, mgl32.Vec3{0.0, 1.0, 0.0}).Rotate(this.vView)
		this.vView = this.vView.Add(this.vEye)
	}
	if deltaY != 0.0 {
		var vAxis mgl32.Vec3 = this.vView.Sub(this.vEye).Cross(this.vUp)
		vAxis = vAxis.Normalize()
		var fAngle float32 = deltaY
		var fNewAngle float32 = fAngle + this.GetAngleX()
		if fNewAngle > -89.80 && fNewAngle < 89.80 {
			this.vView = this.vView.Sub(this.vEye)
			this.vView = mgl32.QuatRotate(deltaY, vAxis).Rotate(this.vView)
			this.vView = this.vView.Add(this.vEye)
		}
	}

	AppMain.oglControl.SdlWindow.WarpMouseInWindow(iCentX, iCentY)
}

/*-----------------------------------------------

  Name:	GetAngleY

  Params:	none

  Result:	Gets Y angle of camera (head turning left
  		and right).

  /*---------------------------------------------*/

func (this *CFlyingCamera) GetAngleY() float32 {
	var vDir mgl32.Vec3 = this.vView.Sub(this.vEye)
	vDir[1] = 0.0
	vDir.Normalize()
	var fAngle float32 = float32(math.Acos(float64(mgl32.Vec3{0, 0, -1}.Dot(vDir)))) * (180.0 / PI)
	if vDir.X() < 0 {
		fAngle = 360.0 - fAngle
	}
	return fAngle
}

/*-----------------------------------------------

  Name:	GetAngleX

  Params:	none

  Result:	Gets X angle of camera (head turning up
  		and down).

  /*---------------------------------------------*/

func (this *CFlyingCamera) GetAngleX() float32 {
	var vDir mgl32.Vec3 = this.vView.Sub(this.vEye)
	vDir = vDir.Normalize()
	var vDir2 mgl32.Vec3 = vDir
	vDir2[1] = 0.0
	vDir2 = vDir2.Normalize()
	var fAngle float32 = float32(math.Acos(float64(vDir2.Dot(vDir)))) * (180.0 / PI)
	if vDir.Y() < 0 {
		fAngle *= -1.0
	}
	return fAngle
}

/*-----------------------------------------------

  Name:	SetMovingKeys

  Params:	a_iForw - move forward Key
  		a_iBack - move backward Key
  		a_iLeft - strafe left Key
  		a_iRight - strafe right Key

  Result:	Sets Keys for moving camera.

  /*---------------------------------------------*/

func (this *CFlyingCamera) SetMovingKeys(a_iForw, a_iBack, a_iLeft, a_iRight int) {
	this.iForw = a_iForw
	this.iBack = a_iBack
	this.iLeft = a_iLeft
	this.iRight = a_iRight
}

/*-----------------------------------------------

  Name:	Update

  Params:	none

  Result:	Performs updates of camera - moving and
  		rotating.

  /*---------------------------------------------*/

func (this *CFlyingCamera) Update() {
	this.RotateWithMouse()

	// Get view direction
	var vMove mgl32.Vec3 = this.vView.Sub(this.vEye)
	vMove = vMove.Normalize()
	vMove = vMove.Mul(this.fSpeed)

	var vStrafe mgl32.Vec3 = this.vView.Sub(this.vEye).Cross(this.vUp)
	vStrafe = vStrafe.Normalize()
	vStrafe = vStrafe.Mul(this.fSpeed)

	//var iMove int = 0
	var vMoveBy mgl32.Vec3
	// Get vector of move
	keys := sdl.GetKeyboardState()
	if keys[sdl.SCANCODE_UP] != 0 { //(this.iForw)
		vMoveBy = vMoveBy.Add(vMove.Mul(AppMain.sof(1.0)))
	}
	if keys[sdl.SCANCODE_DOWN] != 0 { //Keys::Key(this.iBack)
		vMoveBy = vMoveBy.Sub(vMove.Mul(AppMain.sof(1.0)))
	}
	if keys[sdl.SCANCODE_LEFT] != 0 { //Keys::Key(this.iLeft)
		vMoveBy = vMoveBy.Sub(vStrafe.Mul(AppMain.sof(1.0)))
	}
	if keys[sdl.SCANCODE_RIGHT] != 0 { //Keys::Key(this.iRight)
		vMoveBy = vMoveBy.Add(vStrafe.Mul(AppMain.sof(1.0)))
	}
	this.vEye = this.vEye.Add(vMoveBy)
	this.vView = this.vView.Add(vMoveBy)
}

/*-----------------------------------------------

  Name:	ResetMouse

  Params:	none

  Result:	Sets mouse cursor back to the center of
  		window.

  /*---------------------------------------------*/

func (this *CFlyingCamera) ResetMouse() {
	width, height := AppMain.oglControl.SdlWindow.GetSize()
	var iCentX int32 = width / 2
	var iCentY int32 = height / 2
	AppMain.oglControl.SdlWindow.WarpMouseInWindow(iCentX, iCentY)
}

/*-----------------------------------------------

  Name:	Look

  Params:	none

  Result:	Returns proper modelview matrix to make
  		camera look.

  /*---------------------------------------------*/

func (this *CFlyingCamera) Look() mgl32.Mat4 {
	return mgl32.LookAtV(this.vEye, this.vView, this.vUp)
}
