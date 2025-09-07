import { create } from "zustand";

export const useInterviewStore = create((set) => ({
  status: "idle",
  transcript: "",
  questions: [
    "Расскажите коротко о себе.",
    "Опишите самый сложный технический вызов, который вы решали.",
    "Как вы работаете с негативной обратной связью?",
    "Какие цели на ближайший год?"
  ],
  setStatus: (v) => set({ status: v }),
  setTranscript: (t) => set({ transcript: t }),
}));
