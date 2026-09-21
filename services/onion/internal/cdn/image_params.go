package cdn

import (
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type imageParamOptions struct {
	negotiateDefault bool
	defaultFullImage bool
}

type imageParamResult struct {
	params       Params
	varyAccept   bool
	explicit     bool
	original     bool
	surrogateKey string
}

var imageTransformQueryKeys = map[string]bool{
	"auto":             true,
	"ar":               true,
	"aspect":           true,
	"aspect-ratio":     true,
	"b":                true,
	"background":       true,
	"background-color": true,
	"bg-color":         true,
	"blur":             true,
	"border":           true,
	"border-color":     true,
	"brightness":       true,
	"bw":               true,
	"caption":          true,
	"canvas":           true,
	"class":            true,
	"colorize":         true,
	"contrast":         true,
	"crop":             true,
	"disable":          true,
	"dpr":              true,
	"duotone":          true,
	"effect":           true,
	"emboss":           true,
	"enable":           true,
	"e":                true,
	"edge":             true,
	"f":                true,
	"fit":              true,
	"fl":               true,
	"flip":             true,
	"flop":             true,
	"format":           true,
	"frame":            true,
	"gamma":            true,
	"grayscale":        true,
	"gravity":          true,
	"greyscale":        true,
	"g":                true,
	"height":           true,
	"hue":              true,
	"improve":          true,
	"invert":           true,
	"metadata":         true,
	"median":           true,
	"negate":           true,
	"normalize":        true,
	"optimize":         true,
	"opacity":          true,
	"o":                true,
	"oil":              true,
	"orient":           true,
	"pad":              true,
	"padding":          true,
	"padding-color":    true,
	"pixelate":         true,
	"posterize":        true,
	"precrop":          true,
	"preset":           true,
	"q":                true,
	"quality":          true,
	"radius":           true,
	"resize-filter":    true,
	"rotate":           true,
	"round":            true,
	"saturation":       true,
	"scale":            true,
	"sharpen":          true,
	"sepia":            true,
	"t":                true,
	"tc":               true,
	"text":             true,
	"text-opacity":     true,
	"tf":               true,
	"tg":               true,
	"threshold":        true,
	"tint":             true,
	"tm":               true,
	"to":               true,
	"trim":             true,
	"trim-color":       true,
	"tw":               true,
	"txt":              true,
	"vignette":         true,
	"watermark":        true,
	"wm":               true,
	"wm-color":         true,
	"wm-font":          true,
	"wm-gravity":       true,
	"wm-margin":        true,
	"wm-opacity":       true,
	"wm-text":          true,
	"wm-width":         true,
	"w":                true,
	"width":            true,
	"x":                true,
	"y":                true,
}

var imageTransformClasses = map[string]string{
	"thumb":         "w_150,h_150,c_fill,g_auto,r_6,f_auto,q_auto",
	"thumbnail":     "w_150,h_150,c_fill,g_auto,r_6,f_auto,q_auto",
	"micro":         "w_48,h_48,c_fill,g_auto,r_4,f_auto,q_auto:eco",
	"icon":          "w_96,h_96,c_pad,b_transparent,f_auto,q_auto",
	"avatar":        "w_200,h_200,c_fill,g_auto,r_max,f_auto,q_auto",
	"avatar-large":  "w_512,h_512,c_fill,g_auto,r_max,f_auto,q_auto",
	"card":          "w_400,h_300,c_fill,g_auto,f_auto,q_auto",
	"card-2x":       "w_800,h_600,c_fill,g_auto,f_auto,q_auto",
	"gallery":       "w_1200,h_900,c_fill,g_auto,f_auto,q_auto",
	"gallery-thumb": "w_320,h_240,c_fill,g_auto,f_auto,q_auto:eco",
	"hero":          "w_1600,h_600,c_fill,g_auto,f_auto,q_auto",
	"hero-tall":     "w_1600,h_900,c_fill,g_auto,f_auto,q_auto",
	"banner":        "w_1920,h_640,c_fill,g_auto,f_auto,q_auto",
	"og":            "w_1200,h_630,c_fill,g_auto,f_auto,q_auto",
	"story":         "w_1080,h_1920,c_fill,g_auto,f_auto,q_auto",
	"email":         "w_720,c_limit,f_auto,q_auto:eco",
	"email-hero":    "w_1200,h_480,c_fill,g_auto,f_auto,q_auto:eco",
	"blur":          "w_64,c_fit,e_blur:1000,f_auto,q_auto:low",
	"lqip":          "w_32,c_fit,e_blur:1200,f_auto,q_auto:low",
	"placeholder":   "w_48,c_fit,e_blur:1000,f_auto,q_auto:low",
	"square":        "w_800,h_800,c_fill,g_auto,f_auto,q_auto",
	"product":       "w_900,h_900,c_pad,b_ffffff,f_auto,q_auto",
	"product-thumb": "w_300,h_300,c_pad,b_ffffff,f_auto,q_auto:eco",
	"product-card":  "w_600,h_600,c_pad,b_ffffff,f_auto,q_auto",
	"pos":           "w_640,h_640,c_pad,b_ffffff,f_auto,q_auto",
	"pos-thumb":     "w_180,h_180,c_pad,b_ffffff,f_auto,q_auto:eco",
	"pos-detail":    "w_1024,h_1024,c_pad,b_ffffff,f_auto,q_auto",
	"receipt":       "w_512,c_fit,b_ffffff,f_auto,q_auto:eco",
	"logo":          "w_512,c_limit,b_transparent,f_auto,q_auto",
	"favicon":       "w_64,h_64,c_pad,b_transparent,f_auto,q_auto",
}

func parseImageParams(query url.Values, accept string, opts imageParamOptions) (imageParamResult, error) {
	var result imageParamResult
	result.original = uploadURLWantsOriginal(query)
	if result.original {
		return result, nil
	}

	result.explicit = hasImageTransformQuery(query)

	format, varyAccept, err := requestedImageFormat(query, accept, opts.negotiateDefault)
	if err != nil {
		return result, err
	}
	result.varyAccept = varyAccept

	quality, err := queryQuality(query)
	if err != nil {
		return result, err
	}

	width, err := queryPositiveInt(query, "width", "w")
	if err != nil {
		return result, err
	}
	height, err := queryPositiveInt(query, "height", "h")
	if err != nil {
		return result, err
	}

	width, height, err = applyAspectRatio(query, width, height)
	if err != nil {
		return result, err
	}

	dpr, err := queryDPR(query)
	if err != nil {
		return result, err
	}
	if dpr > 0 {
		if width > 0 {
			width = int(math.Round(float64(width) * dpr))
		}
		if height > 0 {
			height = int(math.Round(float64(height) * dpr))
		}
	}

	chain, err := classChain(query)
	if err != nil {
		return result, err
	}
	if direct := strings.TrimSpace(query.Get("t")); direct != "" {
		chain = joinChain(chain, direct)
	}
	if chain == "" {
		chain, width, height, err = fastlyQueryChain(query, width, height)
		if err != nil {
			return result, err
		}
	} else {
		var extra string
		extra, width, height, err = fastlyQueryChain(query, width, height)
		if err != nil {
			return result, err
		}
		chain = joinChain(chain, extra)
	}
	if chainFmt, chainQ := parseChainParams(chain); chainFmt != "" || chainQ > 0 {
		if chainFmt != "" {
			if chainFmt == "auto" {
				if negotiated := negotiateFormat(accept); negotiated != "" {
					format = negotiated
				} else {
					format = "webp"
				}
				result.varyAccept = true
			} else if normalized, chainVary, normalizeErr := normalizeImageFormat(chainFmt, accept); normalizeErr != nil {
				return result, normalizeErr
			} else if normalized != "" {
				format = normalized
				result.varyAccept = result.varyAccept || chainVary
			}
		}
		if chainQ > 0 {
			quality = chainQ
		}
	}

	p := Params{
		Width:   width,
		Height:  height,
		Fit:     canonicalImageFit(query.Get("fit")),
		Format:  format,
		Quality: quality,
	}
	if chain != "" || width > 0 || height > 0 || result.explicit || opts.defaultFullImage {
		p.Chain = appendFormatStep(chain, format)
	}
	p.Defaults()
	if err := p.Validate(); err != nil {
		return result, err
	}
	result.params = p
	result.surrogateKey = imageVariantSurrogateKey(p)
	return result, nil
}

func classChain(query url.Values) (string, error) {
	raw := strings.TrimSpace(query.Get("class"))
	if raw == "" {
		raw = strings.TrimSpace(query.Get("preset"))
	}
	if raw == "" {
		return "", nil
	}
	key := strings.ToLower(raw)
	if chain, ok := imageTransformClasses[key]; ok {
		return chain, nil
	}
	return "", fmt.Errorf("unknown image class: %s", raw)
}

func joinChain(parts ...string) string {
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.Trim(part, "/ ")
		if part != "" {
			out = append(out, part)
		}
	}
	return strings.Join(out, "/")
}

