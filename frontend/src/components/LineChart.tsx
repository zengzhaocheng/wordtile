interface Point {
  label: string;
  value: number;
}

/** 轻量 SVG 折线图（无第三方图表依赖，双主题友好） */
export default function LineChart({
  points,
  height = 160,
  stroke = "var(--primary)",
}: {
  points: Point[];
  height?: number;
  stroke?: string;
}) {
  if (points.length === 0) {
    return null;
  }
  const width = 600;
  const padX = 28;
  const padY = 16;
  const values = points.map((p) => p.value);
  const min = Math.min(...values);
  const max = Math.max(...values);
  const span = max - min || 1;

  const x = (i: number) => padX + (i * (width - padX * 2)) / Math.max(points.length - 1, 1);
  const y = (v: number) => padY + (1 - (v - min) / span) * (height - padY * 2);

  const path = points.map((p, i) => `${i === 0 ? "M" : "L"}${x(i).toFixed(1)},${y(p.value).toFixed(1)}`).join(" ");

  return (
    <svg viewBox={`0 0 ${width} ${height}`} className="w-full" role="img">
      {/* 网格线与刻度 */}
      {[0, 0.5, 1].map((r) => {
        const gy = padY + r * (height - padY * 2);
        const val = max - r * span;
        return (
          <g key={r}>
            <line x1={padX} x2={width - padX} y1={gy} y2={gy} stroke="var(--line)" strokeWidth="1" />
            <text x={4} y={gy + 4} fontSize="10" fill="var(--muted)">
              {formatTick(val)}
            </text>
          </g>
        );
      })}
      <path d={path} fill="none" stroke={stroke} strokeWidth="2.5" strokeLinejoin="round" strokeLinecap="round" />
      {points.map((p, i) => (
        <circle key={i} cx={x(i)} cy={y(p.value)} r="3" fill={stroke}>
          <title>{`${p.label}: ${formatTick(p.value)}`}</title>
        </circle>
      ))}
      {/* 首末标签 */}
      <text x={padX} y={height - 2} fontSize="10" fill="var(--muted)">
        {points[0]?.label}
      </text>
      <text x={width - padX} y={height - 2} fontSize="10" fill="var(--muted)" textAnchor="end">
        {points[points.length - 1]?.label}
      </text>
    </svg>
  );
}

function formatTick(v: number): string {
  if (v >= 1000) return `${(v / 1000).toFixed(1)}k`;
  return String(Math.round(v * 100) / 100);
}
