export default function useColors() {
  return {
    // Legacy colors (for backward compatibility)
    primary: '#040428',
    secondary: '#4485c4',
    accent: '#03AED2',
    darkColor: '#016B63',
    background: '#ffffff',
    bgColor: '#ffffff',
    light: '#fff',
    dark: '#000',
    itemBackground: '#373737',
    textColor: '#fff',
    subTextColor: '#B0B0B0',
    lightGray: '#D9D9D9',

    // New Design System Colors
    button: {
      primary: {
        active: {
          background: '#ffffff',
          icon: '#151515',
          text: '#15151500',
        },
        default: {
          background: '#e9ebdf',
          icon: '#151515',
          text: '#151515',
        },
        hover: {
          background: '#ffffff',
          icon: '#151515',
          text: '#151515',
        },
      },
      secondary: {
        blue: {
          active: { icon: '#e9ebdf', stroke: '#9cbfe7', text: '#e9ebdf00' },
          default: { icon: '#e9ebdf', stroke: '#518dd2', text: '#e9ebdf' },
          hover: { icon: '#e9ebdf', stroke: '#9cbfe7', text: '#e9ebdf' },
        },
        gray: {
          active: { icon: '#e9ebdf', stroke: '#c8bfb5', text: '#e9ebdf00' },
          default: { icon: '#e9ebdf', stroke: '#8b867f', text: '#e9ebdf' },
          hover: { icon: '#e9ebdf', stroke: '#c8bfb5', text: '#e9ebdf' },
        },
        green: {
          active: { icon: '#e9ebdf', stroke: '#9dbdb6', text: '#e9ebdf00' },
          default: { icon: '#e9ebdf', stroke: '#4d9987', text: '#e9ebdf' },
          hover: { icon: '#e9ebdf', stroke: '#9dbdb6', text: '#e9ebdf' },
        },
        orange: {
          active: { icon: '#e9ebdf', stroke: '#ecafa2', text: '#e9ebdf00' },
          default: { icon: '#e9ebdf', stroke: '#e8765e', text: '#e9ebdf' },
          hover: { icon: '#e9ebdf', stroke: '#ecafa2', text: '#e9ebdf' },
        },
        pink: {
          active: { icon: '#e9ebdf', stroke: '#e09ee2', text: '#e9ebdf00' },
          default: { icon: '#e9ebdf', stroke: '#cc64ce', text: '#e9ebdf' },
          hover: { icon: '#e9ebdf', stroke: '#e09ee2', text: '#e9ebdf' },
        },
        purple: {
          active: { icon: '#e9ebdf', stroke: '#bb9fe7', text: '#e9ebdf00' },
          default: { icon: '#e9ebdf', stroke: '#9874d2', text: '#e9ebdf' },
          hover: { icon: '#e9ebdf', stroke: '#bb9fe7', text: '#e9ebdf' },
        },
        yellow: {
          active: { icon: '#e9ebdf', stroke: '#d4ba95', text: '#e9ebdf00' },
          default: { icon: '#e9ebdf', stroke: '#eca438', text: '#e9ebdf' },
          hover: { icon: '#e9ebdf', stroke: '#d4ba95', text: '#e9ebdf' },
        },
      },
      tertiary: {
        active: { background: '#2e2f2d', icon: '#e9ebdf', text: '#e9ebdf00' },
        default: { background: '#242424', icon: '#e9ebdf', text: '#e9ebdf' },
        hover: { background: '#2e2f2d', icon: '#e9ebdf', text: '#e9ebdf' },
      },
    },

    surface: {
      accent: {
        background: {
          blue: '#1b2e44',
          gray: '#282522',
          green: '#0e352c',
          orange: '#491f16',
          pink: '#3d163d',
          purple: '#3a2b53',
          yellow: '#4a2b11',
        },
        base: {
          blue: '#2d4c71',
          gray: '#433e38',
          green: '#185849',
          orange: '#793325',
          pink: '#652466',
          purple: '#53397c',
          yellow: '#7c481c',
        },
        focused: {
          blue: '#518dd2',
          gray: '#8b867f',
          green: '#4d9987',
          orange: '#e8765e',
          pink: '#cc64ce',
          purple: '#9874d2',
          yellow: '#eca438',
        },
        muted: {
          blue: '#9cbfe7',
          gray: '#c8bfb5',
          green: '#9dbdb6',
          orange: '#ecafa2',
          pink: '#e09ee2',
          purple: '#bb9fe7',
          yellow: '#d4ba95',
        },
      },
      background: {
        base: '#151515',
        focused: '#242424',
        muted: '#0e0e0e',
      },
      lightbox: {
        active: '#000000cc',
        inactive: '#15151500',
      },
      details: {
        neutral: '#e9ebdf',
        neutralAlpha6: '#e9ebdf10',
        neutralAlpha12: '#e9ebdf1f',
        neutralAlpha20: '#e9ebdf33',
        neutralAlpha40: '#e9ebdf66',
        neutralAlpha60: '#e9ebdf99',
        neutralMuted: '#cbccc4',
        headerBg: '#ffffff',
        logoBg: '#e9ebdf',
      },
      text: {
        muted: '#cbccc4',
        opacityAlpha20: '#e9ebdf33',
        opacityAlpha40: '#e9ebdf66',
        opacityAlpha60: '#e9ebdf99',
        primary: '#e9ebdf',
      },
    },

    input: {
      default: {
        label: { text: '#e9ebdf', required: '#e9ebdf99' },
        value: { text: '#e9ebdf' },
        placeholder: { text: '#e9ebdf99' },
        error: { text: '#f15264', stroke: '#f15264' },
        stroke: '#e9ebdf10',
        background: '#e9ebdf10',
      },
      hover: {
        label: { text: '#e9ebdf', required: '#e9ebdf99' },
        value: { text: '#e9ebdf' },
        placeholder: { text: '#e9ebdf99' },
        error: { text: '#f15264', stroke: '#f15264' },
        stroke: '#e9ebdf1f',
        background: '#e9ebdf33',
      },
      focused: {
        label: { text: '#e9ebdf', required: '#e9ebdf99' },
        value: { text: '#151515' },
        placeholder: { text: '#15151599' },
        error: { text: '#c72844', stroke: '#c72844' },
        stroke: '#15151500',
        background: '#f7f8f4',
      },
    },

    navigation: {
      background: {
        scrolled: '#000000cc',
        active: '#000000cc',
        default: '#15151500',
      },
      items: {
        active: {
          icon: '#cbccc4',
          selected: '#518dd2',
          strokeAccent: '#e9ebdf66',
          text: '#cbccc4',
        },
        default: {
          icon: '#e9ebdf',
          selected: '#518dd2',
          strokeAccent: '#e9ebdf66',
          text: '#e9ebdf',
        },
        hover: {
          icon: '#f7f8f4',
          selected: '#518dd2',
          strokeAccent: '#e9ebdf66',
          text: '#f7f8f4',
        },
        inactive: {
          icon: '#e9ebdf66',
          selected: '#518dd2',
          strokeAccent: '#e9ebdf66',
          text: '#e9ebdf66',
        },
      },
      subnav: {
        default: {
          icon: '#e9ebdf00',
          selected: '#518dd2',
          strokeAccent: '#e9ebdf1f',
          text: '#e9ebdf',
        },
        hover: {
          icon: '#e9ebdf',
          selected: '#518dd2',
          strokeAccent: '#e9ebdf1f',
          text: '#e9ebdf',
        },
        inactive: {
          icon: '#e9ebdf00',
          selected: '#518dd2',
          strokeAccent: '#e9ebdf1f',
          text: '#e9ebdf66',
        },
      },
    },

    interface: {
      background: '#f7f8f499',
      stroke: '#1515151f',
    },

    eyebrow: {
      background: '#000000',
      text: '#b6b8af',
    },

    footer: {
      background: '#151515',
      text: '#cbccc4',
      logo: '#e9ebdf',
      border: '#e9ebdf99',
    },

    alert: {
      default: {
        icon: '#e79295',
        stroke: '#c72844',
        background: '#651722',
      },
    },

    cursor: {
      active: {
        contextGuide: '#15151599',
        contextGuideStroke: '#c8bfb5',
        cursor: '#ffffff',
        cursorStroke: '#ffffff',
        icon: '#ffffff',
        text: '#e9ebdf',
      },
      default: {
        contextGuide: '#ffffff',
        contextGuideStroke: '#000000',
        cursor: '#ffffff',
        cursorStroke: '#000000',
        icon: '#ffffff',
        text: '#ffffff',
      },
      hover: {
        contextGuide: '#00000099',
        contextGuideStroke: '#8b867f',
        cursor: '#ffffff',
        cursorStroke: '#000000',
        icon: '#ffffff',
        text: '#ffffff',
      },
      message: {
        contextGuide: '#00000099',
        contextGuideStroke: '#e9ebdf00',
        cursor: '#ffffff',
        cursorStroke: '#000000',
        text: '#ffffff',
      },
    },

    experimental: {
      cursorFill: '#000000',
      cursorStroke: '#ffffff',
      uiBackground: '#ffffff',
    },

    transparent: 'transparent',
  };
}
