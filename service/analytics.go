package service

import (
	"log"
	"os/exec"
)

type Direction int

const (
	Stay Direction = 0
	Up   Direction = 1
	Down Direction = -1
)

func getDirection() Direction {
	app := "Rscript"
	// TODO
	cmd := exec.Command(app, "--vanilla", "../../scripts/la1.R", "3m", "BTCUSDT")
	// cmd := exec.Command(app, "--vanilla", "E:/rstudio/test.r", "3m", "BTCUSDT")

	output, err := cmd.Output()
	if err != nil {
		log.Println("os exec output", err)
		return Stay
	}
	dir := string(output)
	switch dir {
	case "-1":
		return Down
	case "1":
		return Up
	case "0":
		return Stay
	}
	return Stay
}
