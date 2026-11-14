/**
 * WordTile 题库 JSON 交换格式（v1）的前端解析与校验。
 * 与后端 internal/deckio 的规则保持一致，先在浏览器给出友好提示，再由服务端兜底校验。
 */
export interface DeckJsonEntry {
  term: string;
  definition: string;
  tags?: string[];
}

export interface DeckJsonFile {
  format?: string;
  version?: number;
  name?: string;
  description?: string;
  exported_at?: string;
  entries: DeckJsonEntry[];
}

export interface ParsedDeck {
  name: string;
  description: string;
  entries: { term: string; definition: string; tags: string[] }[];
}

export const DECK_JSON_FORMAT = "wordtile-deck";
export const DECK_JSON_VERSION = 1;
export const MAX_ENTRIES = 5000;
const MAX_TERM_LEN = 200;
const MAX_DEF_LEN = 1000;
const MAX_TAGS = 20;
const MAX_TAG_LEN = 50;

/** 用户下载后可照着填写的模板。 */
export function sampleDeckJsonText(): string {
  return JSON.stringify(
    {
      format: DECK_JSON_FORMAT,
      version: DECK_JSON_VERSION,
      name: "我的词表",
      description: "可选：题库说明",
      entries: [
        { term: "apple", definition: "苹果", tags: ["食物", "水果"] },
        { term: "banana", definition: "香蕉", tags: ["食物"] },
        { term: "travel", definition: "旅行", tags: [] },
      ],
    },
    null,
    2,
  );
}

/** 解析并校验题库 JSON，失败时抛出带可读信息的 Error。 */
export function parseDeckJson(text: string): ParsedDeck {
  const trimmed = text.trim();
  if (!trimmed) throw new Error("文件内容为空");

  let raw: unknown;
  try {
    raw = JSON.parse(trimmed);
  } catch {
    throw new Error("JSON 格式错误，请检查括号与引号");
  }

  let file: Partial<DeckJsonFile>;
  if (Array.isArray(raw)) {
    file = { entries: raw as DeckJsonEntry[] };
  } else if (raw && typeof raw === "object") {
    file = raw as Partial<DeckJsonFile>;
  } else {
    throw new Error("文件内容应为对象或词条数组");
  }

  if (file.format && file.format !== DECK_JSON_FORMAT) {
    throw new Error(`不支持的 format：${file.format}`);
  }
  if (typeof file.version === "number" && file.version > DECK_JSON_VERSION) {
    throw new Error(`不支持的 version：${file.version}`);
  }
  if (!Array.isArray(file.entries) || file.entries.length === 0) {
    throw new Error("entries 不能为空，至少需要 1 条词条");
  }
  if (file.entries.length > MAX_ENTRIES) {
    throw new Error(`词条数量 ${file.entries.length} 超过上限 ${MAX_ENTRIES}`);
  }

  const entries = file.entries.map((item, i) => {
    if (!item || typeof item !== "object") {
      throw new Error(`第 ${i + 1} 条词条不是对象`);
    }
    const term = String(item.term ?? "").trim();
    const definition = String(item.definition ?? "").trim();
    if (!term || !definition) {
      throw new Error(`第 ${i + 1} 条词条的 term / definition 不能为空`);
    }
    if ([...term].length > MAX_TERM_LEN) {
      throw new Error(`第 ${i + 1} 条词条的 term 超过 ${MAX_TERM_LEN} 字`);
    }
    if ([...definition].length > MAX_DEF_LEN) {
      throw new Error(`第 ${i + 1} 条词条的 definition 超过 ${MAX_DEF_LEN} 字`);
    }
    const rawTags = Array.isArray(item.tags) ? item.tags : [];
    if (rawTags.length > MAX_TAGS) {
      throw new Error(`第 ${i + 1} 条词条的标签超过 ${MAX_TAGS} 个`);
    }
    const tags = rawTags
      .map((tag) => String(tag).trim())
      .filter(Boolean);
    if (tags.some((tag) => [...tag].length > MAX_TAG_LEN)) {
      throw new Error(`第 ${i + 1} 条词条的标签超过 ${MAX_TAG_LEN} 字`);
    }
    return { term, definition, tags };
  });

  return {
    name: String(file.name ?? "").trim(),
    description: String(file.description ?? "").trim(),
    entries,
  };
}

/** 把解析结果规范化成可直接提交给导入 API 的文档。 */
export function toDeckJson(parsed: ParsedDeck, fallbackName = ""): DeckJsonFile {
  return {
    format: DECK_JSON_FORMAT,
    version: DECK_JSON_VERSION,
    name: parsed.name || fallbackName,
    description: parsed.description,
    entries: parsed.entries,
  };
}

/** 把任意 JSON 数据下载成 .json 文件。 */
export function downloadJson(data: unknown, filename: string): void {
  const blob = new Blob([JSON.stringify(data, null, 2)], {
    type: "application/json;charset=utf-8",
  });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  a.remove();
  setTimeout(() => URL.revokeObjectURL(url), 1000);
}

/** 生成安全的文件名（去掉不适合文件系统与 HTTP 头的字符）。 */
export function deckFilename(name: string): string {
  const safe = name
    .replace(/[\\/:*?"<>|\r\n]/g, "_")
    .trim()
    .slice(0, 60);
  return `${safe || "wordtile-deck"}.json`;
}
