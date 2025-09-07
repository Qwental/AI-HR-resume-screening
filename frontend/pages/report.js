import Link from 'next/link';
import { motion } from 'framer-motion';

export default function Report() {
  const candidate = {
    name: 'Иван Петров',
    score: 78,
    strengths: ['Опыт работы 5 лет', 'Хорошие коммуникационные навыки'],
    weaknesses: ['Нет опыта в React', 'Слабый английский']
  };

  return (
    <div className="p-6 bg-gray-50 min-h-screen">
      <motion.h1
        className="text-3xl font-bold mb-6"
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
      >
        Отчет по кандидату
      </motion.h1>
      <div className="bg-white shadow rounded-xl p-6 mb-6">
        <h2 className="text-xl mb-4">{candidate.name}</h2>
        <p className="text-lg mb-2">Оценка: <span className="font-bold text-indigo-600">{candidate.score}%</span></p>
        <h3 className="text-lg font-semibold mt-4 mb-2">Сильные стороны:</h3>
        <ul className="list-disc pl-6">
          {candidate.strengths.map((s, i) => <li key={i}>{s}</li>)}
        </ul>
        <h3 className="text-lg font-semibold mt-4 mb-2">Слабые стороны:</h3>
        <ul className="list-disc pl-6">
          {candidate.weaknesses.map((w, i) => <li key={i}>{w}</li>)}
        </ul>
      </div>
      <Link href="/" className="text-indigo-600 hover:underline">На главную</Link>
    </div>
  );
}