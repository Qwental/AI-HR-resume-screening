// components/VacancyForm.js
import { useState } from "react";
import VacancyEditor from "./VacancyEditor";

export default function VacancyForm({ onSave, onCancel }) {
  const [title, setTitle] = useState("");
  const [weights, setWeights] = useState({ tech: 50, comm: 30, cases: 20 });

  function handleSubmit(e) {
    e.preventDefault();
    const total = weights.tech + weights.comm + weights.cases;

    if (!title.trim()) return alert("Название вакансии обязательно!");
    if (total > 100) return alert("Сумма весов не должна превышать 100%");

    onSave({ id: Date.now(), title, weights, candidates: [] });
  }

  return (
    <form onSubmit={handleSubmit} className="card p-6 space-y-4">
      <h2 className="text-xl font-bold">Создать новую вакансию</h2>

      <div>
        <label className="block font-medium mb-1">Название вакансии:</label>
        <input
          type="text"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          className="w-full p-2 rounded border border-gray-300 dark:border-gray-600"
        />
      </div>

      <VacancyEditor vacancy={{ weights }} onChange={setWeights} />

      <div className="flex gap-3">
        <button type="submit" className="btn btn-primary">Сохранить</button>
        <button type="button" onClick={onCancel} className="btn btn-accent">Отмена</button>
      </div>
    </form>
  );
}
