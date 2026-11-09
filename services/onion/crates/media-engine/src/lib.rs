//! Rust port of the `media-engine` primitives and `--serve` framing.
//!
//! The C++ engine (`media-engine/`, libvips) is the execution sidecar; Go
//! talks to it over framed stdio (`internal/platform/cdn/engine.go`):
//! requests are `u32le length + NUL-separated argv tokens`, responses are
//! `u32le exit code + u32le length + msg bytes` (capped at 4 MiB).
//!
//! This crate ports the vips-free primitives with golden tests matching the
//! C++ `*_test.cpp` vectors, plus the framing codec. Pixel ops that need
//! libvips stay in the C++ engine; the Rust `integrations` crate will wire
//! subprocess calls to the same `--serve` protocol.

use thiserror::Error;

#[derive(Debug, PartialEq, Error)]
#[error("{0}")]
pub struct ProcessingError(pub String);

impl ProcessingError {
    fn msg(text: impl Into<String>) -> Self {
        Self(text.into())
    }
}

/// Width/height in pixels; 0 means unset.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub struct Size {
    pub width: i32,
    pub height: i32,
}

impl Size {
    pub fn empty(&self) -> bool {
        self.width == 0 && self.height == 0
    }

    pub fn area(&self) -> i64 {
        i64::from(self.width) * i64::from(self.height)
    }
}

/// Axis-aligned rectangle with top-left origin.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub struct Box {
    pub x: i32,
    pub y: i32,
    pub width: i32,
    pub height: i32,
}

impl Box {
    pub fn right(&self) -> i32 {
        self.x + self.width
    }

    pub fn bottom(&self) -> i32 {
        self.y + self.height
    }

    pub fn empty(&self) -> bool {
        self.width <= 0 || self.height <= 0
    }
}

