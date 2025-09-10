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

        fetchVacancies();
    }, [isAuthenticated, router]);

    const fetchVacancies = async () => {
        setLoading(true);
        setError('');
        const token = getToken();

        try {
            const response = await fetch('/api/vacancies', {
                headers: {
                    'Authorization': `Bearer ${token}`,
                },
            });

            if (response.ok) {
                const data = await response.json();
                setVacancies(data.vacancies || []);
            } else if (response.status === 401) {
                router.push('/login');
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

    // Функция для ограничения текста
    const truncateText = (text, maxLength = 100) => {
        if (!text) return 'Описание не указано';
        return text.length > maxLength ? text.substring(0, maxLength) + '...' : text;
    };

    return (
        <Layout title="Список вакансий">
            <Head>
                <title>Список вакансий - HR Avatar</title>
            </Head>

            <div className="max-w-6xl mx-auto">
                <div className="flex items-center justify-between mb-8">
                    <h1 className="text-3xl font-bold text-gray-900">📋 Вакансии</h1>
                    <Link
                        href="/vacancies/create"
                        className="inline-flex items-center px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
                    >
                        ➕ Создать вакансию
                    </Link>
                </div>

                {loading && (
                    <div className="text-center py-12">
                        <div className="w-8 h-8 border-4 border-blue-600 border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
                        <p className="text-gray-600">Загрузка вакансий...</p>
                    </div>
                )}

                {error && (
                    <div className="bg-red-50 border border-red-200 rounded-lg p-4 mb-6">
                        <p className="text-red-700">{error}</p>
                    </div>
                )}

                {!loading && !error && (
                    <div className="bg-white rounded-xl shadow-lg border overflow-hidden">
                        {vacancies.length > 0 ? (
                            <div className="overflow-x-auto">
                                <table className="w-full">
                                    <thead className="bg-gray-50">
                                    <tr>
                                        <th className="px-6 py-4 text-left text-sm font-medium text-gray-900">Название</th>
                                        <th className="px-6 py-4 text-left text-sm font-medium text-gray-900">Описание</th>
                                        <th className="px-6 py-4 text-left text-sm font-medium text-gray-900">Дата создания</th>
                                        <th className="px-6 py-4 text-right text-sm font-medium text-gray-900">Действия</th>
                                    </tr>
                                    </thead>
                                    <tbody className="divide-y divide-gray-200">
                                    {vacancies.map((vacancy) => (
                                        <tr key={vacancy.id} className="hover:bg-gray-50 transition-colors">
                                            <td className="px-6 py-4">
                                                <p className="font-medium text-gray-900">{vacancy.title}</p>
                                            </td>
                                            <td className="px-6 py-4">
                                                <p className="text-sm text-gray-600 leading-relaxed">
                                                    {truncateText(vacancy.description, 120)}
                                                </p>
                                            </td>
                                            <td className="px-6 py-4 text-gray-600">
                                                {new Date(vacancy.created_at).toLocaleDateString('ru-RU')}
                                            </td>
                                            <td className="px-6 py-4 text-right">
                                                <Link
                                                    href={`/vacancies/${vacancy.id}`}
                                                    className="inline-flex items-center px-4 py-2 bg-blue-100 text-blue-700 rounded-lg hover:bg-blue-200 transition-colors text-sm font-medium"
                                                >
                                                    Подробнее
                                                </Link>
                                            </td>
                                        </tr>
                                    ))}
                                    </tbody>
                                </table>
                            </div>
                        ) : (
                            <div className="text-center py-12">
                                <div className="text-6xl mb-4">📭</div>
                                <h3 className="text-xl font-medium text-gray-900 mb-2">Нет активных вакансий</h3>
                                <p className="text-gray-600 mb-6">Создайте первую вакансию, чтобы начать работу</p>
                                <Link
                                    href="/vacancies/create"
                                    className="inline-flex items-center px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
                                >
                                    ➕ Создать вакансию
                                </Link>
                            </div>
                        )}
                    </div>
                )}
            </div>
        </Layout>
    );
}
