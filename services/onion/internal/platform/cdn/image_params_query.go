package cdn

import (
	"net/url"
	"strconv"
	"strings"
)

func fastlyQueryChain(query url.Values, width, height int) (string, int, int, error) {
	steps := make([]string, 0, 8)
	bg := sanitizeChainValue(firstQuery(query, "bg-color", "background", "b"))

	if step, err := cropStep(query.Get("precrop"), width, height, "crop"); err != nil {
		return "", width, height, err
	} else if step != "" {
		steps = append(steps, step)
	}

	if trim := strings.TrimSpace(query.Get("trim")); trim != "" && !isFalsey(trim) {
		value := "10"
		if trim != "true" && trim != "1" && trim != "on" {
			value = trim
		}
		steps = append(steps, "e_trim:"+sanitizeChainValue(value))
	}

	geometry, outW, outH, err := geometryStep(query, width, height)
	if err != nil {
		return "", width, height, err
	}
	if geometry != "" {
		width, height = outW, outH
		steps = append(steps, geometry)
	}

	if orient := orientStep(firstQuery(query, "orient", "rotate")); orient != "" {
		steps = append(steps, orient)
	}

	if canvas := canvasStep(query.Get("canvas"), bg); canvas != "" {
		steps = append(steps, canvas)
	}
	if pad := padStep(query.Get("pad"), bg); pad != "" {
		steps = append(steps, pad)
	}
	if border := borderStep(query.Get("border"), firstQuery(query, "border-color", "bg-color", "background", "b")); border != "" {
		steps = append(steps, border)
	}

	colorEffects := make([]string, 0, 4)
	if brightness := query.Get("brightness"); brightness != "" {
		colorEffects = append(colorEffects, "e_brightness:"+sanitizeChainValue(brightness))
	}
	if contrast := query.Get("contrast"); contrast != "" {
		colorEffects = append(colorEffects, "e_contrast:"+sanitizeChainValue(contrast))
	}
	if saturation := query.Get("saturation"); saturation != "" {
		colorEffects = append(colorEffects, "e_saturation:"+sanitizeChainValue(saturation))
	}
	if bw := query.Get("bw"); bw != "" && !isFalsey(bw) {
		colorEffects = append(colorEffects, "e_grayscale")
	}
	if grayscale := firstQuery(query, "grayscale", "greyscale"); grayscale != "" && !isFalsey(grayscale) {
		colorEffects = append(colorEffects, "e_grayscale")
	}
	if sepia := query.Get("sepia"); sepia != "" && !isFalsey(sepia) {
		colorEffects = append(colorEffects, "e_sepia")
	}
	if normalize := firstQuery(query, "normalize", "improve"); normalize != "" && !isFalsey(normalize) {
		colorEffects = append(colorEffects, "e_normalize")
	}
	if negate := query.Get("negate"); negate != "" && !isFalsey(negate) {
		colorEffects = append(colorEffects, "e_negate")
	}
	if invert := query.Get("invert"); invert != "" && !isFalsey(invert) {
		colorEffects = append(colorEffects, "e_invert")
	}
	if hue := query.Get("hue"); hue != "" {
		colorEffects = append(colorEffects, "e_hue:"+sanitizeChainValue(hue))
	}
	if gamma := query.Get("gamma"); gamma != "" {
		colorEffects = append(colorEffects, "e_gamma:"+sanitizeChainValue(gamma))
	}
	if colorize := query.Get("colorize"); colorize != "" {
		colorEffects = append(colorEffects, "e_colorize:"+sanitizeEffectArgs(colorize))
	}
	if tint := query.Get("tint"); tint != "" {
		colorEffects = append(colorEffects, "e_tint:"+sanitizeEffectArgs(tint))
	}
	if duotone := query.Get("duotone"); duotone != "" {
		colorEffects = append(colorEffects, "e_duotone:"+sanitizeEffectArgs(duotone))
	}
	if len(colorEffects) > 0 {
		steps = append(steps, strings.Join(colorEffects, "/"))
	}

	if pixelate := query.Get("pixelate"); pixelate != "" && !isFalsey(pixelate) {
		value := "8"
		if pixelate != "true" && pixelate != "1" && pixelate != "on" {
			value = pixelate
		}
		steps = append(steps, "e_pixelate:"+sanitizeChainValue(value))
	}
	if posterize := query.Get("posterize"); posterize != "" && !isFalsey(posterize) {
		value := "4"
		if posterize != "true" && posterize != "1" && posterize != "on" {
			value = posterize
		}
		steps = append(steps, "e_posterize:"+sanitizeChainValue(value))
	}
	if threshold := query.Get("threshold"); threshold != "" && !isFalsey(threshold) {
		value := "128"
		if threshold != "true" && threshold != "1" && threshold != "on" {
			value = threshold
		}
		steps = append(steps, "e_blackwhite:"+sanitizeChainValue(value))
	}
	if edge := query.Get("edge"); edge != "" && !isFalsey(edge) {
		steps = append(steps, "e_edge_detect")
	}
	if emboss := query.Get("emboss"); emboss != "" && !isFalsey(emboss) {
		steps = append(steps, "e_emboss")
	}
	if oil := query.Get("oil"); oil != "" && !isFalsey(oil) {
		value := "5"
		if oil != "true" && oil != "1" && oil != "on" {
			value = oil
		}
		steps = append(steps, "e_oil_paint:"+sanitizeChainValue(value))
	}
	if median := query.Get("median"); median != "" && !isFalsey(median) {
		value := "5"
		if median != "true" && median != "1" && median != "on" {
			value = median
		}
		steps = append(steps, "e_median:"+sanitizeChainValue(value))
	}
	if sharpen := query.Get("sharpen"); sharpen != "" && !isFalsey(sharpen) {
		value := "100"
		if sharpen != "true" && sharpen != "1" && sharpen != "on" {
			value = sharpen
		}
		steps = append(steps, "e_sharpen:"+sanitizeChainValue(value))
	}
	if blur := query.Get("blur"); blur != "" && !isFalsey(blur) {
		value := "300"
		if blur != "true" && blur != "1" && blur != "on" {
			value = blur
		}
		steps = append(steps, "e_blur:"+sanitizeChainValue(value))
	}
	if vignette := query.Get("vignette"); vignette != "" && !isFalsey(vignette) {
		value := "30"
		if vignette != "true" && vignette != "1" && vignette != "on" {
			value = vignette
		}
		steps = append(steps, "e_vignette:"+sanitizeChainValue(value))
	}
	if effect := firstQuery(query, "effect", "e"); effect != "" {
		steps = append(steps, effectStep(effect))
	}
	if radius := firstQuery(query, "radius", "round"); radius != "" {
		steps = append(steps, "r_"+sanitizeChainValue(radius))
	}
	if opacity := firstQuery(query, "opacity", "o"); opacity != "" {
		steps = append(steps, "o_"+sanitizeChainValue(opacity))
	}
	if flags := firstQuery(query, "fl"); flags != "" {
		steps = append(steps, "fl_"+sanitizeChainValue(strings.ReplaceAll(flags, ",", ".")))
	}
	if flip := query.Get("flip"); flip != "" && !isFalsey(flip) {
		steps = append(steps, "fx_flip")
	}
	if flop := query.Get("flop"); flop != "" && !isFalsey(flop) {
		steps = append(steps, "fx_flop")
	}
	if scale := query.Get("scale"); scale != "" {
		steps = append(steps, "z_"+sanitizeChainValue(scale))
	}
	if text := textStep(query); text != "" {
		steps = append(steps, text)
	}
	watermark, err := watermarkStep(query)
	if err != nil {
		return "", width, height, err
	}
	if watermark != "" {
		steps = append(steps, watermark)
	}

	return strings.Join(nonEmptySteps(steps), "/"), width, height, nil
}

