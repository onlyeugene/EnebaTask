// The 404 page's background: our own pixel shapes (a key, a gem, a heart, a
// plus), each drawn from a tiny grid where "#" is one filled square.
// Decoration only, so screen readers skip it and clicks pass through.
const SHAPES = {
  key: [
    ".###........",
    "#...########",
    "#...#...#.#.",
    ".###........",
  ],
  gem: ["..#..", ".###.", "#####", ".###.", "..#.."],
  heart: [".##.##.", "#######", "#######", ".#####.", "..###..", "...#..."],
  plus: [".#.", "###", ".#."],
};

type Name = keyof typeof SHAPES;

// Where each shape sits (percent of the area), its size in px per square,
// its color token and when its float starts, so they don't move in step.
// Most sit low and to the right, as on Eneba, away from the text.
const PLACED: { shape: Name; left: number; top: number; px: number; color: string; delay: number }[] = [
  { shape: "key", left: 38, top: 62, px: 9, color: "text-yellow", delay: 0 },
  { shape: "gem", left: 52, top: 48, px: 11, color: "text-mint", delay: 1.2 },
  { shape: "heart", left: 63, top: 70, px: 8, color: "text-pink", delay: 2.4 },
  { shape: "plus", left: 70, top: 55, px: 9, color: "text-lime", delay: 0.6 },
  { shape: "gem", left: 79, top: 76, px: 7, color: "text-brand-light", delay: 3 },
  { shape: "key", left: 74, top: 30, px: 6, color: "text-teal", delay: 1.8 },
  { shape: "plus", left: 88, top: 42, px: 12, color: "text-yellow", delay: 4.2 },
  { shape: "heart", left: 22, top: 82, px: 7, color: "text-mint", delay: 3.6 },
  { shape: "plus", left: 47, top: 86, px: 6, color: "text-pink", delay: 2 },
  { shape: "gem", left: 90, top: 14, px: 6, color: "text-lime", delay: 5 },
];

export function PixelShapes() {
  return (
    <div aria-hidden className="pointer-events-none absolute inset-0 overflow-hidden">
      {PLACED.map((p, i) => (
        <div
          key={i}
          className={`absolute opacity-80 motion-safe:animate-float ${p.color}`}
          style={{ left: `${p.left}%`, top: `${p.top}%`, animationDelay: `${p.delay}s` }}
        >
          <Pixels rows={SHAPES[p.shape]} px={p.px} />
        </div>
      ))}
    </div>
  );
}

// One square per "#". currentColor, so the wrapper's text color paints it.
// crispEdges keeps the squares sharp at any size.
function Pixels({ rows, px }: { rows: string[]; px: number }) {
  const w = rows[0].length;
  const h = rows.length;
  return (
    <svg width={w * px} height={h * px} viewBox={`0 0 ${w} ${h}`} shapeRendering="crispEdges" fill="currentColor">
      {rows.flatMap((row, y) =>
        [...row].map((c, x) => (c === "#" ? <rect key={`${x}-${y}`} x={x} y={y} width={1} height={1} /> : null)),
      )}
    </svg>
  );
}
