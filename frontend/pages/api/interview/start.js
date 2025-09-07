export default function handler(req, res) {
  res.status(200).json({ ok: true, sessionId: "mock-session-123" });
}