func hasImageTransformQuery(query url.Values) bool {
	for key := range query {
		if imageTransformQueryKeys[strings.ToLower(key)] {
			return true
		}
	}
	return false
}

func requestedImageFormat(query url.Values, accept string, negotiateDefault bool) (string, bool, error) {
	raw := strings.TrimSpace(query.Get("format"))
	if raw == "" {
		raw = strings.TrimSpace(query.Get("f"))
	}
	if raw == "" {
		raw = formatFromAutoParam(query.Get("auto"))
	}

	if raw == "" && negotiateDefault {
		if negotiated := negotiateFormat(accept); negotiated != "" {
			return negotiated, true, nil
		}
	}
	if raw == "" {
		return "webp", false, nil
	}

	format, varyAccept, err := normalizeImageFormat(raw, accept)
	if err != nil {
		return "", false, err
	}
	return format, varyAccept, nil
}

func formatFromAutoParam(raw string) string {
	raw = strings.ToLower(strings.TrimSpace(raw))
	if raw == "" || raw == "false" || raw == "0" || raw == "off" {
		return ""
	}
	if raw == "true" || raw == "1" || raw == "on" {
		return "auto"
	}
	for _, part := range strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ' '
	}) {
		switch strings.TrimSpace(part) {
		case "avif", "webp":
			return "auto"
		}
	}
	return ""
}

