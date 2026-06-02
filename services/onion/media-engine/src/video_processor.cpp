#include "video_processor.hpp"
#include <stdexcept>

namespace voxire {

// FFmpeg integration is scaffolded here. Each method will construct an
// ffmpeg command-line invocation and execute it via popen/execvp once
// the video pipeline is activated.

void VideoProcessor::extractThumbnail(const VideoOptions& opts) {
    // TODO: invoke ffmpeg -ss <time> -i <input> -vframes 1 -vf scale=<w>:<h> <output>
    throw ProcessingError("video thumbnail extraction not yet implemented");
}

void VideoProcessor::transcode(const VideoOptions& opts) {
    // TODO: invoke ffmpeg with HLS/MP4 output options
    throw ProcessingError("video transcoding not yet implemented");
}

} // namespace voxire