/// Parse `WxH`, `Wx` or `xH`; missing parts are 0.
pub fn parse_size(text: &str) -> Result<Size, ProcessingError> {
    let (width_text, height_text) = text
        .split_once('x')
        .ok_or_else(|| ProcessingError::msg(format!("invalid size (expected WxH): {text}")))?;
    let mut size = Size {
        width: 0,
        height: 0,
    };
    if !width_text.is_empty() {
        size.width = width_text
            .parse::<i32>()
            .map_err(|_| ProcessingError::msg(format!("invalid size (expected WxH): {text}")))?;
    }
    if !height_text.is_empty() {
        size.height = height_text
            .parse::<i32>()
            .map_err(|_| ProcessingError::msg(format!("invalid size (expected WxH): {text}")))?;
    }
    if size.width < 0 || size.height < 0 {
        return Err(ProcessingError::msg(format!(
            "size dimensions must not be negative: {text}"
        )));
    }
    Ok(size)
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum FitMode {
    Cover,
    Contain,
    Fill,
    Inside,
    Outside,
}

pub fn parse_fit_mode(text: &str) -> Result<FitMode, ProcessingError> {
    match text.to_ascii_lowercase().as_str() {
        "cover" => Ok(FitMode::Cover),
        "contain" => Ok(FitMode::Contain),
        "fill" => Ok(FitMode::Fill),
        "inside" => Ok(FitMode::Inside),
        "outside" => Ok(FitMode::Outside),
        other => Err(ProcessingError::msg(format!("unknown fit mode: {other}"))),
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Gravity {
    Center,
    North,
    South,
    East,
    West,
    NorthEast,
    NorthWest,
    SouthEast,
    SouthWest,
}

pub fn parse_gravity(text: &str) -> Result<Gravity, ProcessingError> {
    match text {
        "center" | "centre" => Ok(Gravity::Center),
        "north" | "top" => Ok(Gravity::North),
        "south" | "bottom" => Ok(Gravity::South),
        "east" | "right" => Ok(Gravity::East),
        "west" | "left" => Ok(Gravity::West),
        "north_east" | "top_right" => Ok(Gravity::NorthEast),
        "north_west" | "top_left" => Ok(Gravity::NorthWest),
        "south_east" | "bottom_right" => Ok(Gravity::SouthEast),
        "south_west" | "bottom_left" => Ok(Gravity::SouthWest),
        other => Err(ProcessingError::msg(format!("unknown gravity: {other}"))),
    }
}

/// Top-left offset for placing an inner box inside an outer one.
pub fn offset_for(
    gravity: Gravity,
    inner_w: i32,
    inner_h: i32,
    outer_w: i32,
    outer_h: i32,
) -> (i32, i32) {
    let center_x = (outer_w - inner_w) / 2;
    let center_y = (outer_h - inner_h) / 2;
    let right = outer_w - inner_w;
    let bottom = outer_h - inner_h;
    match gravity {
        Gravity::Center => (center_x, center_y),
        Gravity::North => (center_x, 0),
        Gravity::South => (center_x, bottom),
        Gravity::East => (right, center_y),
        Gravity::West => (0, center_y),
        Gravity::NorthEast => (right, 0),
        Gravity::NorthWest => (0, 0),
        Gravity::SouthEast => (right, bottom),
        Gravity::SouthWest => (0, bottom),
    }
}

/// 8-bit RGBA color.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub struct Color {
    pub r: i32,
    pub g: i32,
    pub b: i32,
    pub a: i32,
}

impl Color {
    fn new(r: i32, g: i32, b: i32, a: i32) -> Self {
        Self { r, g, b, a }
    }

    pub fn bands(&self) -> Vec<f64> {
        if self.a >= 255 {
            vec![f64::from(self.r), f64::from(self.g), f64::from(self.b)]
        } else {
            vec![
                f64::from(self.r),
                f64::from(self.g),
                f64::from(self.b),
                f64::from(self.a),
            ]
        }
    }
}

pub fn named_colors() -> Vec<(&'static str, Color)> {
    vec![
        ("black", Color::new(0, 0, 0, 255)),
        ("white", Color::new(255, 255, 255, 255)),
        ("transparent", Color::new(0, 0, 0, 0)),
        ("red", Color::new(255, 0, 0, 255)),
        ("green", Color::new(0, 128, 0, 255)),
        ("lime", Color::new(0, 255, 0, 255)),
        ("blue", Color::new(0, 0, 255, 255)),
        ("yellow", Color::new(255, 255, 0, 255)),
        ("cyan", Color::new(0, 255, 255, 255)),
        ("aqua", Color::new(0, 255, 255, 255)),
        ("magenta", Color::new(255, 0, 255, 255)),
        ("fuchsia", Color::new(255, 0, 255, 255)),
        ("orange", Color::new(255, 165, 0, 255)),
        ("purple", Color::new(128, 0, 128, 255)),
        ("pink", Color::new(255, 192, 203, 255)),
        ("brown", Color::new(165, 42, 42, 255)),
        ("maroon", Color::new(128, 0, 0, 255)),
        ("navy", Color::new(0, 0, 128, 255)),
        ("olive", Color::new(128, 128, 0, 255)),
        ("teal", Color::new(0, 128, 128, 255)),
        ("gray", Color::new(128, 128, 128, 255)),
        ("grey", Color::new(128, 128, 128, 255)),
    ]
}

fn hex_pair(text: &str, input: &str) -> Result<i32, ProcessingError> {
    let bytes = text.as_bytes();
    if bytes.len() < 2 {
        return Err(ProcessingError::msg(format!(
            "invalid hex color digit in: {input}"
        )));
    }
    let value = |byte: u8| -> i32 {
        match byte {
            b'0'..=b'9' => i32::from(byte - b'0'),
            b'a'..=b'f' => i32::from(byte - b'a' + 10),
            b'A'..=b'F' => i32::from(byte - b'A' + 10),
            _ => -1,
        }
    };
    let hi = value(bytes[0]);
    let lo = value(bytes[1]);
    if hi < 0 || lo < 0 {
        return Err(ProcessingError::msg(format!(
            "invalid hex color digit in: {input}"
        )));
    }
    Ok(hi * 16 + lo)
}

pub fn parse_color(input: &str) -> Result<Color, ProcessingError> {
    for (name, color) in named_colors() {
        if input == name {
            return Ok(color);
        }
    }
    let text = input.strip_prefix('#').unwrap_or(input);
    match text.len() {
        3 => Ok(Color {
            r: hex_pair(&text[0..1].repeat(2), input)?,
            g: hex_pair(&text[1..2].repeat(2), input)?,
            b: hex_pair(&text[2..3].repeat(2), input)?,
            a: 255,
        }),
        4 => Ok(Color {
            r: hex_pair(&text[0..1].repeat(2), input)?,
            g: hex_pair(&text[1..2].repeat(2), input)?,
            b: hex_pair(&text[2..3].repeat(2), input)?,
            a: hex_pair(&text[3..4].repeat(2), input)?,
        }),
        6 => Ok(Color {
            r: hex_pair(&text[0..2], input)?,
            g: hex_pair(&text[2..4], input)?,
            b: hex_pair(&text[4..6], input)?,
            a: 255,
        }),
        8 => Ok(Color {
            r: hex_pair(&text[0..2], input)?,
            g: hex_pair(&text[2..4], input)?,
            b: hex_pair(&text[4..6], input)?,
            a: hex_pair(&text[6..8], input)?,
        }),
        _ => Err(ProcessingError::msg(format!(
            "invalid color (expected hex or named): {input}"
        ))),
    }
}

pub fn clamp_int(value: i32, low: i32, high: i32) -> i32 {
    value.clamp(low, high)
}

pub fn clamp_double(value: f64, low: f64, high: f64) -> f64 {
    value.clamp(low, high)
}

pub fn parse_int_strict(text: &str, field: &str) -> Result<i32, ProcessingError> {
    if text.is_empty() {
        return Err(ProcessingError::msg(format!(
            "{field} must be an integer, got empty value"
        )));
    }
    text.parse::<i32>()
        .map_err(|_| ProcessingError::msg(format!("{field} must be an integer, got: {text}")))
}

pub fn parse_double_strict(text: &str, field: &str) -> Result<f64, ProcessingError> {
    if text.is_empty() {
        return Err(ProcessingError::msg(format!(
            "{field} must be a number, got empty value"
        )));
    }
    text.parse::<f64>()
        .map_err(|_| ProcessingError::msg(format!("{field} must be a number, got: {text}")))
}

/// Response cap, mirroring `maxEngineResponse` in `engine.go`.
pub const MAX_ENGINE_RESPONSE: usize = 4 << 20;

/// Encode argv tokens as a request frame.
pub fn frame_job(tokens: &[&str]) -> Vec<u8> {
    let payload_len: usize =
        tokens.iter().map(|t| t.len()).sum::<usize>() + tokens.len().saturating_sub(1);
    let mut frame = Vec::with_capacity(4 + payload_len);
    frame.extend_from_slice(&(payload_len as u32).to_le_bytes());
    for (index, token) in tokens.iter().enumerate() {
        if index > 0 {
            frame.push(0);
        }
        frame.extend_from_slice(token.as_bytes());
    }
    frame
}

pub fn encode_response(code: u32, message: &str) -> Vec<u8> {
    let bytes = message.as_bytes();
    let mut frame = Vec::with_capacity(8 + bytes.len());
    frame.extend_from_slice(&code.to_le_bytes());
    frame.extend_from_slice(&(bytes.len() as u32).to_le_bytes());
    frame.extend_from_slice(bytes);
    frame
}

pub fn decode_response(frame: &[u8]) -> Result<(u32, String), ProcessingError> {
    if frame.len() < 8 {
        return Err(ProcessingError::msg("engine response header truncated"));
    }
    let code = u32::from_le_bytes([frame[0], frame[1], frame[2], frame[3]]);
    let len = u32::from_le_bytes([frame[4], frame[5], frame[6], frame[7]]) as usize;
    if len > MAX_ENGINE_RESPONSE {
        return Err(ProcessingError::msg(format!(
            "engine response too large ({len} bytes)"
        )));
    }
    if frame.len() < 8 + len {
        return Err(ProcessingError::msg("engine response body truncated"));
    }
    let message = String::from_utf8(frame[8..8 + len].to_vec())
        .map_err(|_| ProcessingError::msg("engine response is not UTF-8"))?;
    Ok((code, message))
}

#[cfg(test)]
mod tests {
    use super::{
        clamp_double, clamp_int, decode_response, encode_response, frame_job, named_colors,
        offset_for, parse_color, parse_double_strict, parse_fit_mode, parse_gravity,
        parse_int_strict, parse_size, Box, Color, FitMode, Gravity, Size,
    };

    #[test]
    fn geometry_goldens() {
        assert_eq!(
            parse_size("256x128"),
            Ok(Size {
                width: 256,
                height: 128
            })
        );
        assert_eq!(
            parse_size("256x"),
            Ok(Size {
                width: 256,
                height: 0
            })
        );
        assert_eq!(
            parse_size("x128"),
            Ok(Size {
                width: 0,
                height: 128
            })
        );
        assert!(Size {
            width: 0,
            height: 0
        }
        .empty());
        assert_eq!(
            Size {
                width: 4,
                height: 5
            }
            .area(),
            20
        );
        assert!(parse_size("nope").is_err());
        assert!(parse_size("-1x10").is_err());
        let boxed = Box {
            x: 2,
            y: 3,
            width: 10,
            height: 20,
        };
        assert_eq!(boxed.right(), 12);
        assert_eq!(boxed.bottom(), 23);
        assert!(!boxed.empty());
        assert!(Box {
            x: 0,
            y: 0,
            width: 0,
            height: 5
        }
        .empty());
        assert_eq!(parse_fit_mode("COVER"), Ok(FitMode::Cover));
        assert!(parse_fit_mode("nope").is_err());
    }

    #[test]
    fn gravity_goldens() {
        assert_eq!(parse_gravity("center"), Ok(Gravity::Center));
        assert_eq!(parse_gravity("top"), Ok(Gravity::North));
        assert_eq!(parse_gravity("bottom_right"), Ok(Gravity::SouthEast));
        assert!(parse_gravity("sideways").is_err());
        assert_eq!(offset_for(Gravity::Center, 100, 100, 300, 300), (100, 100));
        assert_eq!(offset_for(Gravity::NorthWest, 100, 100, 300, 300), (0, 0));
        assert_eq!(
            offset_for(Gravity::SouthEast, 100, 100, 300, 300),
            (200, 200)
        );
        assert_eq!(offset_for(Gravity::North, 100, 100, 300, 300), (100, 0));
    }

    #[test]
    fn color_goldens() {
        let hex6 = parse_color("#ff8800").unwrap();
        assert_eq!((hex6.r, hex6.g, hex6.b, hex6.a), (255, 136, 0, 255));
        assert_eq!(parse_color("ff8800").unwrap().r, 255);
        let hex8 = parse_color("#11223380").unwrap();
        assert_eq!((hex8.r, hex8.a), (0x11, 0x80));
        let short = parse_color("#f80").unwrap();
        assert_eq!((short.r, short.g, short.b), (255, 136, 0));
        assert_eq!(parse_color("black").unwrap().r, 0);
        assert_eq!(parse_color("white").unwrap().r, 255);
        assert_eq!(parse_color("transparent").unwrap().a, 0);
        assert_eq!(parse_color("red").unwrap().r, 255);
        assert_eq!(parse_color("navy").unwrap().b, 128);
        assert_eq!(
            parse_color("grey").unwrap().r,
            parse_color("gray").unwrap().r
        );
        assert_eq!(parse_color("white").unwrap().bands().len(), 3);
        assert_eq!(parse_color("#00000080").unwrap().bands().len(), 4);
        assert!(parse_color("#ggg").is_err());
        assert!(parse_color("nope").is_err());
        assert_eq!(named_colors().len(), 22);
    }

    #[test]
    fn numeric_goldens() {
        assert_eq!(clamp_int(5, 0, 10), 5);
        assert_eq!(clamp_int(-3, 0, 10), 0);
        assert_eq!(clamp_int(99, 0, 10), 10);
        assert_eq!(clamp_double(0.5, 0.0, 1.0), 0.5);
        assert_eq!(clamp_double(2.0, 0.0, 1.0), 1.0);
        assert_eq!(clamp_double(-1.0, 0.0, 1.0), 0.0);
        assert_eq!(parse_int_strict("42", "x"), Ok(42));
        assert_eq!(parse_int_strict("-7", "x"), Ok(-7));
        assert!(parse_int_strict("12abc", "x").is_err());
        assert!(parse_int_strict("", "x").is_err());
        assert_eq!(parse_double_strict("1.5", "x"), Ok(1.5));
        assert!(parse_double_strict("1.5x", "x").is_err());
        assert_eq!(
            parse_color("red").unwrap(),
            Color {
                r: 255,
                g: 0,
                b: 0,
                a: 255
            }
        );
    }

    #[test]
    fn framing_goldens() {
        let frame = frame_job(&["resize", "256x256", "input.png"]);
        assert_eq!(&frame[4..], b"resize\x00256x256\x00input.png");
        let (code, message) = decode_response(&encode_response(0, "ok")).unwrap();
        assert_eq!((code, message.as_str()), (0, "ok"));
        let (code, message) = decode_response(&encode_response(3, "boom")).unwrap();
        assert_eq!((code, message.as_str()), (3, "boom"));
        assert!(decode_response(&[0, 0, 0]).is_err());
        let mut big = Vec::new();
        big.extend_from_slice(&0u32.to_le_bytes());
        big.extend_from_slice(&u32::MAX.to_le_bytes());
        assert!(decode_response(&big).is_err());
    }
}
