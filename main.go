package main

import "antry/graphic"

func main() {
	if !graphic.AppMain.InitializeApp("21_opengl_3_3") {
		return
	}

	graphic.AppMain.RegisterAppClass()

	if !graphic.AppMain.CreateAppWindow("21.) Multilayered Terrain - Tutorial by Michal Bubnar (www.mbsoftworks.sk)") {
		return
	}
	graphic.AppMain.ResetTimer()

	graphic.AppMain.AppBody()
	graphic.AppMain.Shutdown()
}
