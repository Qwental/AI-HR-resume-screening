import { useState, useEffect } from 'react';
import { useRouter } from 'next/router';
import Head from 'next/head';
import Layout from '../../components/Layout';
import { useAuthStore } from '../../utils/store';
import { getToken } from '../../utils/auth';

export default function CreateVacancyPage() {
    const router = useRouter();
    const { isAuthenticated, user } = useAuthStore(state => ({
        isAuthenticated: state.isAuthenticated,
        user: state.user
    }));

    // Состояния для всех полей формы
    const [title, setTitle] = useState('');
    const [description, setDescription] = useState('');
    const [file, setFile] = useState(null);
    const [weightSoft, setWeightSoft] = useState(30);
    const [weightHard, setWeightHard] = useState(50);
    const [weightCase, setWeightCase] = useState(20);

    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');

    useEffect(() => {
        if (!isAuthenticated || user?.role !== 'hr_specialist') {
            router.push('/login');
        }
    }, [isAuthenticated, user, router]);

    const handleFileChange = (e) => {
        if (e.target.files.length > 0) {
            setFile(e.target.files[0]);
        }
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        if (!title || !description || !file) {
            setError('Название, описание и файл обязательны.');
            return;
        }

        setLoading(true);
        setError('');

        const formData = new FormData();
        formData.append('title', title);
        formData.append('description', description);
        formData.append('file', file);
        formData.append('weight_soft', weightSoft);
        formData.append('weight_hard', weightHard);
        formData.append('weight_case', weightCase);

        try {
            const token = getToken();
            if (!token) {
                setError("Ошибка авторизации. Войдите снова.");
                router.push('/login');
                return;
            }

            // ✅ ПРАВИЛЬНЫЙ URL: /api/vacancies
            const response = await fetch('/api/vacancies', {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${token}`
                },
                body: formData,
            });

            if (response.ok) {
                router.push('/vacancies');
            } else {
                const errorData = await response.json().catch(() => ({ message: 'Произошла неизвестная ошибка' }));
                setError(errorData.error || errorData.message);
            }
        } catch (err) {
            console.error('Create vacancy error:', err);
            setError('Ошибка сети. Проверьте соединение и доступность сервера.');
        } finally {
            setLoading(false);
        }
    };

    if (!isAuthenticated) return <Layout><p>Загрузка...</p></Layout>;

    return (
        <Layout>
            <Head>
                <title>Создать вакансию - HR Avatar</title>
            </Head>
            <div className="max-w-3xl mx-auto py-10 px-4">
                <h1 className="text-3xl font-bold text-gray-900 dark:text-white mb-2">Новая вакансия</h1>
                <p className="text-gray-600 dark:text-gray-400 mb-8">Заполните детали вакансии и прикрепите файл с кейсом для кандидатов.</p>

                <form onSubmit={handleSubmit} className="bg-white dark:bg-gray-800 p-8 rounded-lg shadow-lg space-y-6">
                    {error && <div className="bg-red-100 border-l-4 border-red-500 text-red-700 p-4 rounded-md" role="alert"><p>{error}</p></div>}

                    <div>
                        <label htmlFor="title" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Название вакансии</label>
                        <input type="text" id="title" value={title} onChange={e => setTitle(e.target.value)} required className="w-full bg-gray-50 dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm px-3 py-2 focus:ring-indigo-500 focus:border-indigo-500" placeholder="Go-разработчик (Middle)"/>
                    </div>

                    <div>
                        <label htmlFor="description" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Описание</label>
                        <textarea id="description" rows="4" value={description} onChange={e => setDescription(e.target.value)} required className="w-full bg-gray-50 dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm px-3 py-2 focus:ring-indigo-500 focus:border-indigo-500" placeholder="Обязанности, стек, условия..."></textarea>
                    </div>

                    <div>
                        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">Файл с заданием</label>
                        <div className="mt-1 flex justify-center px-6 pt-5 pb-6 border-2 border-gray-300 dark:border-gray-600 border-dashed rounded-md">
                            <div className="space-y-1 text-center">
                                <svg className="mx-auto h-12 w-12 text-gray-400" stroke="currentColor" fill="none" viewBox="0 0 48 48" aria-hidden="true"><path d="M28 8H12a4 4 0 00-4 4v20m32-12v8m0 0v8a4 4 0 01-4 4H12a4 4 0 01-4-4v-4m32-4l-3.172-3.172a4 4 0 00-5.656 0L28 28M8 32l9.172-9.172a4 4 0 015.656 0L28 28m0 0l4 4m4-24h8m-4-4v8" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" /></svg>
                                <div className="flex text-sm text-gray-600 dark:text-gray-400">
                                    <label htmlFor="file-upload" className="relative cursor-pointer bg-white dark:bg-gray-800 rounded-md font-medium text-indigo-600 hover:text-indigo-500 focus-within:outline-none">
                                        <span>Загрузите файл</span>
                                        <input id="file-upload" name="file-upload" type="file" className="sr-only" onChange={handleFileChange} accept=".docx,.pdf,.md" required />
                                    </label>
                                    <p className="pl-1">или перетащите</p>
                                </div>
                                <p className="text-xs text-gray-500 dark:text-gray-500">{file ? file.name : 'PDF, DOCX, MD до 10MB'}</p>
                            </div>
                        </div>
                    </div>

                    <fieldset>
                        <legend className="text-base font-medium text-gray-900 dark:text-white">Веса для оценки (%)</legend>
                        <div className="mt-4 grid grid-cols-1 gap-y-6 sm:grid-cols-3 sm:gap-x-4">
                            <div>
                                <label htmlFor="weight_soft" className="block text-sm font-medium text-gray-700 dark:text-gray-300">Soft Skills</label>
                                <input type="number" id="weight_soft" value={weightSoft} onChange={e => setWeightSoft(e.target.value)} className="mt-1 w-full bg-gray-50 dark:bg-gray-700 border-gray-300 dark:border-gray-600 rounded-md shadow-sm px-3 py-2" />
                            </div>
                            <div>
                                <label htmlFor="weight_hard" className="block text-sm font-medium text-gray-700 dark:text-gray-300">Hard Skills</label>
                                <input type="number" id="weight_hard" value={weightHard} onChange={e => setWeightHard(e.target.value)} className="mt-1 w-full bg-gray-50 dark:bg-gray-700 border-gray-300 dark:border-gray-600 rounded-md shadow-sm px-3 py-2" />
                            </div>
                            <div>
                                <label htmlFor="weight_case" className="block text-sm font-medium text-gray-700 dark:text-gray-300">Кейс</label>
                                <input type="number" id="weight_case" value={weightCase} onChange={e => setWeightCase(e.target.value)} className="mt-1 w-full bg-gray-50 dark:bg-gray-700 border-gray-300 dark:border-gray-600 rounded-md shadow-sm px-3 py-2" />
                            </div>
                        </div>
                    </fieldset>

                    <div className="flex justify-end pt-4">
                        <button type="button" onClick={() => router.back()} className="bg-gray-200 dark:bg-gray-700 text-gray-900 dark:text-white py-2 px-4 rounded-md mr-3">Отмена</button>
                        <button type="submit" disabled={loading} className="bg-indigo-600 hover:bg-indigo-700 text-white font-bold py-2 px-4 rounded-md disabled:opacity-50">
                            {loading ? 'Создание...' : 'Создать вакансию'}
                        </button>
                    </div>
                </form>
            </div>
        </Layout>
    );
}