func normalizeImageFormat(raw, accept string) (string, bool, error) {
	format := strings.ToLower(strings.TrimSpace(raw))
	switch format {
	case "", "source", "original":
		return "", false, nil
	case "auto":
		if negotiated := negotiateFormat(accept); negotiated != "" {
			return negotiated, true, nil
		}
		return "webp", true, nil
	case "jpg", "jpeg", "bjpg", "bjpeg", "pjpg", "pjpeg":
		return "jpeg", false, nil
	case "png", "png8":
		return "png", false, nil
	case "webp", "webply", "webpll":
		return "webp", false, nil
	case "avif", "gif", "tiff", "tif":
		return format, false, nil
	default:
		return "", false, fmt.Errorf("unsupported format: %s", raw)
	}
}

func queryQuality(query url.Values) (int, error) {
	raw := strings.TrimSpace(query.Get("quality"))
	if raw == "" {
		raw = strings.TrimSpace(query.Get("q"))
	}
	if raw == "" {
		return uploadURLWebPQuality, nil
	}
	if q, ok := autoQuality(raw); ok {
		return q, nil
	}
	q, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid quality: %s", raw)
	}
	if q < 1 || q > 100 {
		return 0, fmt.Errorf("quality must be between 1 and 100")
	}
	return q, nil
}

func autoQuality(raw string) (int, bool) {
	raw = strings.ToLower(strings.TrimSpace(raw))
	switch raw {
	case "auto", "q_auto", "auto:good", "q_auto:good":
		return uploadURLWebPQuality, true
	case "auto:best", "q_auto:best":
		return 92, true
	case "auto:eco", "q_auto:eco":
		return 75, true
	case "auto:low", "q_auto:low":
		return 60, true
	default:
		return 0, false
	}
}

func queryPositiveInt(query url.Values, keys ...string) (int, error) {
	for _, key := range keys {
		raw := strings.TrimSpace(query.Get(key))
		if raw == "" {
			continue
		}
		n, err := strconv.Atoi(raw)
		if err != nil {
			return 0, fmt.Errorf("invalid %s: %s", key, raw)
		}
		if n < 0 {
			return 0, fmt.Errorf("%s must not be negative", key)
		}
		return n, nil
	}
	return 0, nil
}

