// The ribbon and lettering follow `currentColor`; the fold takes the active app's accent.
const RIBBON =
  'M342.3 0L378.2 39.8C414 87.6 414 175.2 378.2 203L274.7 294.6C238.8 326.4 226.9 378.2 242.8 422L378.2 577.2C414 621 414 696.6 378.2 728.5L266.7 824C282.6 792.2 282.6 744.4 262.7 720.5L43.8 469.7C15.9 437.9 0 398.1 0 358.3C0 318.5 19.9 278.7 55.7 246.8Z';
const FOLD =
  'M171.2 147.3L342.3 0L378.2 39.8C414 87.6 414 175.2 378.2 203L274.7 294.6C238.8 326.4 226.9 378.2 242.8 422L195.1 366.2C163.2 326.4 143.3 262.7 155.2 207Z';
const WORDMARK =
  'M435 299H479L484 330C492 312 515 295 555 295V348H540C504 348 484 360 484 400V489H435ZM613 247H658V299H708V342H658V425C658 440 664 446 680 446H708V489H672C630 489 608 470 608 430V342H575V299H608ZM735 299H790L845 419L892 299H946L829 572H776L820 468H810ZM979 218H1030V489H979ZM1079 299H1123L1128 330C1136 312 1159 295 1199 295V348H1184C1148 348 1128 360 1128 400V489H1079Z';

interface RtylrLogoProps {
  className?: string;
}

export default function RtylrLogo({ className }: RtylrLogoProps) {
  return (
    <svg
      className={['rtylr-logo', className].filter(Boolean).join(' ')}
      viewBox="0 0 1200 828"
      role="img"
      aria-label="RTYLR"
    >
      <path fill="currentColor" d={RIBBON} />
      <path fill="var(--app-logo-fold, #4583c4)" d={FOLD} />
      <path fill="currentColor" d={WORDMARK} />
    </svg>
  );
}
