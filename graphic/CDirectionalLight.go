package graphic

import "github.com/go-gl/mathgl/mgl32"

type CDirectionalLight struct {
	vColor     mgl32.Vec3 // Color of directional light
	vDirection mgl32.Vec3 // and its direction

	fAmbient float32
}

func NewCDirectionalLight() *CDirectionalLight {
	this := CDirectionalLight{}
	this.vColor = mgl32.Vec3{1.0, 1.0, 1.0}
	this.vDirection = mgl32.Vec3{0.0, -1.0, 0.0}

	this.fAmbient = 0.25
	return &this
}

func NewCDirectionalLightEx(a_vColor, a_vDirection mgl32.Vec3, a_fAmbient float32) *CDirectionalLight {
	this := CDirectionalLight{}
	this.vColor = a_vColor
	this.vDirection = a_vDirection

	this.fAmbient = a_fAmbient
	return &this
}

/*-----------------------------------------------

  Name:	SetUniformData

  Params:	spProgram - shader program
  		sLightVarName - name of directional light variable

  Result:	Sets all directional light data.

  /*---------------------------------------------*/

func (this *CDirectionalLight) SetUniformData(spProgram *CShaderProgram, sLightVarName string) {
	spProgram.SetUniformV3N(sLightVarName+".vColor", &this.vColor, 1)
	spProgram.SetUniformV3N(sLightVarName+".vDirection", &this.vDirection, 1)

	spProgram.SetUniformF32N(sLightVarName+".fAmbient", &this.fAmbient, 1)
}
