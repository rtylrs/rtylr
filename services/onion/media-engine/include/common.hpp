#pragma once

#include <string>
#include <stdexcept>
#include <iostream>

namespace voxire {

struct ProcessingError : std::runtime_error {
    explicit ProcessingError(const std::string& msg) : std::runtime_error(msg) {}
};

enum class FitMode {
    Cover,
    Contain,
    Fill,
    Inside,
    Outside,
};

inline FitMode parseFitMode(const std::string& s) {
    if (s == "cover")   return FitMode::Cover;
    if (s == "contain") return FitMode::Contain;
    if (s == "fill")    return FitMode::Fill;
    if (s == "inside")  return FitMode::Inside;
    if (s == "outside") return FitMode::Outside;
    throw ProcessingError("unknown fit mode: " + s);
}

} // namespace voxire
