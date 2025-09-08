import { useRouter } from "next/router";
import Layout from "../../components/Layout";
import ProgressRing from "../../components/ProgressRing";
import { getCandidate } from "../../utils/mockApi";

export default function Report() {
  const { query } = useRouter();
  const cand = getCandidate(query.id);
  if (!cand) return <Layout title="Отчёт"><div className="card p-6">Кандидат не найден</div></Layout>;

  return (
    <Layout title={`Отчёт: ${cand.name}`}>
      <div className="grid md:grid-cols-3 gap-6">
        <div className="card p-6 md:col-span-2">
          <h2 className="text-xl font-semibold">Итоговая оценка</h2>
          <div className="mt-4 flex items-center gap-8">
            <ProgressRing value={cand.score} />
            <div className="space-y-2 text-sm">
              <div className="flex items-center gap-3"><span className="w-36 text-gray-500">Hard Skills</span><div className="progress"><span style={{width: cand.hard + '%'}}/></div><span className="w-10 text-right">{cand.hard}%</span></div>
              <div className="flex items-center gap-3"><span className="w-36 text-gray-500">Soft Skills</span><div className="progress"><span style={{width: cand.soft + '%'}}/></div><span className="w-10 text-right">{cand.soft}%</span></div>
            </div>
          </div>
          <div className="mt-6">
            <h3 className="font-semibold mb-2">Обоснование (AI)</h3>
            <ul className="list-disc pl-6 text-sm text-gray-600 dark:text-gray-300">
              <li>Подтверждены ключевые навыки: Python, ML, SQL.</li>
              <li>Есть пробелы в командной коммуникации и стори-теллинге.</li>
              <li>Сомнения по стажу: упоминает 5 лет при 3 в резюме.</li>
            </ul>
          </div>
        </div>
        <div className="card p-6">
          <h3 className="font-semibold mb-3">Риски</h3>
          <ul className="list-disc pl-6 text-sm">
            {cand.flags.length ? cand.flags.map((f,i)=>(<li key={i}>{f}</li>)) : <li>не выявлены</li>}
          </ul>
          <button className="btn btn-primary mt-6 w-full">Одобрить</button>
          <button className="btn mt-2 w-full">Запросить доинтервью</button>
        </div>
      </div>
    </Layout>
  );
}
