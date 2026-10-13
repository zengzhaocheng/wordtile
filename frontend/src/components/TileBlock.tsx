import { motion } from "framer-motion";

export type TileState = "normal" | "selected" | "matched" | "wrong" | "hint";

const stateClass: Record<TileState, string> = {
  normal: "bg-surface border-line",
  selected: "bg-primary border-primary text-primary-ink",
  matched: "bg-surface border-success opacity-0",
  wrong: "bg-danger border-danger text-white",
  hint: "bg-success border-success text-white",
};

interface Props {
  text: string;
  state: TileState;
  disabled?: boolean;
  onClick?: () => void;
}

/** 词条 / 释义块（上屏与下屏通用） */
export default function TileBlock({ text, state, disabled, onClick }: Props) {
  return (
    <motion.button
      type="button"
      onClick={onClick}
      disabled={disabled || state === "matched"}
      initial={{ opacity: 0, y: 8 }}
      animate={
        state === "matched"
          ? { opacity: 0, scale: 0.6 }
          : state === "wrong"
            ? { opacity: 1, y: 0, scale: 1, x: [0, -8, 8, -6, 6, 0] }
            : { opacity: 1, y: 0, scale: 1, x: 0 }
      }
      transition={{ duration: state === "wrong" ? 0.5 : 0.25 }}
      className={`min-h-14 rounded-tile border-2 px-3 py-2 text-sm font-medium break-words transition-colors ${stateClass[state]} ${
        disabled ? "cursor-default" : "active:scale-95"
      }`}
    >
      {text}
    </motion.button>
  );
}
