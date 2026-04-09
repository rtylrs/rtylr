import { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { Spinner } from '@voxire/ui';
import api from '../lib/api';

const EmailVerification = () => {
  const { token } = useParams();
  const [message, setMessage] = useState('');
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api.get(`/auth/verify/${token}`)
      .then((r) => setMessage(r.data.message))
      .catch(() => setMessage('Email verification failed. Please try again.'))
      .finally(() => setLoading(false));
  }, [token]);

  return (
    <div className="min-h-screen flex flex-col items-center justify-center bg-[#f8f8fb] text-center p-6">
      <div className="bg-white rounded-2xl shadow-card border border-black/5 p-8 max-w-sm w-full">
        <h1 className="text-2xl font-bold text-primary mb-4">Email Verification</h1>
        {loading ? <Spinner size="lg" className="text-accent mx-auto" /> : <p className="text-muted">{message}</p>}
      </div>
    </div>
  );
};

export default EmailVerification;
