import { describe, expect, it } from "vitest";
import {
  deckFilename,
  parseDeckJson,
  sampleDeckJsonText,
  DECK_JSON_FORMAT,
} from "../deckJson";

describe("parseDeckJson", () => {
  it("parses a full document and normalizes fields", () => {
    const parsed = parseDeckJson(
      JSON.stringify({
        format: DECK_JSON_FORMAT,
        version: 1,
        name: "  水果  ",
        description: " 说明 ",
        entries: [
          { term: " apple ", definition: " 苹果 ", tags: [" 食物 ", "", "水果"] },
          { term: "banana", definition: "香蕉" },
        ],
      }),
    );
    expect(parsed.name).toBe("水果");
    expect(parsed.description).toBe("说明");
    expect(parsed.entries).toEqual([
      { term: "apple", definition: "苹果", tags: ["食物", "水果"] },
      { term: "banana", definition: "香蕉", tags: [] },
    ]);
  });

  it("accepts a bare array of entries", () => {
    const parsed = parseDeckJson(`[{"term":"cat","definition":"猫"}]`);
    expect(parsed.name).toBe("");
    expect(parsed.entries).toHaveLength(1);
  });

  it("accepts the minimal document", () => {
    expect(parseDeckJson(`{"entries":[{"term":"a","definition":"b"}]}`).entries).toHaveLength(1);
  });

  it("rejects invalid input with readable errors", () => {
    const invalid = [
      "",
      "{bad json",
      `{"entries":[]}`,
      `{"entries":[{"term":"a"}]}`,
      `{"format":"other","entries":[{"term":"a","definition":"b"}]}`,
      `{"version":99,"entries":[{"term":"a","definition":"b"}]}`,
      `{"entries":[{"term":"${"x".repeat(201)}","definition":"b"}]}`,
    ];
    for (const text of invalid) {
      expect(() => parseDeckJson(text)).toThrow();
    }
  });

  it("rejects too many entries", () => {
    const entries = Array.from({ length: 5001 }, () => ({ term: "a", definition: "b" }));
    expect(() => parseDeckJson(JSON.stringify({ entries }))).toThrow(/上限/);
  });
});

describe("sampleDeckJsonText", () => {
  it("produces a template that parses back", () => {
    const parsed = parseDeckJson(sampleDeckJsonText());
    expect(parsed.name).toBe("我的词表");
    expect(parsed.entries.length).toBeGreaterThan(0);
  });
});

describe("deckFilename", () => {
  it("strips unsafe characters and appends .json", () => {
    expect(deckFilename('英语/四级:词汇?')).toBe("英语_四级_词汇_.json");
    expect(deckFilename("  ")).toBe("wordtile-deck.json");
  });
});
