// utils/store.js
import { create } from "zustand";
import { getToken, removeToken } from "./auth";

// Состояние для аутентификации
export const useAuthStore = create((set, get) => ({
  user: null,
  isAuthenticated: false,
  loading: true,

  // Инициализация при загрузке приложения
  initialize: () => {
    const token = getToken();
    if (token) {
      set({ isAuthenticated: true, loading: false });
      // Здесь можно добавить декодирование JWT для получения данных пользователя
      // const userData = decodeJWT(token);
      // set({ user: userData });
    } else {
      set({ isAuthenticated: false, loading: false });
    }
  },

  // Логин
  login: (token, userData = null) => {
    set({
      isAuthenticated: true,
      user: userData,
      loading: false
    });
  },

  // Логаут
  logout: () => {
    removeToken();
    set({
      isAuthenticated: false,
      user: null,
      loading: false
    });
  },

  // Установка данных пользователя
  setUser: (userData) => {
    set({ user: userData });
  },
}));

// Состояние для интервью (оставляем как было)
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
