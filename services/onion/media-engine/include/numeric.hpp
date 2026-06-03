#pragma once

#include "common.hpp"
#include <string>
#include <cstdlib>

namespace voxire {

// clampInt constrains v to the inclusive range [lo, hi].
inline int clampInt(int v, int lo, int hi) {
    if (v < lo) return lo;
    if (v > hi) return hi;
    return v;
}

// clampDouble constrains v to the inclusive range [lo, hi].
inline double clampDouble(double v, double lo, double hi) {
    if (v < lo) return lo;
    if (v > hi) return hi;
    return v;
}

// parseIntStrict parses a base-10 integer, throwing ProcessingError on any
// trailing garbage or empty input (unlike atoi, which silently returns 0).
inline int parseIntStrict(const std::string& s, const std::string& field) {
    if (s.empty()) {
        throw ProcessingError(field + " must be an integer, got empty value");
    }
    char* end = nullptr;
    const long v = std::strtol(s.c_str(), &end, 10);
    if (end == s.c_str() || *end != '\0') {
        throw ProcessingError(field + " must be an integer, got: " + s);
    }
    return static_cast<int>(v);
}

// parseDoubleStrict parses a floating-point value with the same strictness.
inline double parseDoubleStrict(const std::string& s, const std::string& field) {
    if (s.empty()) {
        throw ProcessingError(field + " must be a number, got empty value");
    }
    char* end = nullptr;
    const double v = std::strtod(s.c_str(), &end);
    if (end == s.c_str() || *end != '\0') {
        throw ProcessingError(field + " must be a number, got: " + s);
    }
    return v;
}

} // namespace voxire
