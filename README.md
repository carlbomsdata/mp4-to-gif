# mp4-to-gif converter application
This desktop application converts mp4 files in to gif files, simple as that.
It is written in Go and uses the Fyne GUI toolkit.

![Skärminspelning 2024-11-04 kl  18 37 19](https://github.com/user-attachments/assets/d9602ff0-ae4b-4fb9-88a3-01c7c3d84907)

## Dependencies
After cloning the repo, download and place ffmpeg executable inside the folder ffmpeg.
https://www.ffmpeg.org/download.html

On MacOS run:
```bash
go-bindata -o ffmpeg_bindata_unix.go -tags=unix ffmpeg/ffmpeg
```
On Windows run:
```bash
go-bindata -o ffmpeg_bindata_windows.go -tags=windows ffmpeg/ffmpeg.exe
```
```bash
go mod tidy
```
```bash
go get fyne.io/fyne/v2@latest
```

## Package
```bash
go install fyne.io/fyne/v2/cmd/fyne@latest
```
```bash
fyne package -os darwin -icon icon.png
```
```bash
fyne package -os windows -icon icon.png
```
