interface RtylrWordmarkProps {
  className?: string;
  suffix?: string;
}

export function RtylrWordmark({ className = '', suffix }: RtylrWordmarkProps) {
  return (
    <span
      aria-label={suffix ? `rtylr ${suffix}` : 'rtylr'}
      className={`inline-flex items-center ${className}`}
    >
      <img src="/logo-light.png" alt="" className="h-[1.35em] w-auto object-contain" />
    </span>
  );
}
