#pragma once

#include <string>
#include <vector>

namespace voxire {
namespace video {

// scaleFilter builds an ffmpeg "scale=W:H" filter argument. A zero dimension
// becomes -1 so ffmpeg preserves the aspect ratio for that axis.
std::string scaleFilter(int width, int height);

// runFFmpeg invokes ffmpeg with the given arguments (ffmpegBin() is prepended),
// always appending -y to overwrite the output. Throws ProcessingError with the
// captured output when ffmpeg exits non-zero.
void runFFmpeg(const std::vector<std::string>& args);

} // namespace video
} // namespace voxire
