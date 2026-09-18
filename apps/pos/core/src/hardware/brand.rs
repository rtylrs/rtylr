//! USB VID → thermal printer brand classification.

/// Map USB Vendor ID to thermal printer brand name.
///
/// Covers the major brands sold globally plus the Brazilian market
/// (Elgin, Bematech). Falls back to a manufacturer-string heuristic when
/// the VID is unknown.
pub fn classify_thermal_brand(vid: Option<u16>, manufacturer: Option<&str>) -> &'static str {
    match vid {
        Some(0x04b8) => "epson",       // Seiko Epson Corp
        Some(0x0519) => "star",        // Star Micronics
        Some(0x0dd4) => "bematech",    // Bematech S/A (Brazilian market)
        Some(0x20d1) => "elgin",       // Elgin (Brazilian market)
        Some(0x0404) => "ncr",         // NCR Corporation
        Some(0x1a86) => "citizen",     // Citizen Systems (also CH340 clone - check product)
        Some(0x0425) => "custom",      // Custom S.p.A. (kiosk/ticket printers)
        Some(0x154f) => "sewoo",       // Sewoo Tech (Korean POS printers)
        Some(0x0fe6) => "datecs",      // Datecs Ltd (fiscal printers)
        Some(0x2730) => "rongta",      // Rongta Technology
        _ => {
            if let Some(mfr) = manufacturer {
                let m = mfr.to_lowercase();
                if m.contains("epson")    { return "epson"; }
                if m.contains("star")     { return "star"; }
                if m.contains("bematech") { return "bematech"; }
                if m.contains("elgin")    { return "elgin"; }
                if m.contains("ncr")      { return "ncr"; }
                if m.contains("citizen")  { return "citizen"; }
                if m.contains("zebra")    { return "zebra"; }
            }
            "unknown"
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test] fn brand_epson_vid() { assert_eq!(classify_thermal_brand(Some(0x04b8), None), "epson"); }
    #[test] fn brand_star_vid() { assert_eq!(classify_thermal_brand(Some(0x0519), None), "star"); }
    #[test] fn brand_bematech_vid() { assert_eq!(classify_thermal_brand(Some(0x0dd4), None), "bematech"); }
    #[test] fn brand_elgin_vid() { assert_eq!(classify_thermal_brand(Some(0x20d1), None), "elgin"); }
    #[test] fn brand_ncr_vid() { assert_eq!(classify_thermal_brand(Some(0x0404), None), "ncr"); }
    #[test] fn brand_citizen_vid() { assert_eq!(classify_thermal_brand(Some(0x1a86), None), "citizen"); }
    #[test] fn brand_custom_vid() { assert_eq!(classify_thermal_brand(Some(0x0425), None), "custom"); }
    #[test] fn brand_sewoo_vid() { assert_eq!(classify_thermal_brand(Some(0x154f), None), "sewoo"); }
    #[test] fn brand_datecs_vid() { assert_eq!(classify_thermal_brand(Some(0x0fe6), None), "datecs"); }
    #[test] fn brand_rongta_vid() { assert_eq!(classify_thermal_brand(Some(0x2730), None), "rongta"); }

    #[test]
    fn brand_unknown_vid_falls_back_to_manufacturer_string() {
        assert_eq!(classify_thermal_brand(Some(0xffff), Some("BEMATECH Printer")), "bematech");
    }

    #[test]
    fn brand_unknown_vid_and_no_mfr_returns_unknown() {
        assert_eq!(classify_thermal_brand(Some(0xffff), None), "unknown");
    }

    #[test]
    fn brand_no_vid_uses_manufacturer_string() {
        assert_eq!(classify_thermal_brand(None, Some("ELGIN POS")), "elgin");
        assert_eq!(classify_thermal_brand(None, Some("NCR Corporation")), "ncr");
        assert_eq!(classify_thermal_brand(None, Some("Zebra Technologies")), "zebra");
    }
}
