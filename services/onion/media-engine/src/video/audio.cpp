#include "video/audio.hpp"
#include "video/util.hpp"

namespace voxire {
namespace video {

void extractAudio(const std::string& inputPath, const std::string& outputPath) {
    // -vn drops video; copy the audio stream untouched.
    runFFmpeg({
        "-i", inputPath,
        "-vn",
        "-acodec", "copy",
        outputPath,
    });
}

} // namespace video
} // namespace voxire
