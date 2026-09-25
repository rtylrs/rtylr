#pragma once

#include "common.hpp"
#include <string>

namespace voxire {

// VideoOptions holds parameters for future FFmpeg-based video processing.
struct VideoOptions {
    std::string inputPath;
    std::string outputPath;
    int         thumbnailTimeSecs = 1;
    int         width             = 0;
    int         height            = 0;
};

// VideoProcessor is a scaffold for FFmpeg integration.
// Planned features:
//   - thumbnail extraction  (thumbnail subcommand)
//   - HLS transcoding       (transcode subcommand)
//   - format conversion     (convert subcommand)
class VideoProcessor {
public:
    // extractThumbnail extracts a single frame at thumbnailTimeSecs.
    // NOT YET IMPLEMENTED — will invoke FFmpeg subprocess.
    void extractThumbnail(const VideoOptions& opts);

    // transcode converts the video to a different format or bitrate.
    // NOT YET IMPLEMENTED.
    void transcode(const VideoOptions& opts);
};

} // namespace voxire
