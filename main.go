package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"image/color"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"fyne.io/fyne/v2/theme"
)

//go:generate go-bindata -o ffmpeg_bindata.go ffmpeg/ffmpeg ffmpeg/ffmpeg.exe

func extractFFmpeg() (string, error) {
    var data []byte
    var err error
    var tmpFile *os.File

    switch runtime.GOOS {
    case "windows":
        data, err = Asset("ffmpeg/ffmpeg.exe")
        if err != nil {
            return "", err
        }
        tmpFile, err = os.CreateTemp("", "ffmpeg-*.exe")
    case "darwin", "linux":
        data, err = Asset("ffmpeg/ffmpeg")
        if err != nil {
            return "", err
        }
        tmpFile, err = os.CreateTemp("", "ffmpeg-*")
    default:
        return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
    }

    if err != nil {
        return "", err
    }

    if _, err := tmpFile.Write(data); err != nil {
        return "", err
    }

    if err := tmpFile.Chmod(0755); err != nil {
        return "", err
    }

    return tmpFile.Name(), nil
}

func convertMP4ToGIF(inputFile, outputFile string) error {
    ffmpegPath, err := extractFFmpeg()
    if err != nil {
        return err
    }
    defer os.Remove(ffmpegPath)

    palette := "palette.png"
    filters := "fps=15,scale=640:-1:flags=lanczos"

    // Generate the palette
    cmd := exec.Command(ffmpegPath, "-i", inputFile, "-vf", filters+",palettegen=stats_mode=diff", "-y", palette)
    if err := cmd.Run(); err != nil {
        return err
    }

    // Use the palette to create the GIF
    cmd = exec.Command(ffmpegPath, "-i", inputFile, "-i", palette, "-lavfi", filters+",paletteuse=dither=bayer:bayer_scale=5:diff_mode=rectangle", "-y", outputFile)
    if err := cmd.Run(); err != nil {
        return err
    }

    // Clean up the palette file
    os.Remove(palette)

    return nil
}

func main() {
	gui := app.NewWithID("com.example.mp4togif")
	window := gui.NewWindow("mp4 to gif")
	window.Resize(fyne.NewSize(800,600))

	windowBackground := canvas.NewRectangle(color.RGBA{255, 0, 0, 0})
	windowBackground.SetMinSize(fyne.NewSize(800,600))

	contentBackground := canvas.NewRectangle(color.RGBA{255, 0, 0, 0})
	contentBackground.SetMinSize(fyne.NewSize(600,220))
	contentBackground.StrokeColor = color.RGBA{128, 128, 128, 255}
	contentBackground.StrokeWidth = 1
	contentBackground.CornerRadius = 10

	fileFilter := storage.NewExtensionFileFilter([]string{".mp4", ".mov"})

	heading := canvas.NewText("MP4 to GIF Converter", theme.ForegroundColor())
	heading.Alignment = fyne.TextAlignCenter
	heading.TextStyle = fyne.TextStyle{Bold: true}
	heading.TextSize = 24
	
	spacer := canvas.NewRectangle(theme.BackgroundColor())
	spacer.SetMinSize(fyne.NewSize(0, 10))

	fileLabel := widget.NewLabel("No file selected")
	fileLabel.Alignment = fyne.TextAlignCenter

	//fileButton := widget.NewButton("Select File", func() {
	fileButton := &widget.Button{
		Text: "Select File",
		Importance: widget.HighImportance,
		OnTapped: func(){
			fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
				if err != nil {
					dialog.ShowError(err, window)
					return
				}
				if reader == nil {
					return
				}
				fileLabel.SetText(reader.URI().Path())
			}, window)
			fileDialog.SetFilter(fileFilter)
			fileDialog.Show()
		},
	}
	
	//loading := widget.NewActivity()
	progressBar := widget.NewProgressBarInfinite()
	progressBar.Stop()

	var convertButton *widget.Button
	
	//convertButton = widget.NewButton("Convert file", func() {
	convertButton = &widget.Button{
		Text: "Convert File",
		Importance: widget.HighImportance,
		OnTapped: func(){
			if fileLabel.Text == "" || fileLabel.Text == "No file selected" {
				dialog.ShowInformation("Error", "No file selected", window)
				return
			}

			inputFile := fileLabel.Text
			outputFile := strings.TrimSuffix(inputFile, filepath.Ext(inputFile)) + ".gif"
			
			progressBar.Start()
			progressBar.Refresh()
			convertButton.Disable()
			convertButton.SetText("Converting...")
			//loading.Start()
			//loading.Show()
			go func() {
				err := convertMP4ToGIF(inputFile, outputFile)
				if err != nil {
					dialog.ShowError(err, window)
					fmt.Println("Error converting file:", err)
				} else {
					dialog.ShowInformation("Conversion Successful", "The file has been successfully converted to GIF.", window)
					fmt.Println("Conversion successful!")
				}
				//loading.Stop()
				//loading.Hide()
				progressBar.Stop()
				convertButton.Enable()
				convertButton.SetText("Convert file")
				fileLabel.SetText("No file selected")
			}()
		},
	}

	content := container.New(
		layout.NewStackLayout(),
		windowBackground,
		container.New(
			layout.NewCenterLayout(),
			container.NewStack(
				contentBackground,
				container.NewVBox(
					layout.NewSpacer(), // Add a spacer to push the content down
					container.NewPadded(
						container.NewVBox(
							heading,
							spacer,
							fileButton,
							fileLabel,
							convertButton,
							progressBar,
						),
					),
					layout.NewSpacer(), // Add another spacer to push the content up
				),
			),
		),
	)
	window.SetContent(content)

    window.ShowAndRun()
}
