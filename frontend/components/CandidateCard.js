import React from 'react';

function CandidateStatusIcon({ status }) {
  const bgColor = status === 'positive' ? 'bg-green-500' : 'bg-red-500';
  return <span className={`w-3.5 h-3.5 rounded-full ${bgColor}`}></span>;
}

export default function CandidateCard({ candidate }) {
  return (
    <li className="flex items-center justify-between p-4 my-2 rounded-lg bg-gray-800 hover:bg-gray-700 transition">
      <span className="text-gray-200">{candidate.name}</span>
      <div className="flex items-center gap-3">
        <CandidateStatusIcon status={candidate.status} />
        <span className="font-semibold text-blue-400 w-20 text-right">{candidate.score} баллов</span>
      </div>
    </li>
  );
}