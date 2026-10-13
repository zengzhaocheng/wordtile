import { describe, expect, it } from "vitest";
import { useGameStore } from "../game";
import type { GameStartResponse } from "../../types";

const session: GameStartResponse = {
  session_id: "s1",
  mode: "custom",
  level: 0,
  deck_id: "d1",
  deck_name: "测试题库",
  block_count: 2,
  terms: [
    { token: "t1", text: "苹果" },
    { token: "t2", text: "猫" },
  ],
  definitions: [
    { token: "d1", text: "cat" },
    { token: "d2", text: "apple" },
  ],
};

describe("game store", () => {
  it("startGame 重置状态", () => {
    useGameStore.getState().startGame(session);
    const st = useGameStore.getState();
    expect(st.session?.session_id).toBe("s1");
    expect(st.correctCount).toBe(0);
    expect(st.selectedTerm).toBeNull();
  });

  it("正确匹配会消除并累计计数", () => {
    useGameStore.getState().startGame(session);
    const store = useGameStore.getState();
    store.selectTerm("t1");
    expect(useGameStore.getState().selectedTerm).toBe("t1");

    useGameStore.getState().applyMatch("t1", "d2", {
      correct: true,
      correct_count: 1,
      wrong_count: 0,
      finished: false,
    });
    const st = useGameStore.getState();
    expect(st.matchedTermTokens).toContain("t1");
    expect(st.matchedDefTokens).toContain("d2");
    expect(st.correctCount).toBe(1);
    expect(st.selectedTerm).toBeNull();
  });

  it("错误匹配不消除，设置错误对", () => {
    useGameStore.getState().startGame(session);
    useGameStore.getState().selectTerm("t1");
    useGameStore.getState().setWrong("t1", "d1");
    const st = useGameStore.getState();
    expect(st.wrongPair).toEqual({ term: "t1", def: "d1" });
    expect(st.matchedTermTokens).toHaveLength(0);

    useGameStore.getState().clearWrong();
    expect(useGameStore.getState().wrongPair).toBeNull();
  });

  it("再次点击同一词条可取消选中", () => {
    useGameStore.getState().startGame(session);
    useGameStore.getState().selectTerm("t2");
    useGameStore.getState().selectTerm("t2");
    expect(useGameStore.getState().selectedTerm).toBeNull();
  });
});
