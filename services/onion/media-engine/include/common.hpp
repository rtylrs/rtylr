#pragma once

#include <string>
#include <stdexcept>
#include <iostream>
#include <algorithm>
#include <cctype>

namespace voxire {

struct ProcessingError : std::runtime_error {
    explicit ProcessingError(const std::string& msg) : std::runtime_error(msg) {}
};

// toLower returns an ASCII-lowercased copy of s.
inline std::string toLower(const std::string& s) {
    std::string out(s.size(), '\0');
    std::transform(s.begin(), s.end(), out.begin(),
                   [](unsigned char c) { return static_cast<char>(std::tolower(c)); });
    return out;
}

enum class FitMode {
    Cover,
    Contain,
    Fill,
    Inside,
    Outside,
};

inline FitMode parseFitMode(const std::string& s) {
    const std::string m = toLower(s);
    if (m == "cover")   return FitMode::Cover;
    if (m == "contain") return FitMode::Contain;
    if (m == "fill")    return FitMode::Fill;
    if (m == "inside")  return FitMode::Inside;
    if (m == "outside") return FitMode::Outside;
    throw ProcessingError("unknown fit mode: " + s);
}

} // namespace voxire
