#include "video/gif.hpp"
#include "video/util.hpp"
#include "common.hpp"
#include <string>

namespace voxire {
namespace video {

void makeGifPreview(const std::string& inputPath, const std::string& outputPath,
                    int fps, int width) {
    if (fps <= 0) {
        throw ProcessingError("gif fps must be greater than 0");
    }
    if (width <= 0) {
        throw ProcessingError("gif width must be greater than 0");
    }
    // Single-graph palettegen/paletteuse gives far better color than the
    // default 256-color quantizer.
    const std::string filter =
        "fps=" + std::to_string(fps) +
        ",scale=" + std::to_string(width) + ":-1:flags=lanczos,"
        "split[s0][s1];[s0]palettegen[p];[s1][p]paletteuse";

    runFFmpeg({
        "-i", inputPath,
        "-vf", filter,
        "-loop", "0",
        outputPath,
    });
}

} // namespace video
} // namespace voxire
