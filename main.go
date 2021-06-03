/*
CropDroid: Automated Farming & Cultivation
Copyright (C) 2019 Automate The Things, LLC
License: Proprietary
*/
package main

import (
	"github.com/jeremyhahn/cropdroid/app"
	"github.com/jeremyhahn/cropdroid/cmd"
)

func main() {
	cmd.App = app.NewApp()
	cmd.Execute()
}
