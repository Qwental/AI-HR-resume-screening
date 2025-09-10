import Layout from "../components/Layout";
import Link from "next/link";

export default function Home() {
  return (
    <Layout title="AI HR Platform">
      <div className="grid md:grid-cols-2 gap-6">
        <div className="card p-8">
          <h2 className="text-2xl font-bold">Начните с дашборда</h2>
          <p className="mt-2 text-gray-500">Общая картина найма, воронка, лучшие кандидаты.</p>
          <Link href="/dashboard" className="btn btn-primary mt-6">Открыть дашборд</Link>
        </div>
        <div className="card p-8">
          <h2 className="text-2xl font-bold">AI-интервью</h2>
          <p className="mt-2 text-gray-500">Структурированное интервью с динамическими вопросами.</p>
          <Link href="/interview" className="btn btn-accent mt-6">Запустить интервью</Link>
        </div>
      </div>
    </Layout>
  );
}
