package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"
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

func convertMP4ToGIF(ctx context.Context, inputFile, outputFile string) error {
	ffmpegPath, err := extractFFmpeg()
	if err != nil {
		return err
	}
	defer os.Remove(ffmpegPath)

	palette := "palette.png"
	filters := "fps=15,scale=640:-1:flags=lanczos"

	// Generate the palette with cancellation check
	cmd := exec.CommandContext(ctx, ffmpegPath, "-i", inputFile, "-vf", filters+",palettegen=stats_mode=diff", "-y", palette)
	if err := cmd.Run(); err != nil {
		return err
	}

	// Use the palette to create the GIF with cancellation check
	cmd = exec.CommandContext(ctx, ffmpegPath, "-i", inputFile, "-i", palette, "-lavfi", filters+",paletteuse=dither=bayer:bayer_scale=5:diff_mode=rectangle", "-y", outputFile)
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
	window.Resize(fyne.NewSize(800, 600))

	windowBackground := canvas.NewRectangle(color.RGBA{255, 0, 0, 0})
	windowBackground.SetMinSize(fyne.NewSize(800, 600))

	contentBackground := canvas.NewRectangle(color.RGBA{255, 0, 0, 0})
	contentBackground.SetMinSize(fyne.NewSize(600, 220))
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

	fileButton := &widget.Button{
		Text:       "Select File",
		Importance: widget.HighImportance,
		OnTapped: func() {
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

	progressBar := widget.NewProgressBar()
	progressBar.Min = 0
	progressBar.Max = 100
	progressBar.SetValue(0)

	var (
		ctx              context.Context
		cancel           context.CancelFunc
		convertButton    *widget.Button
		convertButtonState int
	)

	convertButton = &widget.Button{
		Text:       "Convert File",
		Importance: widget.HighImportance,
		OnTapped: func() {
			if convertButtonState == 1 { // If already converting, cancel it
				cancel()
				convertButton.SetText("Convert File")
				convertButton.Importance = widget.HighImportance
				convertButtonState = 0
				return
			}

			if fileLabel.Text == "" || fileLabel.Text == "No file selected" {
				dialog.ShowInformation("Error", "No file selected", window)
				return
			}

			// Set up context with cancel function
			ctx, cancel = context.WithCancel(context.Background())

			inputFile := fileLabel.Text
			outputFile := strings.TrimSuffix(inputFile, filepath.Ext(inputFile)) + ".gif"

			progressBar.SetValue(0)
			convertButton.SetText("Cancel")
			convertButton.Importance = widget.DangerImportance
			convertButtonState = 1

			// Run the progress bar in a separate goroutine
			go func() {
				startTime := time.Now()
				ticker := time.NewTicker(100 * time.Millisecond)
				defer ticker.Stop()

				fileInfo, err := os.Stat(inputFile)
				if err != nil {
					fmt.Println("Error:", err)
					return
				}
				inputFileSize := fileInfo.Size()
				estimatedTime := time.Duration(inputFileSize*30/22_000_000) * time.Second

				for range ticker.C {
					if convertButtonState == 0 {
						progressBar.SetValue(0) // Reset if canceled
						return
					}

					elapsed := time.Since(startTime)
					progress := float64(elapsed) / float64(estimatedTime)
					if progress >= 1.0 {
						progressBar.SetValue(100)
						break
					}
					progressBar.SetValue(progress * 100)
				}
			}()

			// Run the conversion in another goroutine
			go func() {
				err := convertMP4ToGIF(ctx, inputFile, outputFile)

				if err != nil {
					// Check if the error is due to context cancellation
					if ctx.Err() == context.Canceled {
						dialog.ShowInformation("Conversion Canceled", "The conversion was canceled.", window)
					} else {
						dialog.ShowError(err, window) // Show other errors normally
					}
				} else {
					convertButton.Importance = widget.HighImportance
					dialog.ShowInformation("Conversion Successful", "The file has been successfully converted to GIF.", window)
				}

				// Reset UI elements
				progressBar.SetValue(0)
				convertButton.SetText("Convert File")
				convertButton.Importance = widget.HighImportance
				fileLabel.SetText("No file selected")
				convertButtonState = 0
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
					layout.NewSpacer(),
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
					layout.NewSpacer(),
				),
			),
		),
	)
	window.SetContent(content)

	window.ShowAndRun()
}
