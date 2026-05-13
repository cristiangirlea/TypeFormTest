'use client';

import { useEffect, useState, use } from 'react';
import { api, Form } from '@/lib/api';
import Link from 'next/link';

export default function Renderer({ params }: { params: Promise<{ slug: string }> }) {
  const { slug } = use(params);
  const [form, setForm] = useState<Form | null>(null);
  const [currentIdx, setCurrentIdx] = useState(0);
  const [answers, setAnswers] = useState<{ [id: string]: string }>({});
  const [isFinished, setIsFinished] = useState(false);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api.getFormBySlug(slug).then(f => {
      setForm(f);
      setLoading(false);
    }).catch(err => {
      console.error(err);
      setLoading(false);
    });
  }, [slug]);

  const handleNext = () => {
    if (currentIdx < (form?.questions.length || 0) - 1) {
      setCurrentIdx(currentIdx + 1);
    } else {
      submit();
    }
  };

  const submit = async () => {
    const payload = Object.entries(answers).map(([id, val]) => ({
      questionID: id,
      value: val
    }));
    try {
      await api.submitResponse(slug, payload);
      setIsFinished(true);
    } catch (err) {
      console.error(err);
    }
  };

  if (loading) return <div className="p-8 text-black">Loading...</div>;
  if (!form) return <div className="p-8 text-black">Form not found</div>;

  if (isFinished) {
    return (
      <main className="h-screen flex items-center justify-center p-8 bg-zinc-50">
        <div className="text-center">
          <h1 className="text-4xl font-bold mb-4 text-black">Thank You!</h1>
          <p className="text-gray-600 mb-8 text-lg font-medium">Your responses have been saved.</p>
          <Link href="/" className="text-blue-500 hover:underline">Create your own form</Link>
        </div>
      </main>
    );
  }

  const question = form.questions[currentIdx];

  return (
    <main className="h-screen flex items-center justify-center p-8 bg-zinc-50 text-black">
      <div className="max-w-xl w-full">
        <p className="text-gray-400 mb-2 font-medium italic">Question {currentIdx + 1} of {form.questions.length}</p>
        <h2 className="text-3xl font-bold mb-8">{question.text}</h2>
        
        <input 
          autoFocus
          type="text"
          value={answers[question.id] || ''}
          onChange={e => setAnswers({ ...answers, [question.id]: e.target.value })}
          className="w-full border-b-2 border-blue-500 p-2 text-2xl outline-none bg-transparent"
          placeholder="Type your answer here..."
          onKeyDown={e => e.key === 'Enter' && handleNext()}
        />

        <div className="mt-8 flex justify-between items-center">
          <button 
            onClick={handleNext}
            className="bg-blue-600 text-white px-8 py-3 rounded text-lg font-bold shadow-lg hover:bg-blue-700 transition"
          >
            {currentIdx === form.questions.length - 1 ? 'Finish' : 'Next'}
          </button>
          <span className="text-gray-400 text-sm font-medium">Press Enter ↵</span>
        </div>
      </div>
    </main>
  );
}
