package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// অ্যাপ স্ট্রাকচারের একটি ইন্সট্যান্স তৈরি
	app := NewApp()

	// Wails অ্যাপ্লিকেশন রান করা
	err := wails.Run(&options.App{
		Title:  "My Go Player", // আপনি চাইলে নাম চেঞ্জ করতে পারেন
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		
		// লাইফসাইকেল মেথডসমূহ
		OnStartup:  app.startup,  // VLC ইঞ্জিন স্টার্ট করবে
		OnShutdown: app.shutdown, // VLC ইঞ্জিন এবং মেমোরি ক্লিন করবে (এটি আমি অ্যাড করেছি)
		
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}