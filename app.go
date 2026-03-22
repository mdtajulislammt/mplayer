package main

/*
#cgo CFLAGS: -I"C:/Users/anoth/Downloads/vlc-3.0.23-win64/vlc-3.0.23/sdk/include"
#cgo LDFLAGS: -L"C:/Users/anoth/Downloads/vlc-3.0.23-win64/vlc-3.0.23/sdk/lib" -lvlc
#include <vlc/vlc.h>
#include <windows.h>

// VLC কে একটি উইন্ডোতে সেট করার ফাংশন
void set_player_hwnd(libvlc_media_player_t *player, HWND hwnd) {
    libvlc_media_player_set_hwnd(player, hwnd);
}

// Wide string এরর ফিক্স করার জন্য হেল্পার
LPCWSTR get_static_class() {
    return L"Static";
}
*/
import "C"

import (
	"context"
	"fmt"
	"reflect"
	"unsafe"

	vlc "github.com/adrg/libvlc-go/v3"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type MediaInfo struct {
	Title    string `json:"title"`
	Duration int64  `json:"duration"`
}

type App struct {
	ctx       context.Context
	player    *vlc.Player
	vlcHwnd   C.HWND
	wailsHwnd C.HWND
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	vlc.Init("--no-video-title-show", "--quiet")
	
	// Wails উইন্ডোর HWND ডিটেক্ট করা
	runtime.EventsOn(ctx, "wails:window:wndproc", func(data ...interface{}) {
		if len(data) > 0 {
			a.wailsHwnd = C.HWND(unsafe.Pointer(uintptr(data[0].(int))))
		}
	})
}

func (a *App) shutdown(ctx context.Context) {
	if a.player != nil {
		a.player.Release()
	}
	vlc.Release()
}

// রিফ্লেকশন ব্যবহার করে ইন্টারনাল প্লেয়ার পয়েন্টার বের করা
func (a *App) getInternalPlayer() unsafe.Pointer {
	if a.player == nil { return nil }
	val := reflect.ValueOf(a.player).Elem()
	field := val.FieldByName("player") 
	return unsafe.Pointer(field.Pointer())
}

func (a *App) SelectAndPlay() (MediaInfo, error) {
	file, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Video",
		Filters: []runtime.FileFilter{
			{DisplayName: "Video Files", Pattern: "*.mp4;*.mkv;*.avi"},
		},
	})

	if err != nil || file == "" {
		return MediaInfo{}, fmt.Errorf("selection canceled")
	}

	// ১. আগের প্লেয়ার এবং উইন্ডো ক্লিনআপ (মেমোরি সেফটি)
	if a.player != nil {
		a.player.Stop()
		a.player.Release()
		a.player = nil
	}
	if a.vlcHwnd != nil {
		C.DestroyWindow(a.vlcHwnd) // আগের চাইল্ড উইন্ডো মুছে ফেলা
		a.vlcHwnd = nil
	}

	// ২. নতুন প্লেয়ার তৈরি
	player, err := vlc.NewPlayer()
	if err != nil {
		return MediaInfo{}, err
	}
	a.player = player

	// ৩. চাইল্ড উইন্ডো তৈরি (সঠিক সাইজ সহ)
	w, h := runtime.WindowGetSize(a.ctx)
	className := C.get_static_class()
	
	a.vlcHwnd = C.CreateWindowExW(
		0, 
		className, 
		nil,
		C.WS_CHILD|C.WS_VISIBLE|C.WS_CLIPSIBLINGS,
		0, 0, C.int(w), C.int(h-100), // শুরুতে উইন্ডো সাইজ সেট করা
		a.wailsHwnd, 
		nil, nil, nil,
	)

	// ৪. VLC কে উইন্ডো হ্যান্ডেল দেওয়া
	internalPlayer := (*C.libvlc_media_player_t)(a.getInternalPlayer())
	if internalPlayer != nil {
		C.set_player_hwnd(internalPlayer, a.vlcHwnd)
	}

	// ৫. মিডিয়া লোড এবং প্লে
	media, err := a.player.LoadMediaFromPath(file)
	if err != nil {
		return MediaInfo{}, err
	}
	defer media.Release()

	err = a.player.Play()
	if err != nil {
		return MediaInfo{}, err
	}
	
	// ৬. মিডিয়া ডিউরেশন নেওয়া (Multiple returns handled)
	d, _ := media.Duration()

	return MediaInfo{
		Title:    file,
		Duration: int64(d),
	}, nil
}

// রিসাইজ লজিক: এটি ফ্রন্টএন্ড থেকে window.onresize এ কল করা উচিত
func (a *App) ResizeVideoWindow() {
	if a.vlcHwnd != nil && a.ctx != nil {
		w, h := runtime.WindowGetSize(a.ctx)
		// UI এর কন্ট্রোল বারের জন্য ১০০ পিক্সেল বাদ রাখা হয়েছে
		C.MoveWindow(a.vlcHwnd, 0, 0, C.int(w), C.int(h-100), C.TRUE)
	}
}

// --- Frontend APIs ---

func (a *App) TogglePlay() bool {
	if a.player != nil {
		a.player.TogglePause()
		return a.player.IsPlaying()
	}
	return false
}

func (a *App) SetVolume(vol int) {
	if a.player != nil {
		a.player.SetVolume(vol)
	}
}

func (a *App) Seek(pos float32) {
	if a.player != nil {
		a.player.SetMediaPosition(pos)
	}
}

func (a *App) GetPlaybackStatus() (int64, int64) {
	if a.player == nil {
		return 0, 0
	}

	// ১. বর্তমান সময় পাওয়ার জন্য (ms)
	t, _ := a.player.MediaTime()

	// ২. ডিউরেশন পাওয়ার জন্য আমাদের বর্তমান মিডিয়াটি নিতে হবে
	media, err := a.player.Media()
	if err != nil || media == nil {
		return int64(t), 0
	}
	
	// মিডিয়া থেকে ডিউরেশন নেওয়া
	d, _ := media.Duration()
	
	return int64(t), int64(d)
}