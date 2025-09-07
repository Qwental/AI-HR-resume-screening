import Layout from "../../components/Layout";
import { useRouter } from "next/router";
import { vacancies } from "../../utils/mockApi";

export default function VacancyDetail() {
  const router = useRouter();
  const { id } = router.query;

  const vacancy = vacancies.find(v => v.id === Number(id));
  if (!vacancy) return <Layout title="Вакансия">Вакансия не найдена</Layout>;

  return (
    <Layout title={vacancy.title}>
      <div className="card p-6 mb-6">
        <h2 className="text-xl font-bold mb-3">{vacancy.title}</h2>
        <div className="text-sm text-gray-500">
          <p>Технические: {vacancy.weights.tech}%</p>
          <p>Коммуникация: {vacancy.weights.comm}%</p>
          <p>Кейсы: {vacancy.weights.cases}%</p>
        </div>
      </div>

      <div className="card p-6">
        <h3 className="text-lg font-semibold mb-4">Кандидаты</h3>
        {vacancy.candidates.length === 0 ? (
          <p className="text-gray-500">Нет кандидатов</p>
        ) : (
          <table className="w-full text-sm">
            <thead>
              <tr className="text-left text-gray-500 border-b">
                <th className="py-2">Кандидат</th>
                <th className="py-2">% соответствия</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {vacancy.candidates.map(c => (
                <tr key={c.id} className="border-b border-gray-100 dark:border-gray-800">
                  <td className="py-3">{c.name}</td>
                  <td>
                    <span className="badge bg-indigo-100 text-indigo-700 dark:bg-indigo-900 dark:text-indigo-200">{c.score}%</span>
                  </td>
                  <td className="text-right">
                    <button className="btn btn-primary">Отчёт</button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </Layout>
  );
}
