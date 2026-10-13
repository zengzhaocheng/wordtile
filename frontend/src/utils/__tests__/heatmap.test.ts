import { describe, expect, it } from "vitest";
import { buildCalendar, levelOf } from "../heatmap";

describe("levelOf", () => {
  it("0 次练习 → 0 级", () => expect(levelOf(0)).toBe(0));
  it("1 次 → 1 级", () => expect(levelOf(1)).toBe(1));
  it("4 次 → 2 级", () => expect(levelOf(4)).toBe(2));
  it("9 次 → 3 级", () => expect(levelOf(9)).toBe(3));
  it("10 次 → 4 级", () => expect(levelOf(10)).toBe(4));
});

describe("buildCalendar", () => {
  const today = new Date();
  const fmt = (d: Date) => d.toISOString().slice(0, 10);

  it("生成 26 周 × 7 天的网格", () => {
    const cal = buildCalendar([], 26);
    expect(cal).toHaveLength(26);
    cal.forEach((col) => expect(col).toHaveLength(7));
  });

  it("包含今天且今天在最后一列", () => {
    const cal = buildCalendar([{ date: fmt(today), sessions: 3, duration_ms: 1000, accuracy: 1 }]);
    const lastCol = cal[cal.length - 1];
    const todayCell = lastCol.find((c) => c.date === fmt(today));
    expect(todayCell).toBeDefined();
    expect(todayCell?.level).toBe(2); // 3 次 → 2 级
  });

  it("未来日期不参与统计", () => {
    const cal = buildCalendar([], 26);
    const flat = cal.flat();
    for (const cell of flat) {
      if (!cell.inRange) expect(cell.level).toBe(0);
    }
  });
});
