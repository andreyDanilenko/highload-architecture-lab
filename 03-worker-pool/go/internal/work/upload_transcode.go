package work

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// RunFFmpegOrUpload runs lavfi/url segment (UploadDir empty) or transcodes an uploaded file (UploadDir set).
func RunFFmpegOrUpload(ctx context.Context, cfg FFmpegSegmentConfig, taskID string) error {
	if cfg.UploadDir != "" {
		if cfg.OutputDir == "" {
			return errors.New("work: upload transcode requires OutputDir")
		}
		return runTranscodeUpload(ctx, cfg.Bin, cfg.UploadDir, cfg.OutputDir, taskID)
	}
	return RunFFmpegSegment(ctx, cfg, taskID)
}

// runTranscodeUpload reads tmp/video-uploads/{jobID}.* and writes tmp/video-outputs/{jobID}.mp4 (real ffmpeg).
func runTranscodeUpload(ctx context.Context, bin, uploadDir, outputDir, jobID string) error {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return err
	}
	safe := SafeTaskFilePart(jobID)
	pattern := filepath.Join(uploadDir, safe+".*")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}
	if len(matches) != 1 {
		return fmt.Errorf("upload for job %s: want 1 file matching %s, got %d", jobID, pattern, len(matches))
	}
	in := matches[0]
	out := filepath.Join(outputDir, safe+".mp4")
	args := []string{
		"-hide_banner", "-loglevel", "error", "-y",
		"-i", in,
		"-c:v", "libx264", "-preset", "fast", "-crf", "23", "-pix_fmt", "yuv420p",
		"-movflags", "+faststart",
		"-an",
		out,
	}
	cmd := exec.CommandContext(ctx, bin, args...)
	return cmd.Run()
}
