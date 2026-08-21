#pragma once

#include "common.hpp"
#include <array>
#include <string>
#include <utility>
#include <vector>
#include <cstdio>

namespace voxire {

// Color is an 8-bit RGBA color. Components are 0-255.
struct Color {
    int r = 0;
    int g = 0;
    int b = 0;
    int a = 255;

    // bands returns the components as a vector suitable for libvips background
    // options: RGB for opaque colors, RGBA when the alpha is not fully opaque.
    std::vector<double> bands() const {
        if (a >= 255) {
            return {static_cast<double>(r), static_cast<double>(g), static_cast<double>(b)};
        }
        return {static_cast<double>(r), static_cast<double>(g),
                static_cast<double>(b), static_cast<double>(a)};
    }
};

namespace detail {
inline int hexPair(const std::string& s, std::size_t i) {
    auto val = [](char c) -> int {
        if (c >= '0' && c <= '9') return c - '0';
        if (c >= 'a' && c <= 'f') return c - 'a' + 10;
        if (c >= 'A' && c <= 'F') return c - 'A' + 10;
        return -1;
    };
    const int hi = val(s[i]);
    const int lo = val(s[i + 1]);
    if (hi < 0 || lo < 0) {
        throw ProcessingError("invalid hex color digit in: " + s);
    }
    return hi * 16 + lo;
}
} // namespace detail

// Named colors accepted by parseColor beyond hex. Covers the CSS basics the
// Cloudinary b_/tc_ directives see in practice.
inline const std::array<std::pair<const char*, Color>, 22>& namedColors() {
    static const std::array<std::pair<const char*, Color>, 22> kNames = {{
        {"black",    {0, 0, 0, 255}},
        {"white",    {255, 255, 255, 255}},
        {"transparent", {0, 0, 0, 0}},
        {"red",      {255, 0, 0, 255}},
        {"green",    {0, 128, 0, 255}},
        {"lime",     {0, 255, 0, 255}},
        {"blue",     {0, 0, 255, 255}},
        {"yellow",   {255, 255, 0, 255}},
        {"cyan",     {0, 255, 255, 255}},
        {"aqua",     {0, 255, 255, 255}},
        {"magenta",  {255, 0, 255, 255}},
        {"fuchsia",  {255, 0, 255, 255}},
        {"orange",   {255, 165, 0, 255}},
        {"purple",   {128, 0, 128, 255}},
        {"pink",     {255, 192, 203, 255}},
        {"brown",    {165, 42, 42, 255}},
        {"maroon",   {128, 0, 0, 255}},
        {"navy",     {0, 0, 128, 255}},
        {"olive",    {128, 128, 0, 255}},
        {"teal",     {0, 128, 128, 255}},
        {"gray",     {128, 128, 128, 255}},
        {"grey",     {128, 128, 128, 255}},
    }};
    return kNames;
}

// parseColor accepts #RGB, #RGBA, #RRGGBB, #RRGGBBAA (with or without the
// leading '#'), plus common named colors (see namedColors). Throws
// ProcessingError on anything else.
inline Color parseColor(const std::string& in) {
    for (const auto& [name, col] : namedColors()) {
        if (in == name) return col;
    }

    std::string s = in;
    if (!s.empty() && s[0] == '#') s = s.substr(1);

    Color c;
    switch (s.size()) {
        case 3: // RGB shorthand
            c.r = detail::hexPair(std::string(2, s[0]), 0);
            c.g = detail::hexPair(std::string(2, s[1]), 0);
            c.b = detail::hexPair(std::string(2, s[2]), 0);
            return c;
        case 4: // RGBA shorthand
            c.r = detail::hexPair(std::string(2, s[0]), 0);
            c.g = detail::hexPair(std::string(2, s[1]), 0);
            c.b = detail::hexPair(std::string(2, s[2]), 0);
            c.a = detail::hexPair(std::string(2, s[3]), 0);
            return c;
        case 6: // RRGGBB
            c.r = detail::hexPair(s, 0);
            c.g = detail::hexPair(s, 2);
            c.b = detail::hexPair(s, 4);
            return c;
        case 8: // RRGGBBAA
            c.r = detail::hexPair(s, 0);
            c.g = detail::hexPair(s, 2);
            c.b = detail::hexPair(s, 4);
            c.a = detail::hexPair(s, 6);
            return c;
        default:
            throw ProcessingError("invalid color (expected hex or named): " + in);
    }
}

} // namespace voxire
