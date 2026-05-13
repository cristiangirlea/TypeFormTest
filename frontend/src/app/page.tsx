'use client';

import { useEffect, useState } from 'react';
import { api, Form } from '@/lib/api';
import Link from 'next/link';

export default function Home() {
  const [forms, setForms] = useState<Form[]>([]);
  const [title, setTitle] = useState('');

  useEffect(() => {
    api.getForms().then(setForms).catch(console.error);
  }, []);

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!title) return;
    try {
      const newForm = await api.createForm(title);
      window.location.href = `/builder/${newForm.id}`;
    } catch (err) {
      console.error(err);
    }
  };

  return (
    <main className="p-8 max-w-4xl mx-auto">
      <h1 className="text-3xl font-bold mb-8 text-black">My Forms</h1>
      
      <form onSubmit={handleCreate} className="mb-12 flex gap-4">
        <input 
          type="text" 
          value={title} 
          onChange={e => setTitle(e.target.value)}
          placeholder="New Form Title"
          className="border p-2 rounded flex-1 text-black"
        />
        <button type="submit" className="bg-blue-500 text-white px-4 py-2 rounded">Create Form</button>
      </form>

      <div className="grid gap-4">
        {forms.map(form => (
          <div key={form.id} className="border p-4 rounded flex justify-between items-center bg-white text-black">
            <div>
              <h2 className="text-xl font-semibold">{form.title}</h2>
              <p className="text-gray-500">{form.questions?.length || 0} questions</p>
            </div>
            <div className="flex gap-4">
              <Link href={`/builder/${form.id}`} className="text-blue-500 hover:underline">Edit Builder</Link>
              {form.shareSlug && (
                <Link href={`/form/${form.shareSlug}`} className="text-green-500 hover:underline">View Public Link</Link>
              )}
            </div>
          </div>
        ))}
      </div>
    </main>
  );
}
