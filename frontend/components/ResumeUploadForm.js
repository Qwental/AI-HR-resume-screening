import { useState } from 'react';
import { useDropzone } from 'react-dropzone';
import { toast } from 'react-hot-toast';
import { getToken } from '../utils/auth';

export default function ResumeUploadForm({ vacancyId, onSuccess }) {
    const [loading, setLoading] = useState(false);
    const [file, setFile] = useState(null);
    const [candidateName, setCandidateName] = useState('');
    const [candidateEmail, setCandidateEmail] = useState('');
    const [error, setError] = useState('');

    const onDrop = (acceptedFiles) => {
        const uploadFile = acceptedFiles[0];
        if (uploadFile) {
            // Поддерживаем больше форматов
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

            if (uploadFile.size > 20 * 1024 * 1024) { // 20MB
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
        if (!file || !candidateName || !candidateEmail) {
            setError('Заполните все поля и выберите файл');
            return;
        }

        setLoading(true);
        setError('');

        try {
            const formData = new FormData();
            formData.append('file', file);
            formData.append('vacancy_id', vacancyId);
            formData.append('candidate_name', candidateName);
            formData.append('candidate_email', candidateEmail);

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
                toast.success('Резюме успешно загружено!');
                onSuccess?.(result);

                // Сброс формы
                setFile(null);
                setCandidateName('');
                setCandidateEmail('');
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

            {/* Поля для данных кандидата */}
            <div className="space-y-4 mb-6">
                <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                        Имя кандидата *
                    </label>
                    <input
                        type="text"
                        value={candidateName}
                        onChange={(e) => setCandidateName(e.target.value)}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                        placeholder="Введите имя кандидата"
                    />
                </div>

                <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                        Email кандидата *
                    </label>
                    <input
                        type="email"
                        value={candidateEmail}
                        onChange={(e) => setCandidateEmail(e.target.value)}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                        placeholder="candidate@example.com"
                    />
                </div>
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
                disabled={loading || !file || !candidateName || !candidateEmail}
                className={`
          w-full mt-6 px-6 py-3 rounded-lg font-medium transition-colors
          ${loading || !file || !candidateName || !candidateEmail
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
