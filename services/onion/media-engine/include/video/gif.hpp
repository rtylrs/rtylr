#pragma once

#include <string>

namespace voxire {
namespace video {

// makeGifPreview converts a clip into an animated GIF using a generated optimal
// palette (palettegen/paletteuse) for good quality. `fps` and `width` keep the
// file small; height follows the aspect ratio. Throws ProcessingError on
// failure.
void makeGifPreview(const std::string& inputPath, const std::string& outputPath,
                    int fps, int width);

} // namespace video
} // namespace voxire
