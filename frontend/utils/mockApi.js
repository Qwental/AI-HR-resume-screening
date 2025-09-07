export const vacancies = [
  {
    id: 1,
    title: "Frontend Developer",
    weights: { tech: 50, comm: 30, cases: 20 },
    candidates: [
      { id: 101, name: "ivan.ivanov@gmail.com", score: 85, status: "negative" },
      { id: 102, name: "masha@gmail.com", score: 78, status: "positive" }
    ]
  },
  {
    id: 2,
    title: "Backend Developer",
    weights: { tech: 60, comm: 20, cases: 20 },
    candidates: [
      { id: 201, name: "lexa@gmail.com", score: 90, status: "negative" },
      { id: 202, name: "olga@gmail.com", score: 75, status: "positive" }
    ]
  }
];

export const candidates = vacancies.flatMap(v =>
    v.candidates.map(c => ({ ...c, vacancyId: v.id }))
);

export const getCandidate = (id) =>
    candidates.find(c => c.id === Number(id)) || null;
