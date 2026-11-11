import {
  Car,
  Castle,
  Compass,
  Crosshair,
  Gamepad,
  Gamepad2,
  Ghost,
  Gift,
  Joystick,
  Monitor,
  Plane,
  Puzzle,
  Sparkles,
  Swords,
  Tag,
  Trophy,
  Zap,
  type LucideIcon,
} from "lucide-react";

// The API names a lucide icon for each category. Only the icons the seed uses
// are imported, since importing lucide's full name map would ship every icon
// to the browser. An unknown name falls back to Tag, so a new category never
// breaks a block.
const ICONS: Record<string, LucideIcon> = {
  Car, Castle, Compass, Crosshair, Gamepad, Gamepad2, Ghost, Gift,
  Joystick, Monitor, Plane, Puzzle, Sparkles, Swords, Trophy, Zap,
};

export function CategoryIcon({ name, className }: { name: string; className?: string }) {
  const Icon = ICONS[name] ?? Tag;
  return <Icon className={className} strokeWidth={1.5} aria-hidden />;
}
