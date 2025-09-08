import Layout from "../components/Layout";
import CardStat from "../components/CardStat";
import SimpleChart from "../components/SimpleChart";
import { candidates } from "../utils/mockApi";

const chartData = [
  { name: "Пн", score: 64 },
  { name: "Вт", score: 72 },
  { name: "Ср", score: 70 },
  { name: "Чт", score: 76 },
  { name: "Пт", score: 74 },
  { name: "Сб", score: 79 },
  { name: "Вс", score: 81 },
];

export default function Dashboard() {
  return (
    <Layout title="Дашборд">
      <div className="grid md:grid-cols-4 gap-6 mb-6">
        <CardStat label="Кандидатов в процессе" value={45} />
        <CardStat label="Средний % соответствия" value="72%" />
        <CardStat label="Экономия времени (неделя)" value="14 ч" />
        <CardStat label="Прошло интервью за 24ч" value="23" />
      </div>

      <div className="grid md:grid-cols-3 gap-6">
        <div className="card p-6 md:col-span-2">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-semibold">Динамика соответствия</h3>
          </div>
          <SimpleChart data={chartData} />
        </div>

        <div className="card p-6">
          <h3 className="text-lg font-semibold mb-4">ТОП кандидаты</h3>
          <ul className="space-y-3">
            {candidates.slice(0,3).map(c => (
              <li key={c.id} className="flex items-center justify-between">
                <span className="truncate">{c.name}</span>
                <span className="badge bg-indigo-100 text-indigo-700 dark:bg-indigo-900 dark:text-indigo-200">{c.score}%</span>
              </li>
            ))}
          </ul>
        </div>
      </div>
    </Layout>
  );
}
