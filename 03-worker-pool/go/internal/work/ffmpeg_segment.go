package work

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// FFmpegSegmentConfig runs one ffmpeg job.
// If UploadDir is set, transcodes uploaded file uploadDir/{taskID}.* → OutputDir/{taskID}.mp4.
// Otherwise: lavfi (empty Input) or short grab from Input (see RunFFmpegSegment).
type FFmpegSegmentConfig struct {
	Bin        string
	WorkDir    string
	Input      string
	StreamCopy bool
	UploadDir  string
	OutputDir  string
}

// RunFFmpegSegment encodes one segment file under WorkDir (created if missing).
func RunFFmpegSegment(ctx context.Context, cfg FFmpegSegmentConfig, taskID string) error {
	if err := os.MkdirAll(cfg.WorkDir, 0o755); err != nil {
		return err
	}

	ext := ".mp4"
	if cfg.Input != "" && cfg.StreamCopy {
		ext = ".ts"
	}
	out := filepath.Join(cfg.WorkDir, fmt.Sprintf("seg_%s_%d%s", SafeTaskFilePart(taskID), time.Now().UnixNano(), ext))

	args := []string{"-hide_banner", "-loglevel", "error", "-y"}
	if cfg.Input == "" {
		args = append(args,
			"-f", "lavfi",
			"-i", "testsrc=size=320x240:rate=25",
			"-t", "1",
			"-c:v", "libx264",
			"-preset", "ultrafast",
			"-pix_fmt", "yuv420p",
			"-an",
			out,
		)
	} else {
		args = append(args, "-i", cfg.Input, "-t", "1")
		if cfg.StreamCopy {
			args = append(args, "-c", "copy", out)
		} else {
			args = append(args,
				"-c:v", "libx264",
				"-preset", "ultrafast",
				"-pix_fmt", "yuv420p",
				"-an",
				out,
			)
		}
	}

	cmd := exec.CommandContext(ctx, cfg.Bin, args...)
	return cmd.Run()
}

// SafeTaskFilePart maps a task/job id to a single path segment (no slashes / metachars).
func SafeTaskFilePart(id string) string {
	const max = 80
	runes := []rune(id)
	if len(runes) > max {
		runes = runes[:max]
	}
	var b strings.Builder
	for _, r := range runes {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	s := b.String()
	if s == "" {
		return "task"
	}
	return s
}
