#pragma once

#include <string>

namespace voxire {
namespace transform {

// presetChain returns the transformation chain for a named preset, or an empty
// string when the name is unknown. Presets are convenience shorthands built
// from supported directives (so they can be passed straight to applyChain).
inline std::string presetChain(const std::string& name) {
    if (name == "thumb")  return "w_150,h_150,c_fill,g_auto,f_auto,q_auto";
    if (name == "avatar") return "w_200,h_200,c_fill,g_auto,r_max,f_auto,q_auto";
    if (name == "card")   return "w_400,h_300,c_fill,g_auto,f_auto,q_auto";
    if (name == "hero")   return "w_1600,h_600,c_fill,g_auto,f_auto,q_auto";
    if (name == "og")     return "w_1200,h_630,c_fill,g_auto,f_auto,q_auto";
    if (name == "blur")   return "w_64,c_fit,e_blur:1000,f_auto,q_auto";
    return "";
}

} // namespace transform
} // namespace voxire
