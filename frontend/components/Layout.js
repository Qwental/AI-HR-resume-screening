// components/Layout.js
import Link from "next/link";
import { useState, useEffect } from "react";
import { useRouter } from 'next/router';
import { useAuthStore } from '../utils/store';

export default function Layout({ title, children }) {
  const [dark, setDark] = useState(false);
  const router = useRouter();
  const { user, logout, isAuthenticated } = useAuthStore();

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

  const handleLogout = () => {
    logout();
    router.push('/login');
  };

  return (
      <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
        {/* Навигационная панель */}
        <nav className="bg-white dark:bg-gray-800 shadow-sm border-b border-gray-200 dark:border-gray-700">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div className="flex justify-between h-16">
              <div className="flex items-center">
                <Link href="/">
                  <a className="text-xl font-bold text-gray-900 dark:text-white">
                    HR Avatar
                  </a>
                </Link>

                {isAuthenticated && (
                    <div className="ml-10 flex space-x-4">
                      <Link href="/vacancies">
                        <a className="text-gray-700 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white px-3 py-2 text-sm font-medium">
                          Вакансии
                        </a>
                      </Link>
                    </div>
                )}
              </div>

              <div className="flex items-center space-x-4">
                {/* Переключатель темы */}
                <button
                    onClick={() => setDark(!dark)}
                    className="p-2 text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
                >
                  {dark ? "🌞" : "🌙"}
                </button>

                {isAuthenticated ? (
                    <>
                      {user && (
                          <span className="text-sm text-gray-600 dark:text-gray-400">
                                            {user.username || user.email}
                                        </span>
                      )}
                      <button
                          onClick={handleLogout}
                          className="text-gray-700 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white px-3 py-2 text-sm font-medium"
                      >
                        Выход
                      </button>
                    </>
                ) : (
                    <div className="flex space-x-2">
                      <Link href="/login">
                        <a className="text-gray-700 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white px-3 py-2 text-sm font-medium">
                          Вход
                        </a>
                      </Link>
                      <Link href="/register">
                        <a className="bg-blue-600 hover:bg-blue-700 text-white px-3 py-2 text-sm font-medium rounded-md">
                          Регистрация
                        </a>
                      </Link>
                    </div>
                )}
              </div>
            </div>
          </div>
        </nav>

        {/* Основной контент */}
        <main className="flex-1">
          {title && (
              <header className="bg-white dark:bg-gray-800 shadow-sm">
                <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
                  <h1 className="text-3xl font-bold text-gray-900 dark:text-white">
                    {title}
                  </h1>
                </div>
              </header>
          )}
          {children}
        </main>
      </div>
  );
}
