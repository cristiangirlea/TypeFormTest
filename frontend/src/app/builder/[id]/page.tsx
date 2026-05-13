'use client';

import { useEffect, useState, use } from 'react';
import { api, Form } from '@/lib/api';
import Link from 'next/link';

export default function Builder({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params);
  const [form, setForm] = useState<Form | null>(null);
  const [newQuestion, setNewQuestion] = useState('');
  const [error, setError] = useState('');

  useEffect(() => {
    api.getForms().then(forms => {
      const f = forms.find(f => f.id === id);
      if (f) setForm(f);
    });
  }, [id]);

  const handleAddQuestion = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newQuestion) return;
    try {
      const q = await api.addQuestion(id, newQuestion);
      setForm(prev => prev ? { ...prev, questions: [...prev.questions, q] } : null);
      setNewQuestion('');
      setError('');
    } catch (err) {
      console.error(err);
    }
  };

  const handleSave = async () => {
    try {
      const saved = await api.saveForm(id);
      setForm(saved);
      setError('');
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    }
  };

  if (!form) return <div className="p-8">Loading...</div>;

  return (
    <main className="p-8 max-w-4xl mx-auto text-black">
      <div className="flex justify-between items-center mb-8">
        <h1 className="text-3xl font-bold">{form.title} (Builder)</h1>
        <Link href="/" className="text-blue-500 hover:underline">Back to List</Link>
      </div>

      <section className="mb-8">
        <h2 className="text-xl font-semibold mb-4">Questions</h2>
        <div className="space-y-2 mb-4">
          {form.questions.map((q, i) => (
            <div key={q.id} className="border p-3 rounded bg-gray-50">
              {i + 1}. {q.text}
            </div>
          ))}
          {form.questions.length === 0 && <p className="text-gray-400">No questions yet. Add one below.</p>}
        </div>

        <form onSubmit={handleAddQuestion} className="flex gap-4">
          <input 
            type="text" 
            value={newQuestion} 
            onChange={e => setNewQuestion(e.target.value)}
            placeholder="Question Text"
            className="border p-2 rounded flex-1 text-black bg-white"
          />
          <button type="submit" className="bg-green-500 text-white px-4 py-2 rounded font-medium">Add Question</button>
        </form>
      </section>

      <section className="border-t pt-8">
        <button 
          onClick={handleSave} 
          className="bg-blue-600 text-white px-6 py-2 rounded font-bold hover:bg-blue-700 transition"
        >
          Save & Generate Link
        </button>
        {error && <p className="text-red-500 mt-2 font-medium">{error}</p>}
        
        {form.shareSlug && (
          <div className="mt-4 p-4 bg-green-100 rounded border border-green-200">
            <p className="font-semibold text-green-800">Form Saved!</p>
            <p className="text-green-700">Shareable URL: 
              <Link href={`/form/${form.shareSlug}`} className="underline ml-2 font-mono">
                /form/{form.shareSlug}
              </Link>
            </p>
          </div>
        )}
      </section>
    </main>
  );
}
