# mp4-to-gif converter application
This gui application converts mp4 files in to gif files, simple as that.
It is written in Go and uses the Fyne GUI toolkit.



https://github.com/user-attachments/assets/a842ccb0-f3d1-4245-950f-0b98651701d0



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
