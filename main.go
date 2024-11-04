package main

import (
    "fmt"
    "os"
    "os/exec"
    "runtime"
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
    inputFile := "input.mov"
    outputFile := "output.gif"
    
    err := convertMP4ToGIF(inputFile, outputFile)
    if err != nil {
        fmt.Println("Error converting file:", err)
    } else {
        fmt.Println("Conversion successful!")
    }
}
