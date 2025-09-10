// components/VacancyEditor.js
import { useState, useEffect } from "react";

export default function VacancyEditor({ vacancy, onChange }) {
  const [weights, setWeights] = useState(vacancy.weights);
  const [warning, setWarning] = useState("");

  function handleChange(field, value) {
    const newWeights = { ...weights, [field]: Number(value) };
    const total = newWeights.tech + newWeights.comm + newWeights.cases;

    if (total > 100) {
      setWarning("⚠️ Сумма весов не должна превышать 100%");
    } else {
      setWarning("");
      onChange?.(newWeights);
    }

    setWeights(newWeights);
  }

  return (
    <div className="space-y-3">
      {["tech", "comm", "cases"].map((key) => (
        <div key={key} className="flex items-center justify-between">
          <label className="font-medium">
            {key === "tech" ? "Технические" : key === "comm" ? "Коммуникация" : "Кейсы"}
          </label>
          <input
            type="number"
            min={0}
            max={100}
            value={weights[key]}
            onChange={(e) => handleChange(key, e.target.value)}
            className="w-20 p-1 rounded border border-gray-300 dark:border-gray-600 text-right"
          />
          <span>%</span>
        </div>
      ))}
      {warning && <p className="text-red-500 text-sm">{warning}</p>}
      <p className="text-gray-400 text-sm">
        Сумма: {weights.tech + weights.comm + weights.cases}%
      </p>
    </div>
  );
}
