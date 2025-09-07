import Layout from "../components/Layout";
import Waveform from "../components/Waveform";
import { useInterviewStore } from "../utils/store";
import { useEffect, useRef, useState } from "react";
import { getMicStream, startSpeechRecognition, fakeTranscribe } from "../utils/speech";

export default function Interview() {
  const {
    status,
    setStatus,
    questionIndex,
    nextQuestion,
    questions,
    transcript,
    setTranscript,
    resetTranscript
  } = useInterviewStore();

  const [stream, setStream] = useState(null);
  const recRef = useRef(null);
  const textareaRef = useRef(null);

  // Очистка ресурсов при размонтировании
  useEffect(() => {
    return () => {
      if (recRef.current) recRef.current.stop();
      if (stream) stream.getTracks().forEach(t => t.stop());
    };
  }, [stream]);

  // Автопрокрутка транскрипта вниз
  useEffect(() => {
    if (textareaRef.current) {
      textareaRef.current.scrollTop = textareaRef.current.scrollHeight;
    }
  }, [transcript]);

  async function startRecording() {
    setStatus("preparing");
    try {
      const s = await getMicStream();
      setStream(s);
      setStatus("recording");

      recRef.current = startSpeechRecognition((text) => {
        setTranscript(text);
      });
    } catch (e) {
      console.error(e);
      setStatus("error");
    }
  }

  async function stopRecording() {
    setStatus("processing");

    if (recRef.current) recRef.current.stop();
    if (stream) stream.getTracks().forEach(t => t.stop());
    setStream(null);

    // Симуляция обработки речи
    const res = await fakeTranscribe();
    setTranscript((t) => t + "\n" + res.text);
    setStatus("idle");
  }

  function handleNextQuestion() {
    nextQuestion();
    resetTranscript(); // транскрипт очищается только при переходе на следующий вопрос
  }

  return (
    <Layout title="AI Интервью">
      <div className="grid md:grid-cols-3 gap-6">
        {/* Вопрос и запись */}
        <div className="card p-6 md:col-span-2 flex flex-col items-center">
          <div className="text-sm text-gray-500 self-start">
            Вопрос {questionIndex + 1} из {questions.length}
          </div>
          <h2 className="text-2xl font-bold text-center mt-2">
            {questions[questionIndex]}
          </h2>

          <div className="mt-8 w-full relative">
            {stream ? (
              <>
                <Waveform stream={stream} />
                <div className="absolute top-2 right-2 w-4 h-4 rounded-full bg-red-500 animate-pulse"></div>
              </>
            ) : (
              <div className="h-[100px] rounded-xl bg-gray-100 dark:bg-gray-800 flex items-center justify-center text-gray-400 transition-colors duration-300">
                Микрофон не активен
              </div>
            )}
          </div>

          <div className="flex gap-3 mt-6">
            {status !== "recording" ? (
              <button onClick={startRecording} className="btn btn-primary flex items-center gap-2">
                🎤 Начать запись
              </button>
            ) : (
              <button onClick={stopRecording} className="btn btn-accent flex items-center gap-2">
                ⏹ Остановить
              </button>
            )}
            <button onClick={handleNextQuestion} className="btn">
              Пропустить
            </button>
          </div>
        </div>

        {/* Транскрипт */}
        <div className="card p-6 flex flex-col">
          <h3 className="font-semibold mb-2">Транскрипт (live)</h3>
          <textarea
            ref={textareaRef}
            value={transcript}
            onChange={(e) => setTranscript(e.target.value)}
            className="w-full h-80 p-3 rounded-xl bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 resize-none focus:outline-none focus:ring-2 focus:ring-brand transition-colors duration-300"
          />
          <div className="text-xs text-gray-400 mt-2">
            * Web Speech API используется, если поддерживается браузером
          </div>
        </div>
      </div>
    </Layout>
  );
}
