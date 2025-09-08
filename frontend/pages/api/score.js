export default function handler(req, res) {
  // Примитивная агрегация — просто возвращаем рандомный балл рядом с 75
  const score = Math.max(0, Math.min(100, Math.round(75 + (Math.random()*10 - 5))));
  res.status(200).json({ score });
}
