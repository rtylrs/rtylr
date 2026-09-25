#include "ffmpeg.hpp"
#include <cstdlib>

namespace voxire {

std::string ffmpegBin() {
    const char* env = std::getenv("FFMPEG_BIN");
    return (env && *env) ? std::string(env) : std::string("ffmpeg");
}

std::string ffprobeBin() {
    const char* env = std::getenv("FFPROBE_BIN");
    return (env && *env) ? std::string(env) : std::string("ffprobe");
}

} // namespace voxire
