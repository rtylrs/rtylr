import React, { useState } from 'react';
import { HelpCircle } from 'lucide-react';

import AuthShell from '../components/AuthShell';
import AuthInput from '../components/AuthInput';
import SubmitButton from '../components/SubmitButton';
import BackLink from '../components/BackLink';
import notify from '../lib/notify/notify';

const QUESTIONS = [
  "What was the name of your first pet?",
  "In what city were you born?",
  "What was the make of your first car?",
  "What is your mother's maiden name?",
  "What was the name of your elementary school?",
  "What's the name of your favorite childhood friend?",
];

const SecurityQuestions: React.FC = () => {
  const [q1, setQ1] = useState(QUESTIONS[0]);
  const [a1, setA1] = useState('');
  const [q2, setQ2] = useState(QUESTIONS[1]);
  const [a2, setA2] = useState('');
  const [q3, setQ3] = useState(QUESTIONS[2]);
  const [a3, setA3] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = () => {
    setLoading(true);
    setTimeout(() => {
      setLoading(false);
      notify.success('Security questions saved');
    }, 500);
  };

  const renderQuestion = (
    idx: number,
    question: string,
    setQuestion: (v: string) => void,
    answer: string,
    setAnswer: (v: string) => void
  ) => (
    <div className="flex flex-col gap-2">
      <label className="text-[11px] font-semibold uppercase tracking-[0.08em] text-white/50">
        Question {idx}
      </label>
      <select
        value={question}
        onChange={(e) => setQuestion(e.target.value)}
        className="w-full h-[44px] rounded-xl bg-white/[0.04] border border-white/10 px-4 text-sm text-white outline-none focus:border-[#4583C4]"
      >
        {QUESTIONS.map((q) => (
          <option key={q} value={q} className="bg-[#0a0b2e]">
            {q}
          </option>
        ))}
      </select>
      <AuthInput
        id={`answer-${idx}`}
        type="text"
        label="Answer"
        icon={HelpCircle}
        placeholder="Your answer"
        value={answer}
        onChange={(e) => setAnswer(e.target.value)}
      />
    </div>
  );

  return (
    <AuthShell
      title="Security questions"
      subtitle="Set up backup recovery questions in case you lose access to your account."
      footer={
        <div className="flex justify-center">
          <BackLink label="Skip for now" />
        </div>
      }
    >
      <form
        onSubmit={(e) => {
          e.preventDefault();
          handleSubmit();
        }}
        className="flex flex-col gap-5"
      >
        {renderQuestion(1, q1, setQ1, a1, setA1)}
        {renderQuestion(2, q2, setQ2, a2, setA2)}
        {renderQuestion(3, q3, setQ3, a3, setA3)}
        <SubmitButton
          label="Save questions"
          loading={loading}
          disabled={!a1 || !a2 || !a3}
        />
      </form>
    </AuthShell>
  );
};

export default SecurityQuestions;
