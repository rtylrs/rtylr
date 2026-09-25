#include "video/mute.hpp"
#include "video/util.hpp"

namespace voxire {
namespace video {

void muteVideo(const std::string& inputPath, const std::string& outputPath) {
    // -an drops audio; copy the video stream so this is a fast remux.
    runFFmpeg({
        "-i", inputPath,
        "-an",
        "-c:v", "copy",
        outputPath,
    });
}

} // namespace video
} // namespace voxire
