import { useState, useEffect } from 'react';
import { useRouter } from 'next/router';
import Head from 'next/head';
import Link from 'next/link';
import Layout from '../../components/Layout';
import CandidateCard from '../../components/CandidateCard';
import ResumeUploadForm from '../../components/ResumeUploadForm';
import { useAuthStore } from '../../utils/store';
import { getToken } from '../../utils/auth';
import { toast } from 'react-hot-toast';

export default function VacancyDetail() {
    const router = useRouter();
    const { id } = router.query;
    const { isAuthenticated } = useAuthStore();

    const [vacancy, setVacancy] = useState(null);
    const [resumes, setResumes] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');
    const [resumesLoading, setResumesLoading] = useState(false);

    useEffect(() => {
        if (!isAuthenticated) {
            router.push('/login');
            return;
        }

        if (!id) return;

        fetchVacancyData();
    }, [id, isAuthenticated, router]);

    const fetchVacancyData = async () => {
        setLoading(true);
        try {
            await Promise.all([
                fetchVacancyDetails(),
                fetchVacancyResumes()
            ]);
        } catch (err) {
            console.error('Error fetching data:', err);
            setError('Ошибка при загрузке данных вакансии');
        } finally {
            setLoading(false);
        }
    };

    const fetchVacancyDetails = async () => {
        const token = getToken();
        const response = await fetch(`/api/vacancies/${id}`, {
            headers: { 'Authorization': `Bearer ${token}` }
        });

        if (response.ok) {
            const data = await response.json();
            setVacancy(data.vacancy || data);
        } else if (response.status === 401) {
            router.push('/login');
        } else {
            throw new Error('Failed to fetch vacancy');
        }
    };

    const fetchVacancyResumes = async () => {
        setResumesLoading(true);
        const token = getToken();

        try {
            const response = await fetch(`/api/vacancies/${id}/resumes`, {
                headers: { 'Authorization': `Bearer ${token}` }
            });

            if (response.ok) {
                const data = await response.json();
                setResumes(data.resumes || []);
            } else {
                console.warn('Failed to fetch resumes');
            }
        } catch (error) {
            console.error('Error fetching resumes:', error);
        } finally {
            setResumesLoading(false);
        }
    };

    const handleResumeUploaded = (newResume) => {
        toast.success('Резюме успешно загружено!');
        fetchVacancyResumes(); // Обновляем список резюме
    };

    if (loading) {
        return (
            <Layout title="Загрузка...">
                <div className="flex items-center justify-center min-h-96">
                    <div className="flex flex-col items-center space-y-4">
                        <div className="w-8 h-8 border-4 border-blue-600 border-t-transparent rounded-full animate-spin"></div>
                        <p className="text-gray-600">Загрузка вакансии...</p>
                    </div>
                </div>
            </Layout>
        );
    }

    if (error) {
        return (
            <Layout title="Ошибка">
                <div className="text-center py-12">
                    <div className="text-6xl mb-4">😞</div>
                    <h2 className="text-2xl font-semibold text-gray-900 mb-2">Произошла ошибка</h2>
                    <p className="text-gray-600 mb-6">{error}</p>
                    <Link
                        href="/vacancies"
                        className="inline-flex items-center px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
                    >
                        ← Вернуться к списку вакансий
                    </Link>
                </div>
            </Layout>
        );
    }

    if (!vacancy) {
        return (
            <Layout title="Вакансия не найдена">
                <div className="text-center py-12">
                    <div className="text-6xl mb-4">📭</div>
                    <h2 className="text-2xl font-semibold text-gray-900 mb-2">Вакансия не найдена</h2>
                    <p className="text-gray-600 mb-6">Запрашиваемая вакансия не существует или была удалена</p>
                    <Link
                        href="/vacancies"
                        className="inline-flex items-center px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
                    >
                        ← Вернуться к списку вакансий
                    </Link>
                </div>
            </Layout>
        );
    }

    return (
        <Layout title={`${vacancy.title} - Детали вакансии`}>
            <Head>
                <title>{vacancy.title} - HR Avatar</title>
            </Head>

            <div className="max-w-6xl mx-auto space-y-8">
                {/* Навигация */}
                <div className="flex items-center space-x-2 text-sm">
                    <Link href="/vacancies" className="text-blue-600 hover:text-blue-800">
                        Вакансии
                    </Link>
                    <span className="text-gray-400">→</span>
                    <span className="text-gray-600">{vacancy.title}</span>
                </div>

                {/* Основная информация о вакансии */}
                <div className="bg-white p-8 rounded-xl shadow-lg border">
                    <div className="flex items-start justify-between mb-6">
                        <div className="flex-1">
                            <h1 className="text-3xl font-bold text-gray-900 mb-2">{vacancy.title}</h1>
                            <div className="flex items-center space-x-4 text-sm text-gray-600">
                                {vacancy.created_at && (
                                    <span>📅 Создано: {new Date(vacancy.created_at).toLocaleDateString('ru-RU')}</span>
                                )}
                                <span className={`px-2 py-1 rounded text-xs font-medium ${
                                    vacancy.status === 'active'
                                        ? 'bg-green-100 text-green-800'
                                        : 'bg-gray-100 text-gray-800'
                                }`}>
                  {vacancy.status === 'active' ? '🟢 Активна' : '⚪ Неактивна'}
                </span>
                            </div>
                        </div>

                        <Link
                            href="/vacancies"
                            className="px-4 py-2 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 transition-colors"
                        >
                            ← Назад к списку
                        </Link>
                    </div>

                    {/* Описание */}
                    {vacancy.description && (
                        <div className="mb-6">
                            <h3 className="text-lg font-semibold text-gray-900 mb-3">📋 Описание</h3>
                            <div className="bg-gray-50 p-4 rounded-lg">
                                <p className="text-gray-700 whitespace-pre-wrap">{vacancy.description}</p>
                            </div>
                        </div>
                    )}

                    {/* Веса критериев */}
                    {(vacancy.weight_soft || vacancy.weight_hard || vacancy.weight_case) && (
                        <div className="mb-6">
                            <h3 className="text-lg font-semibold text-gray-900 mb-3">⚖️ Критерии оценки</h3>
                            <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
                                {vacancy.weight_soft && (
                                    <div className="bg-blue-50 p-4 rounded-lg text-center">
                                        <div className="text-2xl font-bold text-blue-600">{vacancy.weight_soft}%</div>
                                        <div className="text-sm text-gray-700">Soft Skills</div>
                                    </div>
                                )}
                                {vacancy.weight_hard && (
                                    <div className="bg-green-50 p-4 rounded-lg text-center">
                                        <div className="text-2xl font-bold text-green-600">{vacancy.weight_hard}%</div>
                                        <div className="text-sm text-gray-700">Hard Skills</div>
                                    </div>
                                )}
                                {vacancy.weight_case && (
                                    <div className="bg-purple-50 p-4 rounded-lg text-center">
                                        <div className="text-2xl font-bold text-purple-600">{vacancy.weight_case}%</div>
                                        <div className="text-sm text-gray-700">Кейсы</div>
                                    </div>
                                )}
                            </div>
                        </div>
                    )}
                </div>

                {/* Секция загрузки резюме */}
                <ResumeUploadForm
                    vacancyId={id}
                    onSuccess={handleResumeUploaded}
                />

                {/* Список резюме */}
                <div className="bg-white p-8 rounded-xl shadow-lg border">
                    <div className="flex items-center justify-between mb-6">
                        <h2 className="text-2xl font-semibold text-gray-900">
                            📄 Резюме ({resumes.length})
                        </h2>
                        {resumesLoading && (
                            <div className="flex items-center space-x-2 text-gray-600">
                                <div className="w-4 h-4 border-2 border-blue-600 border-t-transparent rounded-full animate-spin"></div>
                                <span>Обновление...</span>
                            </div>
                        )}
                    </div>

                    {resumes.length === 0 ? (
                        <div className="text-center py-12">
                            <div className="text-6xl mb-4">📭</div>
                            <h3 className="text-xl font-medium text-gray-900 mb-2">Нет загруженных резюме</h3>
                            <p className="text-gray-600">
                                К данной вакансии пока не прикреплено ни одного резюме.
                                Используйте форму выше для загрузки резюме.
                            </p>
                        </div>
                    ) : (
                        <div className="space-y-4">
                            {resumes.map((resume, index) => (
                                <div key={resume.id || index} className="bg-gray-50 p-6 rounded-lg border">
                                    <div className="flex items-start justify-between">
                                        <div className="flex-1">
                                            <h4 className="font-medium text-gray-900">
                                                {resume.candidate_name || `Кандидат ${index + 1}`}
                                            </h4>
                                            {resume.candidate_email && (
                                                <p className="text-sm text-gray-600 mt-1">
                                                    📧 {resume.candidate_email}
                                                </p>
                                            )}
                                            <p className="text-sm text-gray-500 mt-1">
                                                📅 {new Date(resume.created_at || Date.now()).toLocaleDateString('ru-RU')}
                                            </p>
                                            {resume.status && (
                                                <span className={`inline-block px-2 py-1 mt-2 rounded-full text-xs font-medium ${
                                                    resume.status === 'processed'
                                                        ? 'bg-green-100 text-green-800'
                                                        : resume.status === 'processing'
                                                            ? 'bg-yellow-100 text-yellow-800'
                                                            : resume.status === 'error'
                                                                ? 'bg-red-100 text-red-800'
                                                                : 'bg-gray-100 text-gray-800'
                                                }`}>
                          {resume.status}
                        </span>
                                            )}
                                        </div>

                                        {resume.file_url && (
                                            <a
                                                href={resume.file_url}
                                                target="_blank"
                                                rel="noopener noreferrer"
                                                className="px-3 py-1 bg-blue-100 text-blue-700 rounded-lg hover:bg-blue-200 transition-colors text-sm"
                                            >
                                                📄 Скачать
                                            </a>
                                        )}
                                    </div>

                                    {resume.text && (
                                        <div className="mt-4 pt-4 border-t border-gray-200">
                                            <p className="text-sm text-gray-700 line-clamp-3">
                                                {resume.text.substring(0, 200)}
                                                {resume.text.length > 200 ? '...' : ''}
                                            </p>
                                        </div>
                                    )}
                                </div>
                            ))}
                        </div>
                    )}
                </div>
            </div>
        </Layout>
    );
}
