import Layout from "../../components/Layout";
import Link from "next/link";
import { vacancies } from "../../utils/mockApi";

export default function VacanciesPage() {
  return (
    <Layout title="Вакансии">
      <div className="grid md:grid-cols-2 gap-6">
        {vacancies.map(v => (
          <div key={v.id} className="card p-6 flex flex-col justify-between">
            <div>
              <h3 className="text-lg font-semibold">{v.title}</h3>
              <div className="mt-2 text-sm text-gray-500">
                Вес критериев: Тех. {v.weights.tech}%, Комм. {v.weights.comm}%, Кейсы {v.weights.cases}%
              </div>
            </div>
            <Link href={`/vacancies/${v.id}`} className="btn btn-primary mt-4 self-start">
              Подробнее
            </Link>
          </div>
        ))}
      </div>
    </Layout>
  );
}
