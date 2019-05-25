package main

import (
	"github.com/kfur/subtitler/app"
	"github.com/kfur/subtitler/app/route"
	"github.com/kfur/subtitler/app/shared/jsonconfig"
	"github.com/kfur/subtitler/app/shared/recaptcha"
	"github.com/kfur/subtitler/app/shared/server"
	"github.com/kfur/subtitler/app/shared/view"
	"github.com/kfur/subtitler/app/shared/view/plugin"
	c "github.com/kfur/subtitler/config"
	"log"
	"os"
	"runtime"
)

// *****************************************************************************
// Application Logic
// *****************************************************************************

func init() {
	// Verbose logging with file name and line number
	log.SetFlags(log.Lshortfile)

	// Use all CPU cores
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Load the configuration file
	jsonconfig.Load("config"+string(os.PathSeparator)+"config.json", c.Config)
}



func main() {

	// Configure the Google reCAPTCHA prior to loading view plugins
	recaptcha.Configure(c.Config.Recaptcha)

	// Init IBM text to speech service
	app.InitRecogniseService()

	// Setup the views
	view.Configure(c.Config.View)
	view.LoadTemplates(c.Config.Template.Root, c.Config.Template.Children)
	view.LoadPlugins(
		plugin.TagHelper(c.Config.View),
		plugin.NoEscape(),
		plugin.PrettyTime(),
		recaptcha.Plugin())

	// Start the listener
	server.Run(route.LoadHTTP(), route.LoadHTTPS(), c.Config.Server)
}
