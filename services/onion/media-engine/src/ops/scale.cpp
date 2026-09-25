#include "ops/scale.hpp"
#include "common.hpp"

namespace voxire {
namespace ops {

vips::VImage scaleBy(vips::VImage img, double factor) {
    if (factor <= 0) {
        throw ProcessingError("scale factor must be greater than 0");
    }
	if (factor > 4.0) {
		factor = 4.0;
	}
	if (factor == 1.0) {
        return img;
    }
    return img.resize(factor);
}

} // namespace ops
} // namespace voxire
