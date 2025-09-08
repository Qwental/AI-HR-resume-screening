import { useState } from 'react';
import { useDropzone } from 'react-dropzone';
import { toast } from 'react-hot-toast';
import { getToken } from '../utils/auth';

export default function ResumeUploadForm({ vacancyId, onSuccess }) {
    const [loading, setLoading] = useState(false);
    const [file, setFile] = useState(null);
    const [error, setError] = useState('');

    const onDrop = (acceptedFiles) => {
        const uploadFile = acceptedFiles[0];
        if (uploadFile) {
            const allowedTypes = [
                'application/pdf',
                'application/msword',
                'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
                'text/plain'
            ];

            if (!allowedTypes.includes(uploadFile.type)) {
                setError('Поддерживаются только файлы PDF, DOC, DOCX, TXT');
                return;
            }

            if (uploadFile.size > 20 * 1024 * 1024) {
                setError('Размер файла не должен превышать 20MB');
                return;
            }

            setFile(uploadFile);
            setError('');
        }
    };

    const { getRootProps, getInputProps, isDragActive } = useDropzone({
        onDrop,
        accept: {
            'application/pdf': ['.pdf'],
            'application/msword': ['.doc'],
            'application/vnd.openxmlformats-officedocument.wordprocessingml.document': ['.docx'],
            'text/plain': ['.txt']
        },
        maxFiles: 1
    });

    const handleUpload = async () => {
        if (!file) {
            setError('Пожалуйста, выберите файл');
            return;
        }

        setLoading(true);
        setError('');

        try {
            const formData = new FormData();
            formData.append('file', file);
            formData.append('vacancy_id', vacancyId);
            // ✅ Убираем отправку candidate_name и candidate_email - их нет в модели

            const token = getToken();
            const response = await fetch('/api/resumes', {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${token}`
                },
                body: formData
            });

            if (response.ok) {
                const result = await response.json();
                toast.success('Резюме успешно загружено! Данные кандидата будут извлечены автоматически.');
                onSuccess?.(result);

                // Сброс формы
                setFile(null);
            } else {
                const error = await response.json();
                toast.error(error.error || 'Ошибка загрузки резюме');
                setError(error.error || 'Ошибка загрузки резюме');
            }
        } catch (error) {
            console.error('Error uploading resume:', error);
            toast.error('Ошибка соединения с сервером');
            setError('Ошибка соединения с сервером');
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="bg-white p-6 rounded-xl shadow-lg border">
            <h3 className="text-xl font-semibold mb-4">📤 Загрузить резюме</h3>

            {/* Информационное сообщение */}
            <div className="bg-blue-50 p-4 rounded-lg mb-6">
                <p className="text-sm text-blue-800">
                    💡 <strong>Автоматическое извлечение данных:</strong> Система автоматически извлечет
                    имя кандидата, email и другую информацию из загруженного резюме с помощью ИИ.
                </p>
            </div>

            {/* Drag & Drop зона */}
            <div
                {...getRootProps()}
                className={`
          border-2 border-dashed rounded-xl p-8 text-center cursor-pointer transition-colors
          ${isDragActive
                    ? 'border-blue-500 bg-blue-50'
                    : 'border-gray-300 hover:border-blue-400 hover:bg-gray-50'
                }
        `}
            >
                <input {...getInputProps()} />

                {file ? (
                    <div className="space-y-2">
                        <div className="text-green-600 text-2xl">✅</div>
                        <p className="font-medium text-gray-900">{file.name}</p>
                        <p className="text-sm text-gray-500">
                            {(file.size / 1024 / 1024).toFixed(2)} MB
                        </p>
                    </div>
                ) : (
                    <div className="space-y-3">
                        <div className="text-4xl text-gray-400">📄</div>
                        <div>
                            <p className="text-lg font-medium text-gray-900">
                                {isDragActive
                                    ? 'Отпустите файл резюме здесь'
                                    : 'Перетащите файл резюме или нажмите для выбора'
                                }
                            </p>
                            <p className="text-sm text-gray-500 mt-1">
                                Поддерживаются форматы: PDF, DOC, DOCX, TXT (макс. 20MB)
                            </p>
                        </div>
                    </div>
                )}
            </div>

            {/* Ошибки */}
            {error && (
                <div className="mt-4 p-3 bg-red-50 border border-red-200 rounded-lg">
                    <p className="text-red-700 text-sm">{error}</p>
                </div>
            )}

            {/* Кнопка загрузки */}
            <button
                onClick={handleUpload}
                disabled={loading || !file}
                className={`
          w-full mt-6 px-6 py-3 rounded-lg font-medium transition-colors
          ${loading || !file
                    ? 'bg-gray-300 text-gray-500 cursor-not-allowed'
                    : 'bg-blue-600 text-white hover:bg-blue-700'
                }
        `}
            >
                {loading ? (
                    <div className="flex items-center justify-center space-x-2">
                        <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
                        <span>Загружается...</span>
                    </div>
                ) : (
                    '📤 Загрузить резюме'
                )}
            </button>
        </div>
    );
}
