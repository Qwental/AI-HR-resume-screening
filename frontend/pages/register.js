import { useState } from 'react';
import { useRouter } from 'next/router';
import Head from 'next/head';
import Link from 'next/link';
import { setToken } from '../utils/auth';
import { useAuthStore } from '../utils/store';

export default function RegisterPage() {
    const [username, setUsername] = useState('');
    const [surname, setSurname] = useState('');
    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');
    const [confirmPassword, setConfirmPassword] = useState('');
    const [error, setError] = useState('');
    const [loading, setLoading] = useState(false);
    const router = useRouter();

    // ✅ ДОБАВЛЯЕМ: получаем функцию login из store
    const login = useAuthStore(state => state.login);

    const handleSubmit = async (e) => {
        e.preventDefault();
        if (!username || !surname || !email || !password) {
            setError('Пожалуйста, заполните все поля.');
            return;
        }
        if (password.length < 6) {
            setError('Пароль должен содержать минимум 6 символов.');
            return;
        }
        if (password !== confirmPassword) {
            setError('Пароли не совпадают.');
            return;
        }
        setError('');
        setLoading(true);
        try {
            const response = await fetch('/api/auth/register', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ username, surname, email, password }),
            });
            const data = await response.json();
            if (response.ok) {
                // ✅ ИСПРАВЛЕНО: Сохраняем токен и обновляем состояние
                setToken(data.data.access_token);
                login(data.data.access_token, data.data.user);

                // ✅ ИСПРАВЛЕНО: Сразу перенаправляем на дашборд
                router.push('/dashboard');
            } else {
                setError(data.message || 'Ошибка регистрации');
            }
        } catch (err) {
            console.error('Register error:', err);
            setError('Ошибка соединения с сервером');
        } finally {
            setLoading(false);
        }
    };

    return (
        <>
            <Head>
                <title>Регистрация — HR Avatar</title>
                <meta name="description" content="Создать аккаунт в HR Avatar" />
            </Head>
            <div className="min-h-screen bg-gray-50 dark:bg-gray-900 flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8">
                <div className="max-w-md w-full space-y-8">
                    <div className="sm:mx-auto sm:w-full sm:max-w-md">
                        <Link href="/">
                            <h2 className="mt-6 text-center text-3xl font-extrabold text-gray-900 dark:text-white cursor-pointer hover:text-blue-600 transition-colors">
                                HR Avatar
                            </h2>
                        </Link>
                        <p className="mt-2 text-center text-sm text-gray-600 dark:text-gray-400">
                            Система управления персоналом
                        </p>
                    </div>
                    <form className="mt-8 space-y-6 bg-white dark:bg-gray-800 p-8 rounded-lg shadow" onSubmit={handleSubmit}>
                        <div className="space-y-4">
                            <div>
                                <label htmlFor="username" className="block text-sm font-medium text-gray-700 dark:text-gray-300">
                                    Имя пользователя
                                </label>
                                <input
                                    id="username"
                                    name="username"
                                    type="text"
                                    required
                                    value={username}
                                    onChange={(e) => setUsername(e.target.value)}
                                    placeholder="Иван"
                                    className="mt-1 block w-full px-3 py-2 border rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-white"
                                />
                            </div>
                            <div>
                                <label htmlFor="surname" className="block text-sm font-medium text-gray-700 dark:text-gray-300">
                                    Фамилия
                                </label>
                                <input
                                    id="surname"
                                    name="surname"
                                    type="text"
                                    required
                                    value={surname}
                                    onChange={(e) => setSurname(e.target.value)}
                                    placeholder="Иванов"
                                    className="mt-1 block w-full px-3 py-2 border rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-white"
                                />
                            </div>
                            <div>
                                <label htmlFor="email" className="block text-sm font-medium text-gray-700 dark:text-gray-300">
                                    Email
                                </label>
                                <input
                                    id="email"
                                    name="email"
                                    type="email"
                                    autoComplete="email"
                                    required
                                    value={email}
                                    onChange={(e) => setEmail(e.target.value)}
                                    placeholder="you@example.com"
                                    className="mt-1 block w-full px-3 py-2 border rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-white"
                                />
                            </div>
                            <div>
                                <label htmlFor="password" className="block text-sm font-medium text-gray-700 dark:text-gray-300">
                                    Пароль
                                </label>
                                <input
                                    id="password"
                                    name="password"
                                    type="password"
                                    required
                                    value={password}
                                    onChange={(e) => setPassword(e.target.value)}
                                    placeholder="••••••••"
                                    className="mt-1 block w-full px-3 py-2 border rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-white"
                                />
                                <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">
                                    Минимум 6 символов
                                </p>
                            </div>
                            <div>
                                <label htmlFor="confirmPassword" className="block text-sm font-medium text-gray-700 dark:text-gray-300">
                                    Подтвердите пароль
                                </label>
                                <input
                                    id="confirmPassword"
                                    name="confirmPassword"
                                    type="password"
                                    required
                                    value={confirmPassword}
                                    onChange={(e) => setConfirmPassword(e.target.value)}
                                    placeholder="••••••••"
                                    className="mt-1 block w-full px-3 py-2 border rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-white"
                                />
                            </div>
                        </div>
                        {error && (
                            <div className="rounded-md bg-red-50 p-4 text-red-700">
                                {error}
                            </div>
                        )}
                        <div>
                            <button
                                type="submit"
                                disabled={loading}
                                className="group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50"
                            >
                                {loading ? 'Регистрация...' : 'Зарегистрироваться'}
                            </button>
                        </div>
                        <div className="text-sm text-center text-gray-600 dark:text-gray-400">
                            Уже есть аккаунт?{' '}
                            <Link href="/login" className="font-medium text-blue-600 hover:text-blue-500">
                                Войти
                            </Link>
                        </div>
                    </form>
                </div>
            </div>
        </>
    );
}
