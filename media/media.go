package media

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type MediaProcessor struct {
	ramDiskPath string
}

func NewMediaProcessor(ramDisk string) *MediaProcessor {
	return &MediaProcessor{ramDiskPath: ramDisk}
}

// FormatToStatic converts input (image or video) to a .webp static sticker
func (m *MediaProcessor) FormatToStatic(inputPath, outputPath string) error {
	// Extract first frame and scale
	scaleFilter := "scale='if(gt(iw,ih),512,-1)':'if(gt(iw,ih),-1,512)'"
	cmd := exec.Command("ffmpeg", "-y", "-i", inputPath, "-vframes", "1", "-vf", scaleFilter, "-c:v", "libwebp", outputPath)
	
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg FormatToStatic error: %v, stderr: %s", err, stderr.String())
	}
	return nil
}

// FormatToVideo converts input (image or video) to a .webm video sticker
func (m *MediaProcessor) FormatToVideo(inputPath, outputPath string, isImage bool) error {
	scaleFilter := "scale='if(gt(iw,ih),512,-2)':'if(gt(iw,ih),-2,512)'" // -2 to keep even dimensions

	var cmd *exec.Cmd
	if isImage {
		// Image to 3s video at 1fps
		cmd = exec.Command("ffmpeg", "-y", "-loop", "1", "-i", inputPath, "-t", "3", "-r", "1", "-vf", scaleFilter, "-c:v", "libvpx-vp9", "-b:v", "200k", "-an", outputPath)
	} else {
		// Video to video sticker
		cmd = exec.Command("ffmpeg", "-y", "-i", inputPath, "-t", "3", "-r", "30", "-vf", scaleFilter, "-c:v", "libvpx-vp9", "-b:v", "256k", "-an", outputPath)
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg FormatToVideo error: %v, stderr: %s", err, stderr.String())
	}
	return nil
}

// RemoveBackgroundStatic removes bg from an image
func (m *MediaProcessor) RemoveBackgroundStatic(inputPath, outputPath string) error {
	cmd := exec.Command("python3", "python/rembg_script.py", inputPath, outputPath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("rembg error: %v, stderr: %s", err, stderr.String())
	}
	return nil
}

// RemoveBackgroundVideo processes video, removes bg from frames, and recombines
func (m *MediaProcessor) RemoveBackgroundVideo(inputPath, outputPath string) error {
	// 1. Create tmp dirs in ramdisk
	tmpDir, err := os.MkdirTemp(m.ramDiskPath, "vid_bg_*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	framesDir := filepath.Join(tmpDir, "frames")
	outFramesDir := filepath.Join(tmpDir, "out_frames")
	os.Mkdir(framesDir, 0755)
	os.Mkdir(outFramesDir, 0755)

	// 2. Pre-process video to reduce size/fps to TG limits first to save ML processing time
	// 30fps max, 3 sec max -> 90 frames max
	preProcessed := filepath.Join(tmpDir, "pre.webm")
	if err := m.FormatToVideo(inputPath, preProcessed, false); err != nil {
		return err
	}

	// 3. Extract frames
	extractCmd := exec.Command("ffmpeg", "-i", preProcessed, filepath.Join(framesDir, "%03d.png"))
	if err := extractCmd.Run(); err != nil {
		return fmt.Errorf("failed to extract frames: %v", err)
	}

	// 4. Run python script on directory
	rembgCmd := exec.Command("python3", "python/rembg_script.py", framesDir, outFramesDir)
	var stderr bytes.Buffer
	rembgCmd.Stderr = &stderr
	if err := rembgCmd.Run(); err != nil {
		return fmt.Errorf("rembg video error: %v, stderr: %s", err, stderr.String())
	}

	// 5. Recombine frames to video with alpha channel
	recombineCmd := exec.Command("ffmpeg", "-y", "-framerate", "30", "-i", filepath.Join(outFramesDir, "%03d.png"), "-c:v", "libvpx-vp9", "-pix_fmt", "yuva420p", outputPath)
	var recombineErr bytes.Buffer
	recombineCmd.Stderr = &recombineErr
	if err := recombineCmd.Run(); err != nil {
		return fmt.Errorf("failed to recombine frames: %v, stderr: %s", err, recombineErr.String())
	}

	return nil
}
