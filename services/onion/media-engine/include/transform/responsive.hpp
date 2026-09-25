#pragma once

#include <cmath>
#include <vector>

namespace voxire {
namespace transform {

// breakpointWidths returns `count` evenly spaced widths from minWidth to
// maxWidth (inclusive), for generating a responsive srcset. count<1 or an
// inverted range yields an empty list; count==1 yields just maxWidth.
inline std::vector<int> breakpointWidths(int minWidth, int maxWidth, int count) {
    std::vector<int> widths;
    if (count < 1 || maxWidth < minWidth) {
        return widths;
    }
    if (count == 1) {
        widths.push_back(maxWidth);
        return widths;
    }
    const double step = static_cast<double>(maxWidth - minWidth) / (count - 1);
    for (int i = 0; i < count; ++i) {
        widths.push_back(static_cast<int>(std::lround(minWidth + i * step)));
    }
    return widths;
}

} // namespace transform
} // namespace voxire
