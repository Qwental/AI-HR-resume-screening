import Link from "next/link";
import { useState, useEffect } from "react";

export default function Layout({ title, children }) {
  const [dark, setDark] = useState(false);

  useEffect(() => {
    // Проверяем тему из localStorage
    const savedTheme = localStorage.getItem("theme");
    if (savedTheme === "dark") {
      setDark(true);
      document.documentElement.classList.add("dark");
    }
  }, []);

  useEffect(() => {
    if (dark) {
      document.documentElement.classList.add("dark");
      localStorage.setItem("theme", "dark");
    } else {
      document.documentElement.classList.remove("dark");
      localStorage.setItem("theme", "light");
    }
  }, [dark]);

  return (
    <div className="min-h-screen bg-white dark:bg-black text-gray-900 dark:text-gray-100 transition-colors duration-300">
      <header className="sticky top-0 z-10 bg-white/70 dark:bg-black/70 backdrop-blur border-b border-gray-200 dark:border-gray-800 transition-colors duration-300">
        <div className="max-w-7xl mx-auto px-4 py-3 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-9 h-9 rounded-xl bg-brand"></div>
            <span className="font-semibold">{title || "AI HR Platform"}</span>
          </div>
          <nav className="flex items-center gap-4 text-sm">
            <Link href="/dashboard" className="hover:text-brand">Дашборд</Link>
            <Link href="/vacancies" className="hover:text-brand">Вакансии</Link>
            <Link href="/login" className="hover:text-brand">Вход</Link>

          </nav>
          <button
            className="btn"
            onClick={() => setDark((v) => !v)}
            aria-label="Toggle theme"
          >
            {dark ? "☀️" : "🌙"}
          </button>
        </div>
      </header>
      <main className="max-w-7xl mx-auto px-4 py-6">{children}</main>
    </div>
  );
}
