// pages/resumes.js
import { useState, useEffect } from 'react';
import { useRouter } from 'next/router';
import Layout from '../components/Layout';
import { useAuthStore } from '../utils/store';
import { getToken } from '../utils/auth';

export default function ResumesPage() {
    const router = useRouter();
    const { isAuthenticated } = useAuthStore();
    const [resumes, setResumes] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');

    useEffect(() => {
        if (!isAuthenticated) {
            router.push('/login');
            return;
        }

        fetchResumes();
    }, [isAuthenticated, router]);

    const fetchResumes = async () => {
        const token = getToken();
        try {
            const response = await fetch('/api/all-resumes', {
                headers: {
                    'Authorization': `Bearer ${token}`,
                },
            });

            if (response.ok) {
                const data = await response.json();
                setResumes(data.resumes || []);
            } else {
                setError('Не удалось загрузить резюме');
            }
        } catch (error) {
            console.error('Error fetching resumes:', error);
            setError('Ошибка сети');
        } finally {
            setLoading(false);
        }
    };

    if (loading) {
        return (
            <Layout>
                <div className="text-center py-8">Загрузка резюме...</div>
            </Layout>
        );
    }

    return (
        <Layout>
            <div className="container mx-auto px-4 py-8">
                <h1 className="text-3xl font-bold mb-8">Все резюме</h1>

                {error && (
                    <div className="bg-red-50 text-red-600 p-4 rounded-md mb-6">
                        {error}
                    </div>
                )}

                {resumes.length === 0 ? (
                    <div className="text-center text-gray-500 py-8">
                        Резюме не найдены
                    </div>
                ) : (
                    <div className="grid gap-6">
                        {resumes.map((resume) => (
                            <div key={resume.id} className="bg-white p-6 rounded-lg shadow-md">
                                <h3 className="text-lg font-semibold">{resume.candidate_name}</h3>
                                <p className="text-gray-600">{resume.candidate_email}</p>
                                <p className="text-sm text-gray-500">
                                    Загружено: {new Date(resume.created_at).toLocaleDateString()}
                                </p>
                            </div>
                        ))}
                    </div>
                )}
            </div>
        </Layout>
    );
}
