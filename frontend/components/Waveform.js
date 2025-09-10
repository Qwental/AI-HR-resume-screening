import { useEffect, useRef } from "react";

export default function Waveform({ stream }) {
  const canvasRef = useRef(null);

  useEffect(() => {
    if (!stream || !canvasRef.current) return;
    const audioCtx = new (window.AudioContext || window.webkitAudioContext)();
    const source = audioCtx.createMediaStreamSource(stream);
    const analyser = audioCtx.createAnalyser();
    analyser.fftSize = 1024;
    source.connect(analyser);
    const bufferLength = analyser.frequencyBinCount;
    const dataArray = new Uint8Array(bufferLength);
    const canvas = canvasRef.current;
    const ctx = canvas.getContext("2d");

    function draw() {
      requestAnimationFrame(draw);
      analyser.getByteTimeDomainData(dataArray);
      ctx.clearRect(0, 0, canvas.width, canvas.height);
      ctx.beginPath();
      const slice = canvas.width / bufferLength;
      let x = 0;
      for (let i = 0; i < bufferLength; i++) {
        const v = dataArray[i] / 128.0;
        const y = (v * canvas.height) / 2;
        if (i === 0) ctx.moveTo(x, y);
        else ctx.lineTo(x, y);
        x += slice;
      }
      ctx.lineWidth = 2;
      ctx.strokeStyle = getComputedStyle(document.documentElement).getPropertyValue("--brand") || "#4F46E5";
      ctx.stroke();
    }
    draw();
    return () => audioCtx.close();
  }, [stream]);

  return <canvas ref={canvasRef} width={560} height={100} className="w-full rounded-xl bg-gray-100 dark:bg-gray-800" />;
}
