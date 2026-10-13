import { create } from "zustand";
import type { FinishSummary, GameStartResponse, MatchResponse } from "../types";

interface GameState {
  session: GameStartResponse | null;
  startedAt: number;
  selectedTerm: string | null;
  matchedTermTokens: string[];
  matchedDefTokens: string[];
  wrongPair: { term: string; def: string } | null;
  hintPair: { term: string; def: string } | null;
  correctCount: number;
  wrongCount: number;
  finished: boolean;
  result: FinishSummary | null;

  startGame: (s: GameStartResponse) => void;
  selectTerm: (token: string) => void;
  applyMatch: (termToken: string, defToken: string, res: MatchResponse) => void;
  setWrong: (termToken: string, defToken: string) => void;
  clearWrong: () => void;
  setHint: (pair: { term: string; def: string } | null) => void;
  setResult: (r: FinishSummary) => void;
  reset: () => void;
}

export const useGameStore = create<GameState>((set) => ({
  session: null,
  startedAt: 0,
  selectedTerm: null,
  matchedTermTokens: [],
  matchedDefTokens: [],
  wrongPair: null,
  hintPair: null,
  correctCount: 0,
  wrongCount: 0,
  finished: false,
  result: null,

  startGame: (s) =>
    set({
      session: s,
      startedAt: Date.now(),
      selectedTerm: null,
      matchedTermTokens: [],
      matchedDefTokens: [],
      wrongPair: null,
      hintPair: null,
      correctCount: 0,
      wrongCount: 0,
      finished: false,
      result: null,
    }),

  selectTerm: (token) => set((st) => ({ selectedTerm: st.selectedTerm === token ? null : token })),

  applyMatch: (termToken, defToken, res) =>
    set((st) => ({
      matchedTermTokens: res.correct ? [...st.matchedTermTokens, termToken] : st.matchedTermTokens,
      matchedDefTokens: res.correct ? [...st.matchedDefTokens, defToken] : st.matchedDefTokens,
      selectedTerm: null,
      correctCount: res.correct_count,
      wrongCount: res.wrong_count,
      finished: res.finished,
    })),

  setWrong: (termToken, defToken) =>
    set({ wrongPair: { term: termToken, def: defToken }, selectedTerm: null }),
  clearWrong: () => set({ wrongPair: null }),
  setHint: (pair) => set({ hintPair: pair }),
  setResult: (r) => set({ result: r }),
  reset: () =>
    set({
      session: null,
      selectedTerm: null,
      matchedTermTokens: [],
      matchedDefTokens: [],
      wrongPair: null,
      hintPair: null,
      correctCount: 0,
      wrongCount: 0,
      finished: false,
      result: null,
    }),
}));
