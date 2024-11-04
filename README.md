# mp4-to-gif
mp4 to gif converter written in go


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
go: added fyne.io/fyne/v2 v2.5.2

└─$ go version
go version go1.23.2 darwin/arm64
