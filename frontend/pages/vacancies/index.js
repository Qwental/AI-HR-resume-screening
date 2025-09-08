import { useState, useEffect } from 'react';
import { useRouter } from 'next/router';
import Head from 'next/head';
import Link from 'next/link';
import Layout from '../../components/Layout';
import ResumeUploadForm from '../../components/ResumeUploadForm';
import { isAuthenticated, apiRequest } from '../../utils/auth';

export default function VacanciesPage() {
    const router = useRouter();
    const [vacancies, setVacancies] = useState([]);
    const [loading, setLoading] = useState(true);
    const [selectedVacancy, setSelectedVacancy] = useState(null);

    useEffect(() => {
        if (!isAuthenticated()) {
            router.push('/login');
            return;
        }
        fetchVacancies();
    }, []);

    const fetchVacancies = async () => {
        try {
            // ✅ ИСПРАВЛЕНО: Используем прямой путь к interview сервису
            const response = await fetch('http://localhost:8081/api/vacancies');
            if (response.ok) {
                const data = await response.json();
                // ✅ ИСПРАВЛЕНО: Правильное извлечение данных
                setVacancies(data.vacancies || []);
            } else {
                console.error('Failed to fetch vacancies:', response.status);
            }
        } catch (error) {
            console.error('Error fetching vacancies:', error);
        } finally {
            setLoading(false);
        }
    };


    const handleResumeUploaded = (resume) => {
        // Обновляем список резюме для вакансии
        setVacancies(prev =>
            prev.map(vacancy =>
                vacancy.id === selectedVacancy?.id
                    ? { ...vacancy, resumes: [...(vacancy.resumes || []), resume] }
                    : vacancy
            )
        );
        setSelectedVacancy(null);
    };

    if (loading) {
        return (
            <div className="flex justify-center items-center min-h-screen">
                <div className="text-lg">Загрузка...</div>
            </div>
        );
    }

    return (
        <>
            <Head>
                <title>Вакансии - HR Avatar</title>
            </Head>

            <Layout title="Управление вакансиями">
                <div className="container mx-auto px-4 py-8">
                    <div className="flex justify-between items-center mb-8">
                        <h1 className="text-3xl font-bold text-gray-900 dark:text-white">
                            Вакансии
                        </h1>
                        <Link href="/vacancies/create">
                            <a className="px-4 py-2 text-sm font-medium text-white bg-blue-600 border border-transparent rounded-md hover:bg-blue-700">
                                Создать вакансию
                            </a>
                        </Link>
                    </div>

                    <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
                        {vacancies.map((vacancy) => (
                            <div key={vacancy.id} className="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6">
                                <div className="flex justify-between items-start mb-4">
                                    <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
                                        {vacancy.title}
                                    </h3>
                                    <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800">
                    Активна
                  </span>
                                </div>

                                <p className="text-gray-600 dark:text-gray-400 mb-4 line-clamp-3">
                                    {vacancy.description}
                                </p>

                                <div className="flex justify-between items-center text-sm text-gray-500 dark:text-gray-400 mb-4">
                                    <span>{vacancy.resumes?.length || 0} резюме</span>
                                    <span>Создана: {new Date(vacancy.created_at).toLocaleDateString()}</span>
                                </div>

                                <div className="grid grid-cols-3 gap-2 text-xs text-gray-600 dark:text-gray-400 mb-4">
                                    <div className="text-center">
                                        <div className="font-medium">Soft</div>
                                        <div>{vacancy.weight_soft}%</div>
                                    </div>
                                    <div className="text-center">
                                        <div className="font-medium">Hard</div>
                                        <div>{vacancy.weight_hard}%</div>
                                    </div>
                                    <div className="text-center">
                                        <div className="font-medium">Опыт</div>
                                        <div>{vacancy.weight_case}%</div>
                                    </div>
                                </div>

                                <div className="flex space-x-2">
                                    <button
                                        onClick={() => setSelectedVacancy(vacancy)}
                                        className="flex-1 px-3 py-2 text-sm font-medium text-blue-600 bg-blue-50 hover:bg-blue-100 rounded-md"
                                    >
                                        Добавить резюме
                                    </button>
                                    <Link href={`/vacancies/${vacancy.id}`}>
                                        <a className="flex-1 px-3 py-2 text-sm font-medium text-gray-700 bg-gray-100 hover:bg-gray-200 rounded-md text-center">
                                            Просмотр
                                        </a>
                                    </Link>
                                </div>
                            </div>
                        ))}
                    </div>

                    {vacancies.length === 0 && (
                        <div className="text-center py-12">
                            <svg className="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                            </svg>
                            <h3 className="mt-2 text-sm font-medium text-gray-900 dark:text-white">
                                Нет вакансий
                            </h3>
                            <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
                                Создайте первую вакансию для начала работы
                            </p>
                            <div className="mt-6">
                                <Link href="/vacancies/create">
                                    <a className="inline-flex items-center px-4 py-2 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700">
                                        Создать вакансию
                                    </a>
                                </Link>
                            </div>
                        </div>
                    )}
                </div>

                {/* Модальное окно для загрузки резюме */}
                {selectedVacancy && (
                    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
                        <div className="bg-white dark:bg-gray-800 rounded-lg max-w-lg w-full max-h-[90vh] overflow-y-auto">
                            <div className="p-4 border-b border-gray-200 dark:border-gray-700">
                                <div className="flex justify-between items-center">
                                    <h3 className="text-lg font-medium">
                                        Добавить резюме к вакансии
                                    </h3>
                                    <button
                                        onClick={() => setSelectedVacancy(null)}
                                        className="text-gray-400 hover:text-gray-600"
                                    >
                                        <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                                        </svg>
                                    </button>
                                </div>
                                <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
                                    {selectedVacancy.title}
                                </p>
                            </div>
                            <div className="p-4">
                                <ResumeUploadForm
                                    vacancyId={selectedVacancy.id}
                                    onSuccess={handleResumeUploaded}
                                />
                            </div>
                        </div>
                    </div>
                )}
            </Layout>
        </>
    );
}
