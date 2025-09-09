import { useState, useEffect } from 'react';
import { useRouter } from 'next/router';
import Head from 'next/head';
import Link from 'next/link';
import Layout from '../../components/Layout';
import ResumeUploadForm from '../../components/ResumeUploadForm';
import { useAuthStore } from '../../utils/store';
import { getToken } from '../../utils/auth';
import { toast } from 'react-hot-toast';

// Компонент для отображения резюме
function ResumeCard({ resume, index, onDelete }) {
    const [isTextExpanded, setIsTextExpanded] = useState(false);
    const [isReportExpanded, setIsReportExpanded] = useState(false);
    const [isDeleting, setIsDeleting] = useState(false);

    // Функция для извлечения имени кандидата
    const getCandidateName = (resume) => {
        try {
            if (resume.resume_analysis_jsonb) {
                let analysisData;
                if (typeof resume.resume_analysis_jsonb === 'string') {
                    analysisData = JSON.parse(resume.resume_analysis_jsonb);
                } else {
                    analysisData = resume.resume_analysis_jsonb;
                }

                // Ищем email в анализе
                if (analysisData.email && Array.isArray(analysisData.email) && analysisData.email[0]) {
                    const email = analysisData.email[0];
                    const namePart = email.split('@')[0];
                    return namePart.replace(/[._]/g, ' ').replace(/\b\w/g, l => l.toUpperCase());
                }

                if (analysisData.candidate_name) return analysisData.candidate_name;
                if (analysisData.personal_info?.name) return analysisData.personal_info.name;
                if (analysisData.name) return analysisData.name;
            }

            if (resume.text) {
                const lines = resume.text.split('\n').slice(0, 5);
                for (const line of lines) {
                    const trimmed = line.trim();
                    if (trimmed.length > 2 && trimmed.length < 50 && /^[а-яёА-ЯЁa-zA-Z\s]+$/.test(trimmed)) {
                        return trimmed;
                    }
                }
            }

            return `Кандидат ${index + 1}`;
        } catch (error) {
            console.error('Error extracting candidate name:', error);
            return `Кандидат ${index + 1}`;
        }
    };

    // Функция для извлечения оценки и детальной информации
    const getAnalysisData = (resume) => {
        try {
            if (!resume.resume_analysis_jsonb) return { score: 0, skills: null, report: null };

            let analysisData;
            if (typeof resume.resume_analysis_jsonb === 'string') {
                analysisData = JSON.parse(resume.resume_analysis_jsonb);
            } else {
                analysisData = resume.resume_analysis_jsonb;
            }

            // Извлекаем общую оценку
            const finalScore = analysisData?.overall_assessment?.final_score || 0;
            const score = Math.min(finalScore, 100);

            // Извлекаем баллы по категориям
            const detailed = analysisData?.detailed_evaluation || {};
            const skills = {
                soft: detailed.communication_skills?.score || 0,
                hard: detailed.primary_skills?.score || 0,
                case: detailed.work_experience?.score || 0
            };

            // Генерируем текстовый отчет
            const report = generateHumanReport(analysisData);

            return { score, skills, report };
        } catch (error) {
            console.error('Error parsing resume analysis:', error);
            return { score: 0, skills: null, report: null };
        }
    };

    // Функция для генерации человекочитаемого отчета
    const generateHumanReport = (analysisData) => {
        try {
            let report = '';

            // Общая оценка
            if (analysisData.overall_assessment) {
                const assessment = analysisData.overall_assessment;
                report += `📊 ОБЩАЯ ОЦЕНКА\n`;
                report += `Итоговый балл: ${Math.min(assessment.final_score || 0, 100)}/100\n`;
                report += `Уровень соответствия: ${getMatchLevelText(assessment.match_level)}\n`;
                report += `Рекомендация: ${getRecommendationText(assessment.recommendation)}\n`;
                if (assessment.summary_comment) {
                    report += `Комментарий: ${assessment.summary_comment}\n`;
                }
                report += '\n';
            }

            // Сильные стороны
            if (analysisData.strengths && analysisData.strengths.length > 0) {
                report += `✅ СИЛЬНЫЕ СТОРОНЫ\n`;
                analysisData.strengths.forEach((strength, index) => {
                    report += `${index + 1}. ${strength}\n`;
                });
                report += '\n';
            }

            // Опасения
            if (analysisData.concerns && analysisData.concerns.length > 0) {
                report += `⚠️ ПОТЕНЦИАЛЬНЫЕ ПРОБЛЕМЫ\n`;
                analysisData.concerns.forEach((concern, index) => {
                    report += `${index + 1}. ${concern}\n`;
                });
                report += '\n';
            }

            // Красные флаги
            if (analysisData.red_flags && analysisData.red_flags.length > 0) {
                report += `🚩 КРИТИЧЕСКИЕ ЗАМЕЧАНИЯ\n`;
                analysisData.red_flags.forEach((flag, index) => {
                    report += `${index + 1}. ${flag}\n`;
                });
                report += '\n';
            }

            // Детальная оценка
            if (analysisData.detailed_evaluation) {
                report += `📋 ДЕТАЛЬНАЯ ОЦЕНКА\n`;
                const detailed = analysisData.detailed_evaluation;

                Object.entries(detailed).forEach(([key, data]) => {
                    const categoryName = getCategoryName(key);
                    report += `\n${categoryName}: ${data.score}/100 (${getStatusText(data.status)})\n`;
                    if (data.comment) report += `  💭 ${data.comment}\n`;
                    if (data.evidence) report += `  📝 ${data.evidence}\n`;
                });
                report += '\n';
            }

            // Следующие шаги
            if (analysisData.next_steps && analysisData.next_steps.length > 0) {
                report += `🎯 РЕКОМЕНДУЕМЫЕ ДЕЙСТВИЯ\n`;
                analysisData.next_steps.forEach((step, index) => {
                    report += `${index + 1}. ${step}\n`;
                });
                report += '\n';
            }

            // Анализ зарплатных ожиданий
            if (analysisData.salary_expectation_analysis) {
                const salary = analysisData.salary_expectation_analysis;
                report += `💰 ЗАРПЛАТНЫЕ ОЖИДАНИЯ\n`;
                if (salary.candidate_expectation) {
                    report += `Ожидания кандидата: ${salary.candidate_expectation}\n`;
                }
                if (salary.market_range) {
                    report += `Рыночный диапазон: ${salary.market_range}\n`;
                }
                if (salary.comment) {
                    report += `Комментарий: ${salary.comment}\n`;
                }
            }

            return report.trim();
        } catch (error) {
            console.error('Error generating human report:', error);
            return 'Ошибка при генерации отчета';
        }
    };

    // Вспомогательные функции для текста
    const getMatchLevelText = (level) => {
        const levels = {
            'full_match': 'Полное соответствие',
            'strong_match': 'Сильное соответствие',
            'partial_match': 'Частичное соответствие',
            'weak_match': 'Слабое соответствие',
            'no_match': 'Не соответствует'
        };
        return levels[level] || level;
    };

    const getRecommendationText = (rec) => {
        const recommendations = {
            'strongly_recommend_for_interview': 'Настоятельно рекомендую к собеседованию',
            'recommend_for_interview': 'Рекомендую к собеседованию',
            'consider_for_interview': 'Рассмотреть для собеседования',
            'not_recommend': 'Не рекомендую',
            'reject': 'Отклонить'
        };
        return recommendations[rec] || rec;
    };

    const getStatusText = (status) => {
        const statuses = {
            'full_match': 'отлично',
            'strong_match': 'хорошо',
            'partial_match': 'удовлетворительно',
            'weak_match': 'слабо',
            'no_match': 'не соответствует'
        };
        return statuses[status] || status;
    };

    const getCategoryName = (key) => {
        const names = {
            'education': '🎓 Образование',
            'location_match': '📍 Локация',
            'primary_skills': '🔧 Технические навыки',
            'work_experience': '💼 Опыт работы',
            'communication_skills': '💬 Коммуникационные навыки'
        };
        return names[key] || key;
    };

    // Функция удаления резюме
    const handleDelete = async () => {
        if (!window.confirm('Вы уверены, что хотите удалить это резюме?')) {
            return;
        }

        setIsDeleting(true);
        try {
            const token = getToken();
            const response = await fetch(`/api/resumes/${resume.id}`, {
                method: 'DELETE',
                headers: {
                    'Authorization': `Bearer ${token}`,
                },
            });

            if (response.ok) {
                toast.success('Резюме успешно удалено');
                onDelete?.(resume.id);
            } else {
                const error = await response.json();
                toast.error(error.error || 'Ошибка при удалении резюме');
            }
        } catch (error) {
            console.error('Error deleting resume:', error);
            toast.error('Ошибка соединения с сервером');
        } finally {
            setIsDeleting(false);
        }
    };

    const candidateName = getCandidateName(resume);
    const { score, skills, report } = getAnalysisData(resume);
    const maxTextLength = 200;
    const maxReportLength = 300;

    const truncatedText = resume.text && resume.text.length > maxTextLength
        ? resume.text.substring(0, maxTextLength) + '...'
        : resume.text;

    const truncatedReport = report && report.length > maxReportLength
        ? report.substring(0, maxReportLength) + '...'
        : report;

    // Функция для определения цвета оценки
    const getScoreColor = (score) => {
        if (score >= 80) return 'text-green-600 bg-green-100';
        if (score >= 60) return 'text-yellow-600 bg-yellow-100';
        if (score >= 40) return 'text-orange-600 bg-orange-100';
        return 'text-red-600 bg-red-100';
    };

    return (
        <div className="bg-gray-50 p-6 rounded-lg border hover:shadow-md transition-shadow">
            <div className="flex items-start justify-between mb-4">
                <div className="flex-1">
                    <div className="flex items-center space-x-3 mb-2">
                        <h4 className="font-medium text-gray-900">
                            {candidateName}
                        </h4>

                        {/* Общая оценка */}
                        <div className={`px-3 py-1 rounded-full text-sm font-semibold ${getScoreColor(score)}`}>
                            🎯 {score.toFixed(0)} баллов
                        </div>
                    </div>

                    {/* Детализированные оценки по категориям */}
                    {skills && (skills.soft > 0 || skills.hard > 0 || skills.case > 0) && (
                        <div className="flex flex-wrap gap-2 mb-2">
                            {skills.soft > 0 && (
                                <span className="px-2 py-1 bg-blue-100 text-blue-700 rounded text-xs font-medium">
                  💬 Soft: {skills.soft}
                </span>
                            )}
                            {skills.hard > 0 && (
                                <span className="px-2 py-1 bg-green-100 text-green-700 rounded text-xs font-medium">
                  🔧 Hard: {skills.hard}
                </span>
                            )}
                            {skills.case > 0 && (
                                <span className="px-2 py-1 bg-purple-100 text-purple-700 rounded text-xs font-medium">
                  💼 Опыт: {skills.case}
                </span>
                            )}
                        </div>
                    )}

                    {/* Email из поля mail */}
                    {resume.mail && (
                        <p className="text-sm text-gray-600 mb-1">
                            📧 {resume.mail}
                        </p>
                    )}

                    {/* Дата загрузки */}
                    <p className="text-sm text-gray-500 mb-2">
                        📅 {new Date(resume.created_at || Date.now()).toLocaleDateString('ru-RU')}
                    </p>

                    {/* Статус */}
                    {resume.status && (
                        <span className={`inline-block px-2 py-1 rounded-full text-xs font-medium ${
                            resume.status === 'processed' || resume.status === 'Прошел парсер'
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

                {/* Кнопки действий */}
                <div className="flex items-center space-x-2">
                    {resume.file_url && (
                        <a
                            href={resume.file_url}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="px-3 py-1 bg-blue-100 text-blue-700 rounded-lg hover:bg-blue-200 transition-colors text-sm font-medium"
                        >
                            📄 Скачать
                        </a>
                    )}

                    {/* Кнопка удаления */}
                    <button
                        onClick={handleDelete}
                        disabled={isDeleting}
                        className={`px-3 py-1 rounded-lg text-sm font-medium transition-colors ${
                            isDeleting
                                ? 'bg-gray-100 text-gray-400 cursor-not-allowed'
                                : 'bg-red-100 text-red-700 hover:bg-red-200'
                        }`}
                    >
                        {isDeleting ? '⏳' : '🗑️ Удалить'}
                    </button>
                </div>
            </div>

            {/* ИИ-анализ кандидата */}
            {report && (
                <div className="mt-4 pt-4 border-t border-gray-200">
                    <div className="flex items-center justify-between mb-2">
                        <h5 className="text-sm font-medium text-gray-700">🤖 ИИ-анализ кандидата:</h5>
                        {report.length > maxReportLength && (
                            <button
                                onClick={() => setIsReportExpanded(!isReportExpanded)}
                                className="text-blue-600 hover:text-blue-800 text-sm font-medium"
                            >
                                {isReportExpanded ? '🔼 Свернуть' : '🔽 Раскрыть полностью'}
                            </button>
                        )}
                    </div>

                    <div className="bg-blue-50 p-3 rounded border text-sm text-gray-700 leading-relaxed">
            <pre className="whitespace-pre-wrap font-sans">
              {isReportExpanded ? report : truncatedReport}
            </pre>
                    </div>
                </div>
            )}

            {/* Текст резюме с возможностью развернуть */}
            {resume.text && (
                <div className="mt-4 pt-4 border-t border-gray-200">
                    <div className="flex items-center justify-between mb-2">
                        <h5 className="text-sm font-medium text-gray-700">📄 Текст резюме:</h5>
                        {resume.text.length > maxTextLength && (
                            <button
                                onClick={() => setIsTextExpanded(!isTextExpanded)}
                                className="text-blue-600 hover:text-blue-800 text-sm font-medium"
                            >
                                {isTextExpanded ? '🔼 Свернуть' : '🔽 Раскрыть полностью'}
                            </button>
                        )}
                    </div>

                    <div className="bg-white p-3 rounded border text-sm text-gray-700 leading-relaxed">
            <pre className="whitespace-pre-wrap font-sans">
              {isTextExpanded ? resume.text : truncatedText}
            </pre>
                    </div>
                </div>
            )}

            {/* Дополнительная информация об анализе */}
            {score > 0 && (
                <div className="mt-4 pt-4 border-t border-gray-200">
                    <p className="text-xs text-gray-500">
                        ⚡ Результат автоматического анализа соответствия вакансии
                    </p>
                </div>
            )}
        </div>
    );
}

// ОСНОВНОЙ КОМПОНЕНТ
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

    // ✅ ДОБАВЛЕННАЯ ФУНКЦИЯ обработки удаления
    const handleResumeDeleted = (deletedResumeId) => {
        setResumes(prevResumes => prevResumes.filter(resume => resume.id !== deletedResumeId));
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
                        <div className="space-y-6">
                            {resumes.map((resume, index) => (
                                <ResumeCard
                                    key={resume.id || index}
                                    resume={resume}
                                    index={index}
                                    onDelete={handleResumeDeleted}  // ✅ Передаем обработчик удаления
                                />
                            ))}
                        </div>
                    )}
                </div>
            </div>
        </Layout>
    );
}
