#include "video_processor.hpp"
#include <stdexcept>

namespace voxire {

// FFmpeg integration is scaffolded here. Each method will construct an
// ffmpeg command-line invocation and execute it via popen/execvp once
// the video pipeline is activated.

void VideoProcessor::extractThumbnail(const VideoOptions& opts) {
    (void)opts; // unused until the FFmpeg pipeline is wired up
    // TODO: invoke ffmpeg -ss <time> -i <input> -vframes 1 -vf scale=<w>:<h> <output>
    throw ProcessingError("video thumbnail extraction is not implemented yet");
}

void VideoProcessor::transcode(const VideoOptions& opts) {
    (void)opts; // unused until the FFmpeg pipeline is wired up
    // TODO: invoke ffmpeg with HLS/MP4 output options
    throw ProcessingError("video transcoding is not implemented yet");
}

} // namespace voxire
