package ffmpeg

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/voxire/onion/internal/platform/logger"
)

var binaryPath = "ffmpeg"

func Init() error {
	path, err := exec.LookPath("ffmpeg")
	if err != nil {
		return fmt.Errorf("ffmpeg not found in PATH: %w", err)
	}

	binaryPath = path

	return nil
}

func IsAvailable() bool {
	_, err := exec.LookPath("ffmpeg")
	return err == nil
}

func Run(ctx context.Context, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, binaryPath, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	logger.Ctx(ctx).Debug("Running ffmpeg", logger.String("args", fmt.Sprintf("%v", args)))

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg failed: %w, stderr: %s", err, stderr.String())
	}

	return stdout.Bytes(), nil
}

func ConcatAudioFiles(ctx context.Context, inputPaths []string, outputPath string) error {
	listFile := filepath.Join(os.TempDir(), fmt.Sprintf("ffmpeg_concat_%d.txt", time.Now().UnixNano()))

	var content string
	for _, p := range inputPaths {
		content += fmt.Sprintf("file '%s'\n", p)
	}

	if err := os.WriteFile(listFile, []byte(content), 0o600); err != nil {
		return fmt.Errorf("failed to write concat list: %w", err)
	}
	defer os.Remove(listFile)

	_, err := Run(ctx, "-y", "-f", "concat", "-safe", "0", "-i", listFile, "-c", "copy", outputPath)

	return err
}

func MixAudioWithMusic(ctx context.Context, voicePath, musicPath, outputPath string, musicVolume float64) error {
	filter := fmt.Sprintf("[1:a]volume=%.2f[music];[0:a][music]amix=inputs=2:duration=first:dropout_transition=2", musicVolume)

	_, err := Run(ctx,
		"-y",
		"-i", voicePath,
		"-i", musicPath,
		"-filter_complex", filter,
		"-ac", "2",
		"-ar", "44100",
		"-b:a", "192k",
		outputPath,
	)

	return err
}

func ConvertToMP3(ctx context.Context, inputPath, outputPath string, bitrate string) error {
	if bitrate == "" {
		bitrate = "192k"
	}

	_, err := Run(ctx, "-y", "-i", inputPath, "-codec:a", "libmp3lame", "-b:a", bitrate, outputPath)

	return err
}

func GetDuration(ctx context.Context, filePath string) (float64, error) {
	cmd := exec.CommandContext(ctx, "ffprobe",
		"-v", "quiet",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		filePath,
	)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return 0, fmt.Errorf("ffprobe failed: %w", err)
	}

	var duration float64
	if _, err := fmt.Sscanf(stdout.String(), "%f", &duration); err != nil {
		return 0, fmt.Errorf("failed to parse duration: %w", err)
	}

	return duration, nil
}

func TrimAudio(ctx context.Context, inputPath, outputPath string, startSeconds, durationSeconds float64) error {
	_, err := Run(ctx,
		"-y",
		"-i", inputPath,
		"-ss", fmt.Sprintf("%.2f", startSeconds),
		"-t", fmt.Sprintf("%.2f", durationSeconds),
		"-c", "copy",
		outputPath,
	)

	return err
}
