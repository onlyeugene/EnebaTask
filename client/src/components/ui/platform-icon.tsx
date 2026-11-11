import { Gamepad, Gamepad2, Joystick, Monitor, Puzzle, type LucideIcon } from "lucide-react";
import type { Platform } from "@/types/api";

// Generic icons, not brand logos (see notes.md, Branding).
const PLATFORM_ICONS: Record<Platform, LucideIcon> = {
  steam: Gamepad2,
  psn: Gamepad,
  xbox: Joystick,
  nintendo: Puzzle,
  windows: Monitor,
};

export function PlatformIcon({ platform, className }: { platform: Platform; className?: string }) {
  const Icon = PLATFORM_ICONS[platform];
  return <Icon className={className} aria-hidden />;
}