func queryDPR(query url.Values) (float64, error) {
	raw := strings.TrimSpace(query.Get("dpr"))
	if raw == "" {
		return 0, nil
	}
	if strings.EqualFold(raw, "auto") {
		return 2, nil
	}
	dpr, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid dpr: %s", raw)
	}
	if dpr < 1 || dpr > 10 {
		return 0, fmt.Errorf("dpr must be between 1 and 10")
	}
	return dpr, nil
}

func applyAspectRatio(query url.Values, width, height int) (int, int, error) {
	raw := firstQuery(query, "aspect-ratio", "aspect", "ar")
	if raw == "" {
		return width, height, nil
	}
	if width <= 0 && height <= 0 {
		return width, height, fmt.Errorf("aspect ratio requires width or height")
	}
	w, h, err := ratioDimensions(raw, width, height)
	if err != nil {
		return width, height, err
	}
	if width <= 0 {
		width = w
	}
	if height <= 0 {
		height = h
	}
	return width, height, nil
}

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

func cropGeometryStep(raw string, width, height int) (string, int, int, error) {
	parts := splitParamList(raw)
	if len(parts) == 0 {
		return "", width, height, nil
	}

	var cropW, cropH int
	var err error
	if strings.Contains(parts[0], ":") {
		cropW, cropH, err = ratioDimensions(parts[0], width, height)
		if err != nil {
			return "", width, height, err
		}
	} else {
		cropW, cropH, err = pixelDimensions(parts[0], parts)
		if err != nil {
			return "", width, height, err
		}
	}
	if cropW == 0 && width > 0 {
		cropW = width
	}
	if cropH == 0 && height > 0 {
		cropH = height
	}
	if cropW <= 0 || cropH <= 0 {
		return "", width, height, fmt.Errorf("crop requires width and height")
	}

	mode := "fill"
	gravity := ""
	extras := []string{}
	for _, part := range parts[1:] {
		switch {
		case part == "smart":
			gravity = "auto"
		case part == "safe":
			// Our crop op clamps to the source bounds, so safe is the default.
		case strings.HasPrefix(part, "x"):
			extras = append(extras, "x_"+strings.TrimPrefix(part, "x"))
			mode = "crop"
		case strings.HasPrefix(part, "y"):
			extras = append(extras, "y_"+strings.TrimPrefix(part, "y"))
			mode = "crop"
		case strings.HasPrefix(part, "offset-x"), strings.HasPrefix(part, "offset-y"):
			return "", width, height, fmt.Errorf("crop offset positioning is not supported")
		}
	}
	values := []string{"w_" + strconv.Itoa(cropW), "h_" + strconv.Itoa(cropH), "c_" + mode}
	if gravity != "" {
		values = append(values, "g_"+gravity)
	}
	values = append(values, extras...)
	return strings.Join(values, ","), cropW, cropH, nil
}

func cropStep(raw string, width, height int, mode string) (string, error) {
	if strings.TrimSpace(raw) == "" {
		return "", nil
	}
	step, _, _, err := cropGeometryStep(raw, width, height)
	if err != nil {
		return "", err
	}
	if mode == "crop" {
		step = strings.Replace(step, "c_fill", "c_crop", 1)
	}
	return step, nil
}

func ratioDimensions(raw string, width, height int) (int, int, error) {
	parts := strings.SplitN(raw, ":", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid crop ratio: %s", raw)
	}
	a, err := strconv.ParseFloat(parts[0], 64)
	if err != nil || a <= 0 {
		return 0, 0, fmt.Errorf("invalid crop ratio: %s", raw)
	}
	b, err := strconv.ParseFloat(parts[1], 64)
	if err != nil || b <= 0 {
		return 0, 0, fmt.Errorf("invalid crop ratio: %s", raw)
	}
	if width <= 0 && height <= 0 {
		return 0, 0, fmt.Errorf("ratio crop requires width or height")
	}
	if width > 0 && height <= 0 {
		height = int(math.Round(float64(width) * b / a))
	}
	if height > 0 && width <= 0 {
		width = int(math.Round(float64(height) * a / b))
	}
	return width, height, nil
}

