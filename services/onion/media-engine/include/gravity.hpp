#pragma once

#include "common.hpp"
#include "geometry.hpp"
#include <string>

namespace voxire {

// Gravity names a placement anchor, mirroring Cloudinary's gravity options. It
// positions an inner box within an outer one (for pad, overlay, watermark…).
enum class Gravity {
    Center,
    North,
    South,
    East,
    West,
    NorthEast,
    NorthWest,
    SouthEast,
    SouthWest,
};

inline Gravity parseGravity(const std::string& s) {
    if (s == "center" || s == "centre") return Gravity::Center;
    if (s == "north"     || s == "top")          return Gravity::North;
    if (s == "south"     || s == "bottom")       return Gravity::South;
    if (s == "east"      || s == "right")        return Gravity::East;
    if (s == "west"      || s == "left")         return Gravity::West;
    if (s == "north_east" || s == "top_right")    return Gravity::NorthEast;
    if (s == "north_west" || s == "top_left")     return Gravity::NorthWest;
    if (s == "south_east" || s == "bottom_right") return Gravity::SouthEast;
    if (s == "south_west" || s == "bottom_left")  return Gravity::SouthWest;
    throw ProcessingError("unknown gravity: " + s);
}

// offsetFor returns the top-left (x,y) at which to place an innerW x innerH box
// inside an outerW x outerH box for the given gravity. Negative results (inner
// larger than outer) are left as-is for the caller to clamp.
inline void offsetFor(Gravity g, int innerW, int innerH, int outerW, int outerH,
                      int& x, int& y) {
    const int cx = (outerW - innerW) / 2;
    const int cy = (outerH - innerH) / 2;
    const int right = outerW - innerW;
    const int bottom = outerH - innerH;

    switch (g) {
        case Gravity::Center:    x = cx;    y = cy;     break;
        case Gravity::North:     x = cx;    y = 0;      break;
        case Gravity::South:     x = cx;    y = bottom; break;
        case Gravity::East:      x = right; y = cy;     break;
        case Gravity::West:      x = 0;     y = cy;     break;
        case Gravity::NorthEast: x = right; y = 0;      break;
        case Gravity::NorthWest: x = 0;     y = 0;      break;
        case Gravity::SouthEast: x = right; y = bottom; break;
        case Gravity::SouthWest: x = 0;     y = bottom; break;
    }
}

} // namespace voxire
