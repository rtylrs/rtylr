#pragma once

#include <map>
#include <string>
#include <vector>

namespace voxire {
namespace transform {

// splitSteps splits a Cloudinary-style transformation chain on '/' into its
// individual steps, e.g. "w_300,c_fill/e_blur:200" -> {"w_300,c_fill",
// "e_blur:200"}. Empty steps are dropped.
inline std::vector<std::string> splitSteps(const std::string& chain) {
    std::vector<std::string> steps;
    std::string cur;
    for (char c : chain) {
        if (c == '/') {
            if (!cur.empty()) steps.push_back(cur);
            cur.clear();
        } else {
            cur.push_back(c);
        }
    }
    if (!cur.empty()) steps.push_back(cur);
    return steps;
}

// parseParams parses one step ("w_300,c_fill,e_blur:200") into a key->value
// map keyed by the short code before the first '_'. The value keeps everything
// after that underscore (so "e_blur:200" -> {"e": "blur:200"}). A token with no
// underscore maps to an empty value.
inline std::map<std::string, std::string> parseParams(const std::string& step) {
    std::map<std::string, std::string> out;
    std::string token;
    auto flush = [&]() {
        if (token.empty()) return;
        const auto us = token.find('_');
        if (us == std::string::npos) {
            out[token] = "";
        } else {
            out[token.substr(0, us)] = token.substr(us + 1);
        }
        token.clear();
    };
    for (char c : step) {
        if (c == ',') {
            flush();
        } else {
            token.push_back(c);
        }
    }
    flush();
    return out;
}

} // namespace transform
} // namespace voxire
