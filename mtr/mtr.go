package main

import (
	"commlib"
	"taskparser"
	"taskrun"
)

func main() {
	// Init logger
	commlib.Mtrloggger, _ = commlib.InitLogger()
	commlib.Mtrloggger.Println("[INFO] Multiple TaskS Running Tool Start")

	// Parse case info from both terminal and file
	taskparser.ParserParams()
	// Execute tasks
	taskrun.RunTaskMultiple(taskparser.S_P_S_MAP)
	commlib.Mtrloggger.Println("[INFO] Multiple TaskS Running Tool End")
}