func geometryStep(query url.Values, width, height int) (string, int, int, error) {
	crop := strings.TrimSpace(query.Get("crop"))
	if crop != "" {
		step, outW, outH, err := cropGeometryStep(crop, width, height)
		return step, outW, outH, err
	}

	fit := canonicalImageFit(query.Get("fit"))
	if width == 0 && height == 0 {
		return "", width, height, nil
	}
	mode := "fit"
	switch fit {
	case "cover", "crop", "fill":
		mode = "fill"
	case "bounds", "contain", "inside", "fit":
		mode = "fit"
	case "limit", "outside":
		mode = "limit"
	case "pad":
		mode = "pad"
	case "scale":
		mode = "scale"
	}
	x := sanitizeChainValue(query.Get("x"))
	y := sanitizeChainValue(query.Get("y"))
	if x != "" || y != "" {
		mode = "crop"
	}

	parts := []string{}
	if width > 0 {
		parts = append(parts, "w_"+strconv.Itoa(width))
	}
	if height > 0 {
		parts = append(parts, "h_"+strconv.Itoa(height))
	}
	parts = append(parts, "c_"+mode)
	if gravity := sanitizeChainValue(firstQuery(query, "gravity", "g")); gravity != "" {
		parts = append(parts, "g_"+gravity)
	}
	if x != "" {
		parts = append(parts, "x_"+x)
	}
	if y != "" {
		parts = append(parts, "y_"+y)
	}
	return strings.Join(parts, ","), width, height, nil
}
