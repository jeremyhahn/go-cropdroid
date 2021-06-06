/*
CropDroid: Automated Farming & Cultivation
Copyright (C) 2019 Automate The Things, LLC
License: Proprietary
*/
package main

import (
	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/cmd"
)

func main() {
	cmd.App = app.NewApp()
	cmd.Execute()
}
