export interface CalendarCell {
  date: string; // YYYY-MM-DD
  sessions: number;
  durationMs: number;
  accuracy: number;
  level: 0 | 1 | 2 | 3 | 4;
  inRange: boolean;
}

/** 根据练习数据计算色阶（0~4 级） */
export function levelOf(sessions: number): 0 | 1 | 2 | 3 | 4 {
  if (sessions <= 0) return 0;
  if (sessions === 1) return 1;
  if (sessions <= 4) return 2;
  if (sessions <= 9) return 3;
  return 4;
}

function fmt(d: Date): string {
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, "0");
  const day = String(d.getDate()).padStart(2, "0");
  return `${y}-${m}-${day}`;
}

/**
 * 构建最近 weeks 周的日历网格（列 = 周，行 = 周一..周日）。
 * 与 GitHub Contribution Calendar 布局一致。
 */
export function buildCalendar(
  days: { date: string; sessions: number; duration_ms: number; accuracy: number }[],
  weeks = 26,
): CalendarCell[][] {
  const byDate = new Map(days.map((d) => [d.date, d]));

  const today = new Date();
  today.setHours(0, 0, 0, 0);

  // 对齐到周一为行首
  const dow = (today.getDay() + 6) % 7; // 0=周一
  const lastMonday = new Date(today);
  lastMonday.setDate(today.getDate() - dow);

  const columns: CalendarCell[][] = [];
  for (let w = weeks - 1; w >= 0; w--) {
    const col: CalendarCell[] = [];
    for (let d = 0; d < 7; d++) {
      const date = new Date(lastMonday);
      date.setDate(lastMonday.getDate() - w * 7 + d);
      const key = fmt(date);
      const rec = byDate.get(key);
      const inRange = date <= today;
      const sessions = rec?.sessions ?? 0;
      col.push({
        date: key,
        sessions,
        durationMs: rec?.duration_ms ?? 0,
        accuracy: rec?.accuracy ?? 0,
        level: inRange ? levelOf(sessions) : 0,
        inRange,
      });
    }
    columns.push(col);
  }
  return columns;
}
