import { useState, useEffect } from 'react';
import { useRouter } from 'next/router';
import Head from 'next/head';
import Link from 'next/link';
import Layout from '../../components/Layout';
import { useAuthStore } from '../../utils/store';
import { getToken } from '../../utils/auth';

export default function VacanciesPage() {
    const router = useRouter();
    const { isAuthenticated } = useAuthStore();
    const [vacancies, setVacancies] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');

    useEffect(() => {
        if (!isAuthenticated) {
            router.push('/login');
            return;
        }

        const fetchVacancies = async () => {
            setLoading(true);
            setError('');
            const token = getToken();

            try {
                // ✅ ПРАВИЛЬНЫЙ URL и АВТОРИЗАЦИЯ
                const response = await fetch('/api/vacancies', {
                    headers: {
                        'Authorization': `Bearer ${token}`,
                    },
                });

                if (response.ok) {
                    const data = await response.json();
                    setVacancies(data.vacancies || []);
                } else {
                    const errorData = await response.json();
                    setError(errorData.error || 'Не удалось загрузить вакансии.');
                }
            } catch (error) {
                console.error('Error fetching vacancies:', error);
                setError('Ошибка сети при загрузке вакансий.');
            } finally {
                setLoading(false);
            }
        };

        fetchVacancies();
    }, [isAuthenticated, router]);

    return (
        <Layout>
            <Head>
                <title>Список вакансий - HR Avatar</title>
            </Head>
            <div className="max-w-7xl mx-auto py-10 px-4">
                <div className="flex justify-between items-center mb-8">
                    <h1 className="text-3xl font-bold text-gray-900 dark:text-white">Вакансии</h1>
                    <Link href="/vacancies/create" className="bg-indigo-600 hover:bg-indigo-700 text-white font-bold py-2 px-4 rounded-md">
                        + Создать вакансию
                    </Link>
                </div>

                {loading && <p>Загрузка вакансий...</p>}
                {error && <div className="bg-red-100 text-red-700 p-4 rounded-md">{error}</div>}

                {!loading && !error && (
                    <div className="bg-white dark:bg-gray-800 shadow-lg rounded-lg overflow-hidden">
                        <table className="min-w-full">
                            <thead className="bg-gray-50 dark:bg-gray-700">
                            <tr>
                                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">Название</th>
                                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">Статус</th>
                                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">Дата создания</th>
                                <th className="relative px-6 py-3"><span className="sr-only">Действия</span></th>
                            </tr>
                            </thead>
                            <tbody className="divide-y divide-gray-200 dark:divide-gray-700">
                            {vacancies.length > 0 ? vacancies.map((vacancy) => (
                                <tr key={vacancy.id} className="hover:bg-gray-50 dark:hover:bg-gray-600">
                                    <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900 dark:text-white">{vacancy.title}</td>
                                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-300">
                                        <span className="px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-green-100 text-green-800">Активна</span>
                                    </td>
                                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-300">{new Date(vacancy.created_at).toLocaleDateString()}</td>
                                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                                        <a href="#" className="text-indigo-600 hover:text-indigo-900">Подробнее</a>
                                    </td>
                                </tr>
                            )) : (
                                <tr>
                                    <td colSpan="4" className="text-center py-10 text-gray-500">
                                        Нет активных вакансий. <Link href="/vacancies/create" className="text-indigo-600">Создайте первую!</Link>
                                    </td>
                                </tr>
                            )}
                            </tbody>
                        </table>
                    </div>
                )}
            </div>
        </Layout>
    );
}
