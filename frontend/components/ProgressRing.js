export default function ProgressRing({ value = 0, size = 120, stroke = 12 }) {
  const radius = (size - stroke) / 2;
  const circ = 2 * Math.PI * radius;
  const offset = circ - (value / 100) * circ;

  return (
    <svg width={size} height={size}>
      <circle cx={size/2} cy={size/2} r={radius} stroke="#eee" strokeWidth={stroke} fill="none"/>
      <circle cx={size/2} cy={size/2} r={radius} stroke="currentColor" strokeWidth={stroke}
        fill="none" strokeDasharray={circ} strokeDashoffset={offset} strokeLinecap="round"
        className="text-brand transition-[stroke-dashoffset] duration-500"/>
      <text x="50%" y="50%" dominantBaseline="middle" textAnchor="middle" className="fill-gray-700 dark:fill-gray-200 font-semibold">
        {value}%
      </text>
    </svg>
  );
}
