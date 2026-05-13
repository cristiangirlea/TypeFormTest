const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export interface Question {
  id: string;
  text: string;
  type: string;
}

export interface Form {
  id: string;
  title: string;
  questions: Question[];
  shareSlug: string;
}

export const api = {
  getForms: async (): Promise<Form[]> => {
    const res = await fetch(`${API_BASE}/forms`);
    return res.json();
  },
  createForm: async (title: string): Promise<Form> => {
    const res = await fetch(`${API_BASE}/forms`, {
      method: 'POST',
      body: JSON.stringify({ title }),
      headers: { 'Content-Type': 'application/json' },
    });
    return res.json();
  },
  addQuestion: async (formId: string, text: string): Promise<Question> => {
    const res = await fetch(`${API_BASE}/forms/${formId}/questions`, {
      method: 'POST',
      body: JSON.stringify({ text, type: 'text' }),
      headers: { 'Content-Type': 'application/json' },
    });
    return res.json();
  },
  saveForm: async (formId: string): Promise<Form> => {
    const res = await fetch(`${API_BASE}/forms/${formId}/save`, {
      method: 'POST',
    });
    if (!res.ok) {
        const err = await res.text();
        throw new Error(err);
    }
    return res.json();
  },
  getFormBySlug: async (slug: string): Promise<Form> => {
    const res = await fetch(`${API_BASE}/form/${slug}`);
    return res.json();
  },
  submitResponse: async (slug: string, answers: { questionID: string, value: string }[]) => {
    const res = await fetch(`${API_BASE}/form/${slug}/responses`, {
      method: 'POST',
      body: JSON.stringify({ answers }),
      headers: { 'Content-Type': 'application/json' },
    });
    return res.json();
  }
};
