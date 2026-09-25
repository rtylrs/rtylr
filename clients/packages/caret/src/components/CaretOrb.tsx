import React from "react";
import { ThinkingOrb, type OrbSize, type OrbState } from "thinking-orbs";

export type CaretOrbState = OrbState;

export interface CaretOrbProps {
  state: CaretOrbState;
  size?: OrbSize;
  speed?: number;
  paused?: boolean;
  className?: string;
  style?: React.CSSProperties;
  "aria-label"?: string;
}

/** Let the orb resolve its ink from the host panel's light or dark theme. */
export const CaretOrb: React.FC<CaretOrbProps> = ({
  state,
  size = 20,
  ...rest
}) => <ThinkingOrb state={state} size={size} theme="auto" {...rest} />;

const WRITE_ACTION =
  /\.(create|update|delete|set|record|approve|receive|refund|reply|send|schedule|publish|invite)/;
const EXTERNAL_ACTION =
  /campaign|email|store\.|support\.|schedule|publish|invite|[._]send/;
const ANALYSIS_READ =
  /report|insight|analytic|reconcil|journal|tax|forecast|dashboard|margin|revenue|cash_flow|aging/;
const PLAIN_READ = /\.list_|\.get_|search|snapshot|map_/;

export function orbStateForTool(name: string): CaretOrbState {
  const tool = name.trim().toLowerCase();
  if (tool === "memory.remember") return "weaving";
  if (WRITE_ACTION.test(tool)) {
    return EXTERNAL_ACTION.test(tool) ? "connecting" : "working";
  }
  if (ANALYSIS_READ.test(tool)) return "solving";
  if (PLAIN_READ.test(tool)) return "searching";
  return "working";
}

/** Inline status orb for a streaming assistant turn. */
export const StreamingOrb: React.FC<{ hasText: boolean }> = ({ hasText }) => (
  <span
    style={{
      display: "inline-flex",
      verticalAlign: "text-bottom",
      marginLeft: 4,
    }}
  >
    <CaretOrb state={hasText ? "composing" : "breathing"} />
  </span>
);

export default CaretOrb;
