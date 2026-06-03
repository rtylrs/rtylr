#pragma once

#include <string>

namespace voxire {
namespace video {

// VideoCodec selects the output video codec for transcoding.
enum class VideoCodec {
    H264, // MP4 / libx264 — broad compatibility
    VP9,  // WebM / libvpx-vp9 — better compression for the web
};

VideoCodec parseVideoCodec(const std::string& s);

// transcode re-encodes the input to outputPath with the given codec and CRF
// quality (lower = higher quality; ~23 for H264, ~31 for VP9 are typical), and
// scales to width x height (0 keeps aspect). Audio is re-encoded to a matching
// codec. Throws ProcessingError on failure.
void transcode(const std::string& inputPath, const std::string& outputPath,
               VideoCodec codec, int crf, int width, int height);

} // namespace video
} // namespace voxire
