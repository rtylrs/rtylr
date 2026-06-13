interface RtylrWordmarkProps {
  className?: string;
  suffix?: string;
}

export function RtylrWordmark({ className = '', suffix }: RtylrWordmarkProps) {
  return (
    <span
      aria-label={suffix ? `rtylr ${suffix}` : 'rtylr'}
      className={`inline-flex items-baseline gap-1 font-extrabold tracking-normal text-pos-text ${className}`}
    >
      <span>rtylr</span>
      {suffix ? <span className="font-semibold text-pos-muted">{suffix}</span> : null}
    </span>
  );
}
