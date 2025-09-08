// Простая обёртка над Web Speech API + getUserMedia для записи и транскрипции-заглушки

export async function getMicStream() {
  if (!navigator.mediaDevices?.getUserMedia) throw new Error("getUserMedia не поддерживается");
  return navigator.mediaDevices.getUserMedia({ audio: true });
}

export function startSpeechRecognition(onResult) {
  const SR = window.SpeechRecognition || window.webkitSpeechRecognition;
  if (!SR) return { stop(){}, supported:false };
  const rec = new SR();
  rec.lang = "ru-RU";
  rec.interimResults = true;
  rec.continuous = true;
  rec.onresult = (e) => {
    let text = "";
    for (let i = e.resultIndex; i < e.results.length; i++) {
      text += e.results[i][0].transcript;
    }
    onResult(text);
  };
  rec.start();
  return { stop: () => rec.stop(), supported: true };
}

// Заглушка STT (эмуляция сервера): просто возвращает то же, что и вошло
export async function fakeTranscribe(blob) {
  return { text: "Транскрипция (mock): ответ принят" };
}
