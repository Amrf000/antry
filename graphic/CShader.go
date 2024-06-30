package graphic

import (
	"bufio"
	"fmt"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type CShader struct {
	uiShader uint32 // ID of shader
	iType    uint32 // GL_VERTEX_SHADER, GL_FRAGMENT_SHADER...
	bLoaded  bool   // Whether shader was loaded and compiled
}

type CShaderProgram struct {
	uiProgram uint32 // ID of program
	bLinked   bool   // Whether program was linked and is ready to use
}

const NUMSHADERS = 6

func NewCShader() *CShader {
	this := CShader{}
	this.bLoaded = false
	return &this
}

var shShaders [NUMSHADERS]CShader
var spMain, spOrtho2D, spFont2D CShaderProgram

func PrepareShaderPrograms() bool {
	// Load shaders and create shader program

	var sShaderFileNames []string = []string{"main_shader.vert", "main_shader.frag", "ortho2D.vert",
		"ortho2D.frag", "font2D.frag", "dirLight.frag",
	}

	for i := 0; i < NUMSHADERS; i++ {
		var sExt string = filepath.Ext(sShaderFileNames[i])
		var iShaderType uint32 = 0
		if sExt == ".vert" {
			iShaderType = gl.VERTEX_SHADER
		} else {
			if sExt == ".frag" {
				iShaderType = gl.FRAGMENT_SHADER
			} else {
				iShaderType = gl.GEOMETRY_SHADER
			}
		}
		shShaders[i].LoadShader("data\\shaders\\"+sShaderFileNames[i], iShaderType)
	}

	// Create shader programs

	spMain.CreateProgram()
	spMain.AddShaderToProgram(&shShaders[0])
	spMain.AddShaderToProgram(&shShaders[1])
	spMain.AddShaderToProgram(&shShaders[5])

	if !spMain.LinkProgram() {
		panic("spMain.LinkProgram")
		return false
	}

	spOrtho2D.CreateProgram()
	spOrtho2D.AddShaderToProgram(&shShaders[3])
	spOrtho2D.AddShaderToProgram(&shShaders[3])
	if !spOrtho2D.LinkProgram() {
		panic("spOrtho2D.LinkProgram")
		return false
	}

	spFont2D.CreateProgram()
	spFont2D.AddShaderToProgram(&shShaders[2])
	spFont2D.AddShaderToProgram(&shShaders[4])
	if !spFont2D.LinkProgram() {
		panic("spFont2D.LinkProgram")
		return false
	}

	return true
}

//	func loadShaderSource(filename string) (string, error) {
//		data, err := os.ReadFile(filename)
//		if err != nil {
//			return "", err
//		}
//
//		source := string(data)
//		includeRegex := regexp.MustCompile(`#include "(.*)"`)
//
//		matches := includeRegex.FindAllStringSubmatch(source, -1)
//		for _, match := range matches {
//			includeFile := match[1]
//			includeSource, err := loadShaderSource(includeFile)
//			if err != nil {
//				return "", err
//			}
//
//			includeStatement := match[0]
//			source = strings.Replace(source, includeStatement, includeSource, 1)
//		}
//
//		return source, nil
//	}
func (this *CShader) LoadShader(sFile string, a_iType uint32) bool {
	//src, err := loadShaderSource(sFile)
	//if err != nil {
	//	return false
	//}
	var sLines []string
	//
	if !this.GetLinesFromFile(sFile, false, &sLines) {
		panic("err")
		return false
	}
	sstr := string(strings.Join(sLines, "\n"))
	//fmt.Println(sstr)
	//fmt.Println("################################")
	glSrcs, freeFn := gl.Strs(sstr + "\x00")
	defer freeFn()
	//sProgram := make([][]uint8, len(sLines))
	//for i := 0; i < len(sLines); i++ {
	//	sProgram[i] = sLines[i]
	//}

	this.uiShader = gl.CreateShader(a_iType)

	gl.ShaderSource(this.uiShader, 1, glSrcs, nil) //int32(len(sLines)) (*string)(unsafe.Pointer(&sProgram[0]))
	gl.CompileShader(this.uiShader)

	//sProgram = nil

	var iCompilationStatus int32
	gl.GetShaderiv(this.uiShader, gl.COMPILE_STATUS, &iCompilationStatus)

	if iCompilationStatus == gl.FALSE {
		sInfoLog := make([]uint8, 1024)
		var sFinalMessage string
		var iLogLength int32
		gl.GetShaderInfoLog(this.uiShader, 1024, &iLogLength, &sInfoLog[0])
		sFinalMessage = fmt.Sprintf("Error! Shader file %s wasn't compiled! The compiler returned:\n\n%s", sFile, sInfoLog)
		panic(sFinalMessage)
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
		//	Title:       "Error",
		//	Message:     sFinalMessage,
		//	Buttons:     buttons,
		//	ColorScheme: &colorScheme})
		return false
	}
	this.iType = a_iType
	this.bLoaded = true

	return true
}

func (this *CShader) GetLinesFromFile(sFile string, bIncludePart bool, vResult *[]string) bool {
	fp, err := os.Open(sFile) //, "rt"
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	defer fp.Close()

	var sDirectory string
	sDirectory = filepath.Dir(sFile)

	// Get all lines from a file

	var bInIncludePart bool = false

	fileScanner := bufio.NewScanner(fp)

	fileScanner.Split(bufio.ScanLines)

	for fileScanner.Scan() {
		sLine := fileScanner.Text()
		fields := strings.Fields(sLine)
		if len(fields) > 0 {
			if fields[0] == "#include" {
				var sFileName string = fields[1]
				if len(sFileName) > 0 && strings.HasPrefix(sFileName, "\"") && strings.HasSuffix(sFileName, "\"") {
					sFileName = strings.Replace(sFileName, "\"", "", -1)
					this.GetLinesFromFile(path.Join(sDirectory, sFileName), true, vResult)
				}

			} else if fields[0] == "#include_part" {
				bInIncludePart = true
			} else if fields[0] == "#definition_part" {
				bInIncludePart = false
			} else if !bIncludePart || (bIncludePart && bInIncludePart) {
				*vResult = append(*vResult, sLine)
			}
		}
	}

	return true
}

func (this *CShader) IsLoaded() bool {
	return this.bLoaded
}

func (this *CShader) GetShaderID() uint32 {
	return this.uiShader
}

func (this *CShader) DeleteShader() {
	if !this.IsLoaded() {
		return
	}
	this.bLoaded = false
	gl.DeleteShader(this.uiShader)
}

func NewCShaderProgram() *CShaderProgram {
	this := CShaderProgram{}
	this.bLinked = false
	return &this
}

func (this *CShaderProgram) CreateProgram() {
	this.uiProgram = gl.CreateProgram()
}

func (this *CShaderProgram) AddShaderToProgram(shShader *CShader) bool {
	if !shShader.IsLoaded() {
		return false
	}

	gl.AttachShader(this.uiProgram, shShader.GetShaderID())

	return true
}
func (this *CShaderProgram) LinkProgram() bool {
	gl.LinkProgram(this.uiProgram)
	var iLinkStatus int32
	gl.GetProgramiv(this.uiProgram, gl.LINK_STATUS, &iLinkStatus)
	this.bLinked = iLinkStatus == gl.TRUE
	if iLinkStatus == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(this.uiProgram, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(this.uiProgram, logLength, nil, gl.Str(log))

		fmt.Printf("无法链接程序: %v\n", log)
		panic(log)
	}

	return this.bLinked
}
func (this *CShaderProgram) DeleteProgram() {
	if !this.bLinked {
		return
	}
	this.bLinked = false
	gl.DeleteProgram(this.uiProgram)
}
func (this *CShaderProgram) UseProgram() {
	if this.bLinked {
		gl.UseProgram(this.uiProgram)
	}
}
func (this *CShaderProgram) GetProgramID() uint32 {
	return this.uiProgram
}

func (this *CShaderProgram) SetUniformF32N(sName string, fValues *float32, iCount int32) {
	var iLoc int32 = gl.GetUniformLocation(this.uiProgram, gl.Str(sName+"\x00"))
	gl.Uniform1fv(iLoc, iCount, fValues)
}

func (this *CShaderProgram) SetUniformF32(sName string, fValue float32) {
	var iLoc int32 = gl.GetUniformLocation(this.uiProgram, gl.Str(sName+"\x00"))
	gl.Uniform1fv(iLoc, 1, &fValue)
}

// Setting vectors

func (this *CShaderProgram) SetUniformV2N(sName string, vVectors *mgl32.Vec2, iCount int32) {
	var iLoc int32 = gl.GetUniformLocation(this.uiProgram, gl.Str(sName+"\x00"))
	gl.Uniform2fv(iLoc, iCount, &(*vVectors)[0])
}

func (this *CShaderProgram) SetUniformV2(sName string, vVector mgl32.Vec2) {
	var iLoc int32 = gl.GetUniformLocation(this.uiProgram, gl.Str(sName+"\x00"))
	gl.Uniform2fv(iLoc, 1, &vVector[0])
}

func (this *CShaderProgram) SetUniformV3N(sName string, vVectors *mgl32.Vec3, iCount int32) {
	var iLoc int32 = gl.GetUniformLocation(this.uiProgram, gl.Str(sName+"\x00"))
	gl.Uniform3fv(iLoc, iCount, &(*vVectors)[0])
}

func (this *CShaderProgram) SetUniformV3(sName string, vVector mgl32.Vec3) {
	var iLoc int32 = gl.GetUniformLocation(this.uiProgram, gl.Str(sName+"\x00"))
	gl.Uniform3fv(iLoc, 1, &vVector[0])
}

func (this *CShaderProgram) SetUniformV4N(sName string, vVectors *mgl32.Vec4, iCount int32) {
	var iLoc int32 = gl.GetUniformLocation(this.uiProgram, gl.Str(sName+"\x00"))
	gl.Uniform4fv(iLoc, iCount, &((*vVectors)[0]))
}

func (this *CShaderProgram) SetUniformV4(sName string, vVector mgl32.Vec4) {
	var iLoc int32 = gl.GetUniformLocation(this.uiProgram, gl.Str(sName+"\x00"))
	gl.Uniform4fv(iLoc, 1, &vVector[0])
}

// Setting 3x3 matrices

func (this *CShaderProgram) SetUniformM3N(sName string, mMatrices *mgl32.Mat3, iCount int32) {
	var iLoc int32 = gl.GetUniformLocation(this.uiProgram, gl.Str(sName+"\x00"))
	gl.UniformMatrix3fv(iLoc, iCount, false, &((*mMatrices)[0]))
}

func (this *CShaderProgram) SetUniformM3(sName string, mMatrix mgl32.Mat3) {
	var iLoc int32 = gl.GetUniformLocation(this.uiProgram, gl.Str(sName+"\x00"))
	gl.UniformMatrix3fv(iLoc, 1, false, &mMatrix[0])
}

// Setting 4x4 matrices

func (this *CShaderProgram) SetUniformM4N(sName string, mMatrices *mgl32.Mat4, iCount int32) {
	var iLoc int32 = gl.GetUniformLocation(this.uiProgram, gl.Str(sName+"\x00"))
	gl.UniformMatrix4fv(iLoc, iCount, false, &(*mMatrices)[0])
}

func (this *CShaderProgram) SetUniformM4(sName string, mMatrix mgl32.Mat4) {
	var iLoc int32 = gl.GetUniformLocation(this.uiProgram, gl.Str(sName+"\x00"))
	gl.UniformMatrix4fv(iLoc, 1, false, &mMatrix[0])
}

// Setting integers

func (this *CShaderProgram) SetUniformI32N(sName string, iValues *int32, iCount int32) {
	var iLoc int32 = gl.GetUniformLocation(this.uiProgram, gl.Str(sName+"\x00"))
	gl.Uniform1iv(iLoc, iCount, iValues)
}

func (this *CShaderProgram) SetUniformI32(sName string, iValue int32) {
	var iLoc int32 = gl.GetUniformLocation(this.uiProgram, gl.Str(sName+"\x00"))
	gl.Uniform1i(iLoc, iValue)
}

func (this *CShaderProgram) SetModelAndNormalMatrix(sModelMatrixName, sNormalMatrixName string, mModelMatrix mgl32.Mat4) {
	this.SetUniformM4(sModelMatrixName, mModelMatrix)
	this.SetUniformM4(sNormalMatrixName, mModelMatrix.Inv().Transpose())
}

func (this *CShaderProgram) SetModelAndNormalMatrixP(sModelMatrixName, sNormalMatrixName string, mModelMatrix *mgl32.Mat4) {
	this.SetUniformM4N(sModelMatrixName, mModelMatrix, 1)
	this.SetUniformM4(sNormalMatrixName, mModelMatrix.Inv().Transpose())
}
