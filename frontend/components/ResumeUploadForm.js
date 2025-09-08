import { useState } from 'react';
import { useDropzone } from 'react-dropzone';
import { toast } from 'react-hot-toast';

export default function ResumeUploadForm({ vacancyId, onSuccess }) {
    const [loading, setLoading] = useState(false);
    const [file, setFile] = useState(null);
    const [error, setError] = useState('');

    const onDrop = (acceptedFiles) => {
        const uploadFile = acceptedFiles[0];
        if (uploadFile) {
            // Валидация типа файла
            if (uploadFile.type !== 'application/vnd.openxmlformats-officedocument.wordprocessingml.document') {
                setError('Поддерживаются только файлы .docx');
                return;
            }

            // Валидация размера (5MB)
            if (uploadFile.size > 5 * 1024 * 1024) {
                setError('Размер файла не должен превышать 5MB');
                return;
            }

            setFile(uploadFile);
            setError('');
        }
    };

    const { getRootProps, getInputProps, isDragActive } = useDropzone({
        onDrop,
        accept: {
            'application/vnd.openxmlformats-officedocument.wordprocessingml.document': ['.docx']
        },
        maxFiles: 1
    });

    const handleUpload = async () => {
        if (!file) {
            setError('Выберите файл для загрузки');
            return;
        }

        setLoading(true);

        try {
            const formData = new FormData();
            formData.append('file', file);
            formData.append('vacancy_id', vacancyId);

            const response = await fetch('/api/interview/resumes', {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('token')}`
                },
                body: formData
            });

            if (response.ok) {
                const result = await response.json();
                toast.success('Резюме успешно загружено!');
                onSuccess?.(result);
                setFile(null);
            } else {
                const error = await response.json();
                toast.error(error.message || 'Ошибка загрузки резюме');
            }
        } catch (error) {
            console.error('Error uploading resume:', error);
            toast.error('Ошибка соединения с сервером');
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6">
            <h3 className="text-lg font-medium text-gray-900 dark:text-white mb-4">
                Загрузка резюме
            </h3>

            <div
                {...getRootProps()}
                className={`border-2 border-dashed rounded-lg p-6 text-center cursor-pointer transition-colors ${
                    isDragActive
                        ? 'border-blue-400 bg-blue-50 dark:bg-blue-900/20'
                        : 'border-gray-300 dark:border-gray-600 hover:border-gray-400'
                }`}
            >
                <input {...getInputProps()} />

                {file ? (
                    <div className="text-green-600 dark:text-green-400">
                        <svg className="mx-auto h-8 w-8 mb-2" fill="currentColor" viewBox="0 0 20 20">
                            <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                        </svg>
                        <p className="font-medium">{file.name}</p>
                        <p className="text-sm text-gray-500 dark:text-gray-400">
                            {(file.size / 1024 / 1024).toFixed(2)} MB
                        </p>
                    </div>
                ) : (
                    <div className="text-gray-600 dark:text-gray-400">
                        <svg className="mx-auto h-8 w-8 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                        </svg>
                        <p>
                            {isDragActive
                                ? 'Отпустите файл резюме здесь'
                                : 'Перетащите файл резюме .docx или нажмите для выбора'
                            }
                        </p>
                        <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
                            Максимальный размер: 5MB
                        </p>
                    </div>
                )}
            </div>

            {error && (
                <p className="mt-2 text-sm text-red-600">{error}</p>
            )}

            {file && (
                <div className="mt-4 flex justify-end space-x-3">
                    <button
                        onClick={() => setFile(null)}
                        className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50"
                    >
                        Отменить
                    </button>
                    <button
                        onClick={handleUpload}
                        disabled={loading}
                        className="px-4 py-2 text-sm font-medium text-white bg-blue-600 border border-transparent rounded-md hover:bg-blue-700 disabled:opacity-50"
                    >
                        {loading ? 'Загрузка...' : 'Загрузить резюме'}
                    </button>
                </div>
            )}
        </div>
    );
}
