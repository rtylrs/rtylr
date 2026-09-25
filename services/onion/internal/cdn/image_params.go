package cdn

import (
	"math"
	"net/url"
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
