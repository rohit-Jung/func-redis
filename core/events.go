package core

func Shutdown() {
	// close your file // flush your socket etc
	evalRewriteAOF()
}
