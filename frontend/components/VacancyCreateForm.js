// components/VacancyCreateForm.js
import { useState } from 'react';
import { apiRequest } from '../utils/auth';

export default function VacancyCreateForm({ onSuccess }) {
    const [formData, setFormData] = useState({
        title: '',
        description: '',
        users_id: '1', // TODO: Получить из контекста пользователя
        weight_soft: 33,
        weight_hard: 33,
        weight_case: 34,
    });
    const [file, setFile] = useState(null);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');

    const handleSubmit = async (e) => {
        e.preventDefault();

        if (!file) {
            setError('Пожалуйста, выберите файл');
            return;
        }

        setLoading(true);
        setError('');

        try {
            const formDataToSend = new FormData();
            Object.keys(formData).forEach(key => {
                formDataToSend.append(key, formData[key]);
            });
            formDataToSend.append('file', file);

            const response = await fetch('http://localhost:8081/api/vacancies', {
                method: 'POST',
                body: formDataToSend,
                // Не устанавливаем Content-Type для FormData
            });

            if (response.ok) {
                const vacancy = await response.json();
                onSuccess?.(vacancy);
                // Сброс формы
                setFormData({
                    title: '',
                    description: '',
                    users_id: '1',
                    weight_soft: 33,
                    weight_hard: 33,
                    weight_case: 34,
                });
                setFile(null);
            } else {
                const errorData = await response.json();
                setError(errorData.error || 'Ошибка создания вакансии');
            }
        } catch (err) {
            setError('Ошибка соединения с сервером');
            console.error('Create vacancy error:', err);
        } finally {
            setLoading(false);
        }
    };

    return (
        <form onSubmit={handleSubmit} className="space-y-6">
            <div>
                <label className="block text-sm font-medium text-gray-700">
                    Название вакансии
                </label>
                <input
                    type="text"
                    value={formData.title}
                    onChange={(e) => setFormData({...formData, title: e.target.value})}
                    className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                    required
                />
            </div>

            <div>
                <label className="block text-sm font-medium text-gray-700">
                    Описание
                </label>
                <textarea
                    value={formData.description}
                    onChange={(e) => setFormData({...formData, description: e.target.value})}
                    rows={4}
                    className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                />
            </div>

            <div>
                <label className="block text-sm font-medium text-gray-700">
                    Файл требований
                </label>
                <input
                    type="file"
                    onChange={(e) => setFile(e.target.files[0])}
                    accept=".pdf,.doc,.docx,.txt"
                    className="mt-1 block w-full"
                    required
                />
            </div>

            <div className="grid grid-cols-3 gap-4">
                <div>
                    <label className="block text-sm font-medium text-gray-700">
                        Soft Skills (%)
                    </label>
                    <input
                        type="number"
                        value={formData.weight_soft}
                        onChange={(e) => setFormData({...formData, weight_soft: parseInt(e.target.value)})}
                        min="0"
                        max="100"
                        className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                    />
                </div>

                <div>
                    <label className="block text-sm font-medium text-gray-700">
                        Hard Skills (%)
                    </label>
                    <input
                        type="number"
                        value={formData.weight_hard}
                        onChange={(e) => setFormData({...formData, weight_hard: parseInt(e.target.value)})}
                        min="0"
                        max="100"
                        className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                    />
                </div>

                <div>
                    <label className="block text-sm font-medium text-gray-700">
                        Опыт работы (%)
                    </label>
                    <input
                        type="number"
                        value={formData.weight_case}
                        onChange={(e) => setFormData({...formData, weight_case: parseInt(e.target.value)})}
                        min="0"
                        max="100"
                        className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                    />
                </div>
            </div>

            {error && (
                <div className="text-red-600 text-sm">
                    {error}
                </div>
            )}

            <button
                type="submit"
                disabled={loading}
                className="w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50"
            >
                {loading ? 'Создание...' : 'Создать вакансию'}
            </button>
        </form>
    );
}
