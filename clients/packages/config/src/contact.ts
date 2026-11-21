export type RtylrContactChannelId = "whatsapp" | "phone" | "email";

export interface RtylrContactChannel {
  readonly id: RtylrContactChannelId;
  readonly href: string;
  readonly value: string;
}

/** Public RTYLR / Voxire contact channels shared across marketing and Rennili. */
export const RTYLR_CONTACT_CHANNELS: readonly RtylrContactChannel[] = [
  { id: "whatsapp", href: "https://wa.me/9613940708", value: "+961 3 940 708" },
  { id: "phone", href: "tel:+9613940708", value: "+961 3 940 708" },
  { id: "email", href: "mailto:info@voxire.com", value: "info@voxire.com" },
] as const;
