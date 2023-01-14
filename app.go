package main

import (
    "context"
    "encoding/json"
    "fmt"
    "github.com/Gurkengewuerz/goffmpeg/transcoder"
    "github.com/mitchellh/mapstructure"
    "github.com/wailsapp/wails/v2/pkg/runtime"
    "os"
    "path/filepath"
    "strings"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Shutdown is called at application termination
func (a *App) Shutdown(ctx context.Context) {
	// Perform your teardown here
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

type ConvertFile struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Type string `json:"type"`
}

func (a *App) SelectFiles() {
	res, err := runtime.OpenMultipleFilesDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Files",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Images/Videos (*.png;*.jpg;*.mov;*.mp4)",
				Pattern:     "*.png;*.jpg;*.mov;*.mp4",
			},
		},
		ShowHiddenFiles:            true,
		CanCreateDirectories:       false,
		ResolvesAliases:            true,
		TreatPackagesAsDirectories: false,
	})
	if err != nil {
		return
	}

	var files []ConvertFile
	for _, file := range res {
		var fileType string
		filePathLower := strings.ToLower(file)
		if strings.HasSuffix(filePathLower, ".mp4") {
			fileType = "vid"
		} else if strings.HasSuffix(filePathLower, ".png") || strings.HasSuffix(filePathLower, ".jpg") || strings.HasSuffix(filePathLower, ".jpeg") {
			fileType = "img"
		}

		if fileType == "" {
			continue
		}

		files = append(files, ConvertFile{
			Path: file,
			Name: filepath.Base(file),
			Type: fileType,
		})
	}

	if len(files) == 0 {
		return
	}

	marshaled, err := json.Marshal(files)
	if err != nil {
		return
	}

	runtime.EventsEmit(a.ctx, "selectFiles", string(marshaled))
}

func (a *App) SelectWatermark() {
	res, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Watermark",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Watermark (*.png;*.jpg)",
				Pattern:     "*.png;*.jpg",
			},
		},
		ShowHiddenFiles:            true,
		CanCreateDirectories:       false,
		ResolvesAliases:            true,
		TreatPackagesAsDirectories: false,
	})
	if err != nil {
		return
	}
	runtime.EventsEmit(a.ctx, "selectWatermark", res)
}

func (a *App) SelectOutputFolder() {
	res, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title:                      "Select Output Folder",
		CanCreateDirectories:       true,
		ResolvesAliases:            true,
		TreatPackagesAsDirectories: false,
	})
	if err != nil {
		return
	}
	runtime.EventsEmit(a.ctx, "selectOutputFolder", res)
}

type Progress struct {
	Percentage  int          `json:"percentage"`
	CurrentFile *ConvertFile `json:"currentFile"`
	Running     bool         `json:"running"`
	Error       string       `json:"error"`
}

type ProcessData struct {
	Files        []ConvertFile `json:"files"`
	Transparent  int           `json:"transparent"`
	Size         int           `json:"size"`
	Watermark    string        `json:"watermark"`
	Prefix       string        `json:"prefix"`
	Position     string        `json:"position"`
	OutputFolder string        `json:"outputFolder"`
}

func (a *App) getAppDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	appDir := filepath.Join(configDir, "WatermarkGenerator")
	err = os.MkdirAll(appDir, os.ModePerm)
	if err != nil {
		return "", err
	}
	return appDir, nil
}

func (a *App) writeSettings(data ProcessData) {
	configDir, err := a.getAppDir()
	if err != nil {
		return
	}

	configFile := filepath.Join(configDir, "config.json")
	data.Files = []ConvertFile{}

	file, _ := json.MarshalIndent(data, "", "  ")
	_ = os.WriteFile(configFile, file, 0644)
}

func (a *App) ReadSettings() string {
	configDir, err := a.getAppDir()
	if err != nil {
		return ""
	}

	configFile := filepath.Join(configDir, "config.json")

	file, err := os.ReadFile(configFile)
	if err != nil {
		return ""
	}

	return string(file)
}

func (a *App) reportProgress(progress *Progress) {
	marshaled, err := json.Marshal(progress)
	if err != nil {
		return
	}

	runtime.EventsEmit(a.ctx, "progress", string(marshaled))
}

func (a *App) ProcessData(data interface{}) {
	var myProcessData ProcessData
	err := mapstructure.Decode(data, &myProcessData)
	if err != nil {
		return
	}
	a.writeSettings(myProcessData)

	for idx, file := range myProcessData.Files {
		dir := filepath.Dir(file.Path)
		outName := fmt.Sprintf("%s%s", myProcessData.Prefix, file.Name)
		out := filepath.Join(dir, outName)

		if myProcessData.OutputFolder != "" {
			out = filepath.Join(myProcessData.OutputFolder, outName)
		}

		// Create new instance of transcoder
		trans := new(transcoder.Transcoder)

		// Initialize transcoder passing the input file path and output file path
		err = trans.Initialize(file.Path, out)
		if err != nil {
			a.reportProgress(&Progress{
				Percentage:  100,
				CurrentFile: &file,
				Running:     false,
				Error:       err.Error(),
			})
			return
		}

		var overlay string
		switch myProcessData.Position {
		case "top-left":
			overlay = "5:5"
			break
		case "top-right":
			overlay = "W-w-5:5"
			break
		case "bottom-left":
			overlay = "5:H-h-5"
			break
		default:
		case "bottom-right":
			overlay = "W-w-5:H-h-5"
			break
		}

		// Handle error...
		trans.MediaFile().SetAdditionalInputPath(myProcessData.Watermark)
		// https://stackoverflow.com/questions/10918907/how-to-add-transparent-watermark-in-center-of-a-video-with-ffmpeg/10920872#10920872
		trans.MediaFile().SetFilterComplex(fmt.Sprintf("[1][0]scale2ref=w=oh*mdar:h=ih*%.2f[logo][video];[logo]format=rgba,colorchannelmixer=aa=%.2f[logo];[video][logo]overlay=%s:format=auto,format=yuv420p", float32(myProcessData.Size)/100, float32(100-myProcessData.Transparent)/100, overlay))
		runtime.LogDebugf(a.ctx, "%v\r\n", trans.GetCommand())
		// Start transcoder process without checking progress
		done := trans.Run(false)

		// This channel is used to wait for the process to end
		err = <-done
		// Handle error..
		if err != nil {
			a.reportProgress(&Progress{
				Percentage:  100,
				CurrentFile: &file,
				Running:     false,
				Error:       err.Error(),
			})
			return
		}

		if len(myProcessData.Files) == idx+1 {
			a.reportProgress(&Progress{
				Percentage:  100,
				CurrentFile: &file,
				Running:     false,
				Error:       "",
			})
		} else {
			a.reportProgress(&Progress{
				Percentage:  int(float32(idx+1) / float32(len(myProcessData.Files)) * 100),
				CurrentFile: &file,
				Running:     true,
				Error:       "",
			})
		}

	}
}