func pixelDimensions(first string, parts []string) (int, int, error) {
	width, err := parseDimension(first)
	if err != nil {
		return 0, 0, err
	}
	height := 0
	if len(parts) > 1 {
		if n, parseErr := parseDimension(parts[1]); parseErr == nil {
			height = n
		}
	}
	return width, height, nil
}

func parseDimension(raw string) (int, error) {
	raw = strings.TrimSuffix(strings.TrimSpace(raw), "p")
	n, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid dimension: %s", raw)
	}
	if n < 0 {
		return 0, fmt.Errorf("dimension must not be negative")
	}
	return int(math.Round(n)), nil
}

// orientStep maps the request's orient/rotate value onto a chain step. Bare
// h/v agree with the hflip/vflip aliases: fx_flop is a left-to-right mirror
// and fx_flip a top-to-bottom one (media-engine apply.cpp).
func orientStep(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "auto":
		return ""
	case "90", "r90":
		return "a_90"
	case "180", "r180":
		return "a_180"
	case "270", "r270", "-90":
		return "a_270"
	case "v", "flip", "vflip":
		return "fx_flip"
	case "h", "flop", "hflip":
		return "fx_flop"
	default:
		return "a_" + sanitizeChainValue(raw)
	}
}

func canvasStep(raw, bg string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	raw = strings.NewReplacer("x", ",", "X", ",").Replace(raw)
	parts := splitParamList(raw)
	if len(parts) < 2 {
		return ""
	}
	w, errW := parseDimension(parts[0])
	h, errH := parseDimension(parts[1])
	if errW != nil || errH != nil || w <= 0 || h <= 0 {
		return ""
	}
	step := fmt.Sprintf("w_%d,h_%d,c_pad", w, h)
	if bg != "" {
		step += ",b_" + bg
	}
	return step
}

func padStep(raw, bg string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || isFalsey(raw) {
		return ""
	}
	n, err := parseDimension(splitParamList(raw)[0])
	if err != nil || n <= 0 || n > maxTransformBorder {
		return ""
	}
	if bg == "" {
		bg = "transparent"
	}
	return "bo_" + strconv.Itoa(n) + ":" + bg
}

func borderStep(raw, bg string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || isFalsey(raw) {
		return ""
	}
	parts := splitParamList(raw)
	n, err := parseDimension(parts[0])
	if err != nil || n <= 0 || n > maxTransformBorder {
		return ""
	}
	color := bg
	if len(parts) > 1 {
		color = sanitizeChainValue(parts[1])
	}
	if color == "" {
		color = "black"
	}
	return "bo_" + strconv.Itoa(n) + ":" + color
}

func appendFormatStep(chain, format string) string {
	if chainFmt, _ := parseChainParams(chain); chainFmt != "" {
		return chain
	}
	step := "f_" + strings.ToLower(format)
	if strings.TrimSpace(chain) == "" {
		return step
	}
	return chain + "/" + step
}

func canonicalImageFit(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "bounds":
		return "bounds"
	case "crop":
		return "crop"
	case "cover":
		return "cover"
	case "contain", "inside":
		return "contain"
	case "fill":
		return "fill"
	case "limit", "outside":
		return "limit"
	case "pad":
		return "pad"
	case "scale":
		return "scale"
	default:
		return "cover"
	}
}

func splitParamList(raw string) []string {
	values := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ';'
	})
	out := make([]string, 0, len(values))
	for _, value := range values {
		if v := strings.TrimSpace(value); v != "" {
			out = append(out, v)
		}
	}
	return out
}

func sanitizeChainValue(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "#")
	raw = strings.ReplaceAll(raw, " ", "_")
	raw = strings.ReplaceAll(raw, "/", "")
	raw = strings.ReplaceAll(raw, "\x00", "")
	raw = strings.ReplaceAll(raw, "\r", "")
	raw = strings.ReplaceAll(raw, "\n", "")
	if raw == "" {
		return raw
	}
	return raw
}

