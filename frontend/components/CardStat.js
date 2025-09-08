export default function CardStat({ label, value, hint }) {
  return (
    <div className="card p-6">
      <div className="text-gray-500 text-sm">{label}</div>
      <div className="text-3xl font-bold mt-1">{value}</div>
      {hint && <div className="text-xs text-gray-400 mt-1">{hint}</div>}
    </div>
  );
}
