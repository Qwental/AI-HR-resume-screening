import React from 'react';
import Link from 'next/link';
import ProgressRing from './ProgressRing';

function CandidateStatusIcon({ status }) {
    const bgColor = status === 'positive' ? 'bg-green-500' :
        status === 'reviewed' ? 'bg-yellow-500' :
            'bg-red-500';
    return <div className={`w-3 h-3 rounded-full ${bgColor}`}></div>;
}

export default function CandidateCard({ candidate, showScore = true }) {
    return (
        <div className="bg-white rounded-lg shadow-md p-6 hover:shadow-lg transition-shadow">
            <div className="flex items-center justify-between">
                <div className="flex-1">
                    <div className="flex items-center gap-3 mb-2">
                        <h3 className="text-lg font-semibold text-gray-900">
                            {candidate.name || candidate.candidate_name}
                        </h3>
                        <CandidateStatusIcon status={candidate.status || 'new'} />
                    </div>

                    <div className="text-sm text-gray-600 space-y-1">
                        {candidate.candidate_email && (
                            <p>📧 {candidate.candidate_email}</p>
                        )}
                        <p>📅 {new Date(candidate.created_at || Date.now()).toLocaleDateString()}</p>
                    </div>

                    {candidate.skills && (
                        <div className="mt-3">
                            <div className="flex flex-wrap gap-2">
                                {candidate.skills.slice(0, 3).map((skill, index) => (
                                    <span
                                        key={index}
                                        className="px-2 py-1 bg-blue-100 text-blue-800 text-xs rounded-full"
                                    >
                    {skill}
                  </span>
                                ))}
                                {candidate.skills.length > 3 && (
                                    <span className="px-2 py-1 bg-gray-100 text-gray-600 text-xs rounded-full">
                    +{candidate.skills.length - 3}
                  </span>
                                )}
                            </div>
                        </div>
                    )}
                </div>

                {showScore && candidate.score && (
                    <div className="ml-4 flex flex-col items-center">
                        <ProgressRing value={candidate.score} size={80} stroke={8} />
                        <p className="text-xs text-gray-500 mt-2">соответствие</p>
                    </div>
                )}
            </div>

            <div className="flex justify-between items-center mt-4 pt-4 border-t">
        <span className={`px-3 py-1 rounded-full text-xs font-medium ${
            candidate.status === 'new' ? 'bg-blue-100 text-blue-800' :
                candidate.status === 'reviewed' ? 'bg-yellow-100 text-yellow-800' :
                    candidate.status === 'accepted' ? 'bg-green-100 text-green-800' :
                        'bg-gray-100 text-gray-800'
        }`}>
          {candidate.status === 'new' ? 'Новое' :
              candidate.status === 'reviewed' ? 'Просмотрено' :
                  candidate.status === 'accepted' ? 'Принято' : 'Неактивно'}
        </span>

                {candidate.id && (
                    <Link
                        href={`/report/${candidate.id}`}
                        className="text-blue-600 hover:text-blue-800 text-sm font-medium"
                    >
                        Отчёт →
                    </Link>
                )}
            </div>
        </div>
    );
}