func sanitizeEffectArgs(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.ReplaceAll(raw, ",", ":")
	raw = strings.ReplaceAll(raw, "#", "")
	raw = strings.ReplaceAll(raw, " ", "_")
	raw = strings.ReplaceAll(raw, "/", "")
	raw = strings.ReplaceAll(raw, "\x00", "")
	raw = strings.ReplaceAll(raw, "\r", "")
	raw = strings.ReplaceAll(raw, "\n", "")
	return raw
}

func firstQuery(query url.Values, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(query.Get(key)); value != "" {
			return value
		}
	}
	return ""
}

func effectStep(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if strings.HasPrefix(raw, "e_") {
		return sanitizeChainValue(raw)
	}
	return "e_" + sanitizeEffectArgs(raw)
}

func textStep(query url.Values) string {
	text := firstQuery(query, "text", "txt", "caption")
	if text == "" {
		return ""
	}
	return textOverlayStep(text, firstQuery(query, "font", "tf"), query.Get("tc"), query.Get("tw"), query.Get("tg"), query.Get("tm"), firstQuery(query, "text-opacity", "to"))
}

func watermarkStep(query url.Values) (string, error) {
	text := firstQuery(query, "wm-text", "watermark", "wm")
	if text == "" {
		return "", nil
	}
	if looksLikeRemoteOverlay(text) {
		return "", fmt.Errorf("remote image watermarks are not supported")
	}
	return textOverlayStep(
		text,
		firstQuery(query, "wm-font", "font", "tf"),
		firstQuery(query, "wm-color", "tc"),
		query.Get("wm-width"),
		firstQuery(query, "wm-gravity", "tg"),
		firstQuery(query, "wm-margin", "tm"),
		query.Get("wm-opacity"),
	), nil
}

func textOverlayStep(text, font, color, maxWidth, gravity, margin, opacity string) string {
	parts := []string{"txt_" + sanitizeTextToken(text)}
	if font := sanitizeTextToken(font); font != "" {
		parts = append(parts, "font_"+font)
	}
	if color := sanitizeChainValue(color); color != "" {
		parts = append(parts, "tc_"+color)
	}
	if maxWidth := sanitizeChainValue(maxWidth); maxWidth != "" {
		parts = append(parts, "tw_"+maxWidth)
	}
	if gravity := sanitizeChainValue(gravity); gravity != "" {
		parts = append(parts, "tg_"+gravity)
	}
	if margin := sanitizeChainValue(margin); margin != "" {
		parts = append(parts, "tm_"+margin)
	}
	if opacity := sanitizeChainValue(opacity); opacity != "" {
		parts = append(parts, "to_"+opacity)
	}
	return strings.Join(parts, ",")
}

func looksLikeRemoteOverlay(raw string) bool {
	lower := strings.ToLower(strings.TrimSpace(raw))
	return strings.HasPrefix(lower, "http://") ||
		strings.HasPrefix(lower, "https://") ||
		strings.HasPrefix(lower, "//")
}

func sanitizeTextToken(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.ReplaceAll(raw, ",", " ")
	raw = strings.ReplaceAll(raw, "/", " ")
	raw = strings.ReplaceAll(raw, "\x00", "")
	raw = strings.ReplaceAll(raw, "\r", " ")
	raw = strings.ReplaceAll(raw, "\n", " ")
	raw = strings.Join(strings.Fields(raw), "_")
	if len(raw) > 200 {
		raw = raw[:200]
	}
	return raw
}

func nonEmptySteps(steps []string) []string {
	out := steps[:0]
	for _, step := range steps {
		if strings.TrimSpace(step) != "" {
			out = append(out, step)
		}
	}
	return out
}

func isFalsey(raw string) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "0", "false", "off", "none", "no":
		return true
	default:
		return false
	}
}

func imageVariantSurrogateKey(p Params) string {
	return "cdn-variant-" + p.CacheKey("variant")
}

func setImageVariantHeaders(w http.ResponseWriter, p Params, varyAccept bool, tags ...string) {
	if varyAccept {
		w.Header().Set("Vary", "Accept")
	}
	if len(tags) > 0 {
		w.Header().Set("Surrogate-Key", strings.Join(tags, " "))
	}
	w.Header().Set("Voxire-CDN-Transform", p.Chain)
}

func queryInt(query url.Values, key string) int {
	v := query.Get(key)
	if v == "" {
		return 0
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0
	}
	return n
}
